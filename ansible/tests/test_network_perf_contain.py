"""Tests of run-network-perf's contain.py with real processes; no cluster or playbook.

Run: python3 -m unittest discover -s ansible/tests -v
"""

import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time
import unittest

CONTAIN = Path(__file__).resolve().parents[1] / "roles/run-network-perf/files/contain.py"


def running(pid):
    """Whether a process exists and is not a zombie."""
    try:
        with open(f"/proc/{pid}/stat", encoding="ascii", errors="replace") as stat:
            return stat.read().rsplit(") ", 1)[1].split()[0] != "Z"
    except OSError:
        return False


def start_time(pid):
    with open(f"/proc/{pid}/stat", encoding="ascii", errors="replace") as stat:
        return stat.read().rsplit(") ", 1)[1].split()[19]


def wait_for(condition, seconds=10):
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        if condition():
            return True
        time.sleep(0.02)
    return False


class Contain(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.journal = self.root / "journal"
        self.journal.mkdir()
        self.pids = []
        self.addCleanup(self.kill_leftovers)

    def kill_leftovers(self):
        for path in self.root.glob("*.pid"):
            try:
                os.kill(int(path.read_text()), signal.SIGKILL)
            except (OSError, ValueError):
                pass

    def contain(self, *command, grace="1", **kwargs):
        return subprocess.run([sys.executable, str(CONTAIN), "--grace", grace, str(self.journal), "ctx", *command],
                              capture_output=True, text=True, timeout=60, check=False, **kwargs)

    def start(self, *command, grace="1"):
        return subprocess.Popen([sys.executable, str(CONTAIN), "--grace", grace, str(self.journal), "ctx", *command],
                                stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    def entries(self):
        return sorted(p.name for p in (self.journal / "commands").glob("*/*"))

    def pid(self, name):
        path = self.root / f"{name}.pid"
        self.assertTrue(wait_for(lambda: path.exists() and path.read_text().strip()), f"{name} did not start")
        return int(path.read_text())

    def record(self, **changes):
        with open("/proc/sys/kernel/hostname", encoding="ascii") as name, \
                open("/proc/sys/kernel/random/boot_id", encoding="ascii") as boot:
            data = {"hostname": name.read().strip(), "boot_id": boot.read().strip(), "uid": str(os.getuid()),
                    "pid": "1", "start": "0"}
        data.update(changes)
        (self.journal / "invocation.json").write_text(json.dumps(data))

    def stop(self):
        result = subprocess.run([sys.executable, str(CONTAIN), "--stop", str(self.journal)],
                                capture_output=True, text=True, timeout=60, check=False)
        return result.returncode, json.loads(result.stdout)

    def test_status_and_streams_pass_through(self):
        self.assertEqual(self.contain("true").returncode, 0)
        self.assertEqual(self.contain("sh", "-c", "exit 3").returncode, 3)
        self.assertEqual(self.contain("sh", "-c", "kill -KILL $$").returncode, 128 + signal.SIGKILL)
        result = self.contain("sh", "-c", "cat; echo err >&2", input="in\n")
        self.assertEqual((result.stdout, result.stderr), ("in\n", "err\n"))
        self.assertEqual(self.entries(), [])

    def test_missing_command_is_reported_and_forgotten(self):
        self.assertEqual(self.contain("/nonexistent/command").returncode, 127)
        self.assertEqual(self.entries(), [])

    def test_entry_exists_while_the_command_runs(self):
        proc = self.start("sh", "-c", f"echo $$ > {self.root}/main.pid; exec sleep 300")
        self.pid("main")
        self.assertEqual(self.entries(), [f"{proc.pid}-{start_time(proc.pid)}"])
        proc.send_signal(signal.SIGTERM)
        self.assertEqual(proc.wait(timeout=30), 128 + signal.SIGTERM)
        proc.stdout.close()
        proc.stderr.close()
        self.assertEqual(self.entries(), [])

    def test_detached_child_is_ended_before_returning(self):
        # A child in its own session whose parent has exited: only a subreaper still sees it.
        result = self.contain("sh", "-c", f"setsid sh -c 'echo $$ > {self.root}/orphan.pid; exec sleep 300' & sleep 0.3; exit 0")
        self.assertEqual(result.returncode, 0)
        self.assertFalse(running(self.pid("orphan")))
        self.assertEqual(self.entries(), [])

    def test_adopted_child_ignoring_term_is_killed(self):
        result = self.contain("sh", "-c", f"setsid sh -c 'trap \"\" TERM; echo $$ > {self.root}/deaf.pid; exec sleep 300' & sleep 0.3; exit 0", grace="1")
        self.assertEqual(result.returncode, 0)
        self.assertFalse(running(self.pid("deaf")))
        self.assertEqual(self.entries(), [])

    def test_containment_finishes_after_the_caller_has_gone(self):
        # The caller dies and its pipes close; the supervisor must still end the tree.
        code = (f"import subprocess, sys; subprocess.Popen([sys.executable, {str(CONTAIN)!r}, '--grace', '1', {str(self.journal)!r}, 'ctx', "
                f"'sh', '-c', 'echo $$ > {self.root}/late.pid; while :; do echo x; echo y >&2; sleep 0.1; done'], "
                "stdout=subprocess.PIPE, stderr=subprocess.PIPE); import time; time.sleep(0.5)")
        subprocess.run([sys.executable, "-c", code], check=True, timeout=30)
        late = self.pid("late")
        self.assertTrue(wait_for(lambda: not running(late) and self.entries() == []))

    def test_child_ignoring_term_is_killed_after_the_grace_period(self):
        begin = time.monotonic()
        result = self.contain("sh", "-c", f"(trap '' TERM; echo $$ > {self.root}/stubborn.pid; exec sleep 300) & sleep 0.3; exit 5", grace="1")
        self.assertEqual(result.returncode, 5)
        self.assertGreaterEqual(time.monotonic() - begin, 1)
        self.assertFalse(running(self.pid("stubborn")))
        self.assertEqual(self.entries(), [])

    def test_interrupt_ends_the_whole_tree(self):
        proc = self.start("sh", "-c", f"(echo $$ > {self.root}/deep.pid; exec sleep 300) & echo $$ > {self.root}/top.pid; wait")
        deep, top = self.pid("deep"), self.pid("top")
        proc.send_signal(signal.SIGINT)
        self.assertEqual(proc.wait(timeout=30), 128 + signal.SIGINT)
        proc.stdout.close()
        proc.stderr.close()
        self.assertFalse(running(deep) or running(top))
        self.assertEqual(self.entries(), [])

    def test_stopped_journal_starts_nothing(self):
        (self.journal / "stopped").touch()
        result = self.contain("sh", "-c", f"touch {self.root}/started")
        self.assertEqual(result.returncode, 97)
        self.assertFalse((self.root / "started").exists())
        self.assertEqual(self.entries(), [])

    def test_unwritable_journal_starts_nothing(self):
        (self.journal / "commands").write_text("not a directory")
        result = self.contain("sh", "-c", f"touch {self.root}/started")
        self.assertEqual(result.returncode, 97)
        self.assertFalse((self.root / "started").exists())

    def test_killed_supervisor_leaves_its_entry(self):
        proc = self.start("sh", "-c", f"echo $$ > {self.root}/left.pid; exec sleep 300")
        self.pid("left")
        entry = self.entries()
        proc.kill()
        proc.wait(timeout=30)
        proc.stdout.close()
        proc.stderr.close()
        self.assertEqual(self.entries(), entry)
        self.record()
        code, verdict = self.stop()
        self.assertEqual(code, 1)
        self.assertIn("ended without confirming", verdict["reasons"][0])

    def test_stop_reports_a_running_command_and_blocks_new_ones(self):
        self.record()
        proc = self.start("sh", "-c", f"echo $$ > {self.root}/busy.pid; exec sleep 300")
        self.pid("busy")
        code, verdict = self.stop()
        self.assertEqual(code, 1)
        self.assertIn(f"kill {proc.pid}", verdict["reasons"][0])
        self.assertEqual(self.contain("true").returncode, 97)
        proc.send_signal(signal.SIGTERM)
        proc.wait(timeout=30)
        proc.stdout.close()
        proc.stderr.close()
        self.assertEqual(self.stop(), (0, {"ended": True, "reasons": []}))

    def test_stop_refuses_while_the_playbook_runs(self):
        self.record(pid=str(os.getpid()), start=start_time(os.getpid()))
        code, verdict = self.stop()
        self.assertEqual(code, 1)
        self.assertIn(f"PID {os.getpid()}", verdict["reasons"][0])

    def test_stop_treats_a_reused_pid_as_ended(self):
        self.record(pid=str(os.getpid()), start="1")
        self.assertEqual(self.stop()[0], 0)

    def test_stop_changes_nothing_for_another_host_or_user(self):
        for changes in ({"hostname": "elsewhere"}, {"uid": "0" if os.getuid() else "1"}):
            self.record(**changes)
            code, verdict = self.stop()
            self.assertEqual(code, 1)
            self.assertIn("was started by", verdict["reasons"][0])
            self.assertFalse((self.journal / "stopped").exists())

    def test_stop_after_a_reboot_ignores_recorded_processes(self):
        self.record(boot_id="another-boot", pid=str(os.getpid()), start=start_time(os.getpid()))
        (self.journal / "commands" / "ctx").mkdir(parents=True)
        (self.journal / "commands" / "ctx" / f"{os.getpid()}-{start_time(os.getpid())}").touch()
        self.assertEqual(self.stop(), (0, {"ended": True, "reasons": []}))

    @unittest.skipIf(os.geteuid() == 0, "root ignores file modes")
    def test_stop_fails_closed_on_an_unreadable_journal(self):
        self.record()
        hidden = self.journal / "commands" / "ctx"
        hidden.mkdir(parents=True)
        (hidden / "1-1").touch()
        hidden.chmod(0)
        self.addCleanup(hidden.chmod, 0o755)
        code, verdict = self.stop()
        self.assertEqual(code, 1)
        self.assertIn("cannot tell", verdict["reasons"][0])

    @unittest.skipIf(os.geteuid() == 0, "root ignores file modes")
    def test_stop_fails_closed_on_an_unreadable_record(self):
        self.record()
        (self.journal / "invocation.json").chmod(0)
        code, verdict = self.stop()
        self.assertEqual(code, 1)
        self.assertIn("cannot be read", verdict["reasons"][0])

    def test_stop_fails_closed_on_a_damaged_record(self):
        for content in (None, "{", json.dumps({"hostname": "x"})):
            if content is not None:
                (self.journal / "invocation.json").write_text(content)
            code, verdict = self.stop()
            self.assertEqual(code, 1)
            self.assertIn("cannot be read", verdict["reasons"][0])


if __name__ == "__main__":
    unittest.main()
