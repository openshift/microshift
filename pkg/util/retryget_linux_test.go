//go:build linux

package util

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The black-hole regression test runs RetryGet in a child process so that a
// stalled, uncancellable attempt can be killed and reaped on watchdog expiry
// instead of being left behind in this test binary.
const (
	blackHoleSubjectEnv    = "MICROSHIFT_TEST_RETRYGET_BLACKHOLE_ADDR"
	blackHoleSubjectBudget = 8 * time.Second
)

// blackHoleListener returns a loopback address whose SYNs the kernel drops:
// a listening socket with a tiny backlog whose accept queue has been filled
// and is never drained. Connects to it hang in SYN-SENT exactly like a
// connect to an unreachable address. The filler connections and the socket
// are closed by t.Cleanup.
//
// Skip policy: the test is skipped when the kernel's SYN retry schedule is
// unknown or too short to tell an unbounded connect from a bounded one, or
// when the kernel does not drop SYNs on accept-queue overflow (every filler
// connects). Any other error while building the fixture is a test failure.
func blackHoleListener(t *testing.T) string {
	t.Helper()
	retries, err := readSysctlInt("/proc/sys/net/ipv4/tcp_syn_retries")
	if err != nil {
		t.Skipf("cannot read tcp_syn_retries: %v", err)
	}
	if retries < 4 {
		t.Skipf("tcp_syn_retries=%d gives up in under 31 s; too short to discriminate", retries)
	}
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		t.Fatalf("socket: %v", err)
	}
	t.Cleanup(func() {
		if err := syscall.Close(fd); err != nil {
			t.Errorf("close listener: %v", err)
		}
	})
	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if err := syscall.Listen(fd, 1); err != nil {
		t.Fatalf("listen: %v", err)
	}
	sa, err := syscall.Getsockname(fd)
	if err != nil {
		t.Fatalf("getsockname: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", sa.(*syscall.SockaddrInet4).Port)

	// Fill the accept queue. The kernel admits backlog+1 established
	// connections; the first connect after that must time out.
	const fillerTimeout = 2 * time.Second
	for i := 0; i < 8; i++ {
		c, err := net.DialTimeout("tcp", addr, fillerTimeout)
		if err == nil {
			t.Cleanup(func() {
				if err := c.Close(); err != nil {
					t.Errorf("close filler connection: %v", err)
				}
			})
			continue
		}
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return addr
		}
		t.Fatalf("unexpected dial error while filling the accept queue: %v", err)
	}
	t.Skip("the kernel accepted every filler connection; cannot build a loopback black hole")
	return ""
}

func readSysctlInt(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

// TestRetryGetBlackHoleSubject is the child side of
// TestRetryGet_UnansweredConnectIsBounded. It does nothing unless invoked
// with the black-hole address in the environment.
func TestRetryGetBlackHoleSubject(t *testing.T) {
	addr := os.Getenv(blackHoleSubjectEnv)
	if addr == "" {
		t.Skip("child process helper for TestRetryGet_UnansweredConnectIsBounded")
	}
	ctx, cancel := context.WithTimeout(context.Background(), blackHoleSubjectBudget)
	defer cancel()
	start := time.Now()
	status := RetryGet(ctx, fmt.Sprintf("http://%s/healthz", addr), "")
	if _, err := fmt.Printf("subject status=%d elapsed_ms=%d\n", status, time.Since(start).Milliseconds()); err != nil {
		t.Fatalf("write result: %v", err)
	}
}

func TestRetryGet_UnansweredConnectIsBounded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in -short mode: takes about 10 s")
	}
	addr := blackHoleListener(t)

	// The child's budget is shorter than the poll's own 120 s so the test
	// stays fast. Without a dial timeout a single attempt blocks for the
	// kernel's SYN retry schedule (about two minutes with tcp_syn_retries=6);
	// with it, attempts are bounded and the poll returns shortly after the
	// child's deadline. An attempt started just before that deadline may run
	// for one more dial timeout, so allow that plus scheduling slack.
	maxElapsed := blackHoleSubjectBudget + retryGetDialTimeout + 2*time.Second
	watchdog := maxElapsed + 5*time.Second

	// Independent watchdog: when it expires the child is killed and reaped
	// by exec, so a stalled attempt cannot outlive this test.
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("locate the test binary: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchdog)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestRetryGetBlackHoleSubject$", "-test.v")
	cmd.Env = append(os.Environ(), blackHoleSubjectEnv+"="+addr)
	out, err := cmd.Output()
	if ctx.Err() != nil {
		t.Fatalf("RetryGet did not return within %s: a stalled connect is not bounded\n%s", watchdog, out)
	}
	if err != nil {
		t.Fatalf("child process failed: %v\n%s", err, out)
	}

	var status int
	var elapsedMs int64
	for _, line := range strings.Split(string(out), "\n") {
		if _, err := fmt.Sscanf(line, "subject status=%d elapsed_ms=%d", &status, &elapsedMs); err == nil {
			break
		}
	}
	if elapsedMs == 0 {
		t.Fatalf("no result line from the child process:\n%s", out)
	}
	elapsed := time.Duration(elapsedMs) * time.Millisecond
	if status != 0 {
		t.Fatalf("status = %d, want 0 (nothing answers)", status)
	}
	if elapsed > maxElapsed {
		t.Fatalf("took %s, want at most %s: a stalled connect is not bounded", elapsed, maxElapsed)
	}
	if elapsed < blackHoleSubjectBudget-500*time.Millisecond {
		t.Fatalf("took %s, want about the child's budget %s", elapsed, blackHoleSubjectBudget)
	}
}
