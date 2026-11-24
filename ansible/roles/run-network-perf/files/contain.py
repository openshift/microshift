#!/usr/bin/env python3
"""Run a command and return only after every process it started has ended.

Usage:
  contain.py [--grace SECONDS] <journal dir> <context> <command> [args...]
  contain.py --stop <journal dir>

The command is registered as <journal dir>/commands/<context>/<pid>-<start time>
before it starts, and the entry is removed only once this process has no children
left. As a child subreaper it inherits every process the command leaves behind,
however that process detached. An entry that is still there means the command, or
something it started, may still be running: that, not the exit status, is what a
caller must check before it relies on the command having ended. Children left after
the command ended, and all children on INT, TERM or HUP, are ended: TERM, then KILL
after the grace period. Only this process's own children are ever signalled.

Exit status: the command's, or 128 + signal; 97 when nothing was started.

--stop makes every later contain.py for that journal refuse to start, then prints
as JSON whether the run that wrote the journal has ended (exit 0) or not (exit 1),
with the reasons. It reads <journal dir>/invocation.json and never signals anything.
"""

import argparse
import ctypes
import json
import os
import signal
import sys
import time

UNCONTAINED = 97
PR_SET_CHILD_SUBREAPER = 36
# Seconds after KILL before giving up on children that do not end
GIVE_UP = 30
SIGNALS = (signal.SIGINT, signal.SIGTERM, signal.SIGHUP)


class Stop(Exception):
    """A signal asked this process to end the command."""


def on_signal(signum, _frame):
    raise Stop(signum)


def say(message):
    """Report on stderr; whoever started this process may be gone already."""
    try:
        print(f"contain.py: {message}", file=sys.stderr, flush=True)
    except OSError:
        pass


def stat_fields(pid):
    """Fields of /proc/<pid>/stat after the command name, which may contain spaces."""
    with open(f"/proc/{pid}/stat", encoding="ascii", errors="replace") as stat:
        return stat.read().rsplit(") ", 1)[1].split()


def start_time(pid):
    """Start time in clock ticks; with the PID it identifies a process until reboot."""
    return stat_fields(pid)[19]


def alive(pid, start):
    """Whether that process still runs. Only a missing PID or another start time say no."""
    try:
        return start_time(pid) == str(start)
    except (FileNotFoundError, ProcessLookupError):
        return False


def children():
    """PIDs of this process's children, adopted orphans included.

    Only this process reaps them and it does so between calls, never during one, so a
    PID found here still names that child when it is signalled.
    """
    me = str(os.getpid())
    found = []
    for name in os.listdir("/proc"):
        if name.isdigit():
            try:
                if stat_fields(name)[1] == me:
                    found.append(int(name))
            except (OSError, IndexError):
                continue
    return found


def reap():
    """(pid, status) of a child that ended, (0, 0) if none has, None without children."""
    try:
        return os.waitpid(-1, os.WNOHANG)
    except ChildProcessError:
        return None


