from robot.libraries.BuiltIn import BuiltIn

import libostree

_log = BuiltIn().log


def get_journal_cursor() -> str:
    """Return the cursor value for the MicroShift logs.

    Return a value that can be used to find new MicroShift log
    output. See --cursor argument to journalctl for details.
    """
    stdout, rc = libostree.remote_sudo_rc(
        "journalctl -u microshift --show-cursor --no-pager -n 0"
    )
    BuiltIn().should_be_equal_as_integers(rc, 0)
    # Produces output like:
    # -- No entries --
    # -- cursor: s=d1b6ab3ee650471cacc2a5f694b500a6;i=1772e;b=efc55827851c451db78f496a81a9de88;m=29024a3bb1;t=605bddb046dcb;x=f3b4472fb6bbb542
    lines = stdout.splitlines()
    cursor_line = lines[-1]
    cursor = cursor_line.partition(': ')[-1]
    return cursor


def get_log_output_with_pattern(cursor: str, pattern: str) -> tuple[str, int]:
    """Get the logs since the cursor matching the pattern and return the log content and exit code."""
    stdout, rc = libostree.remote_sudo_rc(
        f"journalctl -u microshift --cursor='{cursor}' --no-pager --grep '{pattern}'"
    )
    BuiltIn().log(f"log lines matching '{pattern}':\n{stdout}")
    return stdout, rc


def pattern_should_not_appear_in_log_output(cursor, pattern):
    """Get the logs since the cursor and verify that the pattern does not appear."""
    stdout, rc = get_log_output_with_pattern(cursor, pattern)
    # The grep argument causes journalctl to exit with an error if the
    # pattern is not found, therefore we want the return code to be 1,
    # indicating that there was no match.
    BuiltIn().should_be_equal_as_integers(rc, 1)


def pattern_should_appear_in_log_output(cursor, pattern):
    """Get the logs since the cursor and verify that the pattern does not appear."""
    stdout, rc = get_log_output_with_pattern(cursor, pattern)
    # The grep argument causes journalctl to exit with an error if the
    # pattern is not found, therefore we want the return code to be 0,
    # indicating that there was a match.
    BuiltIn().should_be_equal_as_integers(rc, 0)
