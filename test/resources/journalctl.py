from robot.libraries.BuiltIn import BuiltIn

import libostree
import time

_log = BuiltIn().log


def get_journal_cursor(unit="microshift") -> str:
    """Return the cursor value for the unit's logs.

    Return a value that can be used to find new service log
    output. See --cursor argument to journalctl for details.

    Optional argument `unit` may be used to specify a systemd unit other than microshift, for example microshift-observability.service
    """
    stdout, rc = libostree.remote_sudo_rc(
        f"journalctl -u {unit} --show-cursor --no-pager -n 0"
    )
    BuiltIn().should_be_equal_as_integers(rc, 0)
    # Produces output like:
    # -- No entries --
    # -- cursor: s=d1b6ab3ee650471cacc2a5f694b500a6;i=1772e;b=efc55827851c451db78f496a81a9de88;m=29024a3bb1;t=605bddb046dcb;x=f3b4472fb6bbb542
    lines = stdout.splitlines()
    cursor_line = lines[-1]
    cursor = cursor_line.partition(': ')[-1]
    return cursor


def get_log_output_with_pattern(cursor: str, pattern: str, unit="microshift") -> tuple[str, int]:
    """
    Get the logs since the cursor matching the pattern and return the log content and exit code.
    Optional argument `unit` may be used to specify a systemd unit other than microshift,
    for example microshift-observability.service.
    """
    stdout, rc = libostree.remote_sudo_rc(
        f"journalctl -u {unit} --cursor='{cursor}' --no-pager --grep '{pattern}'"
    )
    BuiltIn().log(f"log lines matching '{pattern}':\n{stdout}")
    return stdout, rc


def pattern_should_not_appear_in_log_output(cursor, pattern, unit="microshift", retries=30, wait=10):
    """Get the logs since the cursor and verify that the pattern does not appear."""
    # The grep argument causes journalctl to exit with an error if the
    # pattern is not found, therefore we want the return code to be 1,
    # indicating that there was no match.

    for attempt in range(1, retries + 2):
        stdout, rc = get_log_output_with_pattern(cursor, pattern, unit)
        if rc == 1 or attempt > retries:
            BuiltIn().should_be_equal_as_integers(rc, 1)
            return
        BuiltIn().log(f"Attempt {attempt}/{retries} failed. Retrying in {wait}s...")
        time.sleep(wait)


def pattern_should_appear_in_log_output(cursor, pattern, unit="microshift", retries=30, wait=10):
    """Get the logs since the cursor and verify that the pattern does appear."""
    # The grep argument causes journalctl to exit with an error if the
    # pattern is not found, therefore we want the return code to be 0,
    # indicating that there was a match.

    for attempt in range(1, retries + 2):
        stdout, rc = get_log_output_with_pattern(cursor, pattern, unit)
        if rc == 0 or attempt > retries:
            BuiltIn().should_be_equal_as_integers(rc, 0)
            return
        BuiltIn().log(f"Attempt {attempt}/{retries} failed. Retrying in {wait}s...")
        time.sleep(wait)
