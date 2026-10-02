package certificates

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/opencontainers/selinux/go-selinux"
	"golang.org/x/sys/unix"
)

var ErrBusy = errors.New("MicroShift startup or another certificate operation is using the PKI")

// Lock uses a stable file shared by run, status, and renewal.
// Readers hold shared locks; startup and pending renewal require an exclusive
// operation lock. A separate runtime lock prevents recovery while running.
// The file must not be unlinked on release.
func Lock(path string, exclusive bool) (*os.File, error) {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("cannot create certificate lock directory: %w", err)
	}
	if err := labelLockPath(directory); err != nil {
		return nil, fmt.Errorf("cannot label certificate lock directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("certificate lock must be a regular file")
	}
	if err := labelLockPath(path); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("cannot label certificate lock: %w", err)
	}
	mode := unix.LOCK_SH
	if exclusive {
		mode = unix.LOCK_EX
	}
	if err := flock(file, mode); err != nil {
		_ = file.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("cannot lock certificate operation: %w", err)
	}
	return file, nil
}

func labelLockPath(path string) error {
	if !selinux.GetEnabled() {
		return nil
	}
	label, err := selinux.FileLabel(path)
	if err != nil {
		return err
	}
	// The existing backup-directory policy handles creation by the confined
	// service. An unconfined CLI must apply the directory and file contexts too.
	if strings.Contains(label, ":container_var_lib_t:") {
		return nil
	}
	return exec.Command("restorecon", "--", path).Run()
}

func flock(file *os.File, mode int) error {
	//nolint:gosec // An open OS file descriptor fits the syscall's signed int argument.
	return unix.Flock(int(file.Fd()), mode|unix.LOCK_NB)
}