def run(journal, context, command, grace):
    libc = ctypes.CDLL(None, use_errno=True)
    if libc.prctl(PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0) != 0:
        say("cannot become a child subreaper")
        return UNCONTAINED
    entry = os.path.join(journal, "commands", context, f"{os.getpid()}-{start_time(os.getpid())}")
    try:
        os.makedirs(os.path.dirname(entry), exist_ok=True)
        os.close(os.open(entry, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o644))
    except OSError as err:
        say(f"cannot register the command: {err}")
        return UNCONTAINED
    # Registered first, checked second: --stop writes "stopped" first and reads the
    # entries second, so one of the two always sees the other.
    if os.path.exists(os.path.join(journal, "stopped")):
        os.unlink(entry)
        say(f"{journal} is stopped, not starting {command[0]}")
        return UNCONTAINED

    main, status, interrupted = None, None, 0
    try:
        for signum in SIGNALS:
            signal.signal(signum, on_signal)
        try:
            main = os.posix_spawnp(command[0], command, os.environ, setsigdef=[signal.SIGPIPE])
        except OSError as err:
            say(f"cannot start {command[0]}: {err}")
            status = 127 << 8
        while status is None:
            pid, ended = os.waitpid(-1, 0)
            if pid == main:
                status = ended
    except Stop as stop:
        interrupted = stop.args[0]
    finally:
        for signum in SIGNALS:
            signal.signal(signum, signal.SIG_IGN)

    # The command has ended or must end: whatever is left is ended too. Done when
    # waitpid reports that no child exists; an empty children() is not that proof.
    termed = set()
    kill_at = time.monotonic() + grace
    while True:
        got = reap()
        if got is None:
            break
        if got[0]:
            termed.discard(got[0])
            if got[0] == main:
                status = got[1]
            continue
        now = time.monotonic()
        if now >= kill_at + GIVE_UP:
            say(f"children of {command[0]} did not end: {children()}")
            return 128 + signal.SIGKILL
        for pid in children():
            try:
                if now >= kill_at:
                    os.kill(pid, signal.SIGKILL)
                elif pid not in termed:
                    termed.add(pid)
                    os.kill(pid, signal.SIGTERM)
            except OSError:
                continue
        time.sleep(0.05)

    os.unlink(entry)
    if interrupted or status is None:
        return 128 + (interrupted or signal.SIGTERM)
    code = os.waitstatus_to_exitcode(status)
    return code if code >= 0 else 128 - code


def raise_error(err):
    raise err


def stop(journal):
    """Whether the run behind a journal has ended, as (ended, reasons)."""
    try:
        with open(os.path.join(journal, "invocation.json"), encoding="utf-8") as record:
            run_info = json.load(record)
        with open("/proc/sys/kernel/hostname", encoding="ascii") as name:
            here = name.read().strip()
        with open("/proc/sys/kernel/random/boot_id", encoding="ascii") as boot:
            boot_id = boot.read().strip()
        recorded = (run_info["hostname"], str(run_info["uid"]), run_info["boot_id"], run_info["pid"], run_info["start"])
    except (OSError, ValueError, KeyError) as err:
        return False, [f"the run record cannot be read: {err}"]
    if (recorded[0], recorded[1]) != (here, str(os.getuid())):
        return False, [f"the run was started by uid {recorded[1]} on {recorded[0]}, not by uid {os.getuid()} on {here}"]
    with open(os.path.join(journal, "stopped"), "a", encoding="ascii"):
        pass
    if recorded[2] != boot_id:
        return True, []
    reasons = []
    try:
        if alive(recorded[3], recorded[4]):
            reasons.append(f"ansible-playbook of that run is still running (PID {recorded[3]})")
        entries = []
        commands = os.path.join(journal, "commands")
        if os.path.lexists(commands):
            for _, _, names in os.walk(commands, onerror=raise_error):
                entries += names
        for entry in sorted(entries):
            pid, _, start = entry.partition("-")
            if alive(pid, start):
                reasons.append(f"a command of that run is still running (PID {pid}); wait for it or stop it with: kill {pid}")
            else:
                reasons.append(f"the command with PID {pid} ended without confirming that nothing it started is left")
    except (OSError, IndexError) as err:
        reasons.append(f"cannot tell whether that run has ended: {err}")
    return not reasons, reasons


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--grace", type=float, default=30)
    parser.add_argument("--stop", action="store_true")
    parser.add_argument("journal")
    parser.add_argument("command", nargs=argparse.REMAINDER, help="<context> <command> [args...]")
    args = parser.parse_args()
    if args.stop:
        ended, reasons = stop(args.journal)
        print(json.dumps({"ended": ended, "reasons": reasons}))
        return 0 if ended else 1
    if len(args.command) < 2:
        parser.error("no context or no command")
    return run(args.journal, args.command[0], args.command[1:], args.grace)


if __name__ == "__main__":
    sys.exit(main())
