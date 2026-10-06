import os
import time
from types import SimpleNamespace

import greenboot_diagnostics
from robot.utils import is_truthy


class GreenbootDiagnosticsTestlib:
    ROBOT_LIBRARY_SCOPE = "SUITE"

    def __init__(self):
        self.mode = "valid"
        self.reconnect_timeout = None
        self.command_timeout = None
        self.command = None
        self.return_stderr = None
        self.healthcheck_calls = 0
        self.healthcheck_command = None
        self.healthcheck_options = None
        self.healthcheck_return_code = 0
        self.diagnostic_calls = 0
        self.diagnostic_trigger = None
        self.diagnostic_primary_rc = None
        self.current_connection = 1
        self.next_connection = 2
        self.authenticated_connections = {1}
        self.clock_restore_calls = 0

    def reset_diagnostic_ssh(self):
        self.mode = "valid"
        self.reconnect_timeout = None
        self.command_timeout = None
        self.command = None
        self.return_stderr = None
        self.healthcheck_calls = 0
        self.healthcheck_command = None
        self.healthcheck_options = None
        self.healthcheck_return_code = 0
        self.diagnostic_calls = 0
        self.diagnostic_trigger = None
        self.diagnostic_primary_rc = None
        self.current_connection = 1
        self.next_connection = 2
        self.authenticated_connections = {1}
        self.clock_restore_calls = 0

    def set_diagnostic_ssh_mode(self, mode):
        self.mode = mode

    def set_healthcheck_return_code(self, return_code):
        self.healthcheck_return_code = int(return_code)

    def execute_sensitive_healthcheck_command(self, command, **options):
        self.healthcheck_calls += 1
        self.healthcheck_command = command
        self.healthcheck_options = options
        sentinel = os.environ["ROBOT_PRIVACY_SENTINEL"]
        stdout = f"stdout-{sentinel}"
        stderr = f"stderr-{sentinel}"
        result = []
        if is_truthy(options.get("return_stdout", True)):
            result.append(stdout)
        if is_truthy(options.get("return_stderr", False)):
            result.append(stderr)
        if is_truthy(options.get("return_rc", False)):
            result.append(self.healthcheck_return_code)
        if len(result) == 1:
            return result[0]
        return result

    def get_connection(self):
        if self.current_connection is None:
            raise RuntimeError("no current connection")
        return SimpleNamespace(index=self.current_connection)

    def close_connection(self):
        self.authenticated_connections.discard(self.current_connection)
        self.current_connection = None

    def open_connection(self, host, port=22, timeout=None, **_options):
        self.reconnect_timeout = timeout
        connection = self.next_connection
        self.next_connection += 1
        self.current_connection = connection
        return connection

    def switch_connection(self, connection):
        if connection not in self.authenticated_connections:
            raise RuntimeError("connection is not authenticated")
        self.current_connection = connection

    def login(self, user, **_options):
        if self.mode == "reconnect-failure":
            raise RuntimeError(os.environ["ROBOT_PRIVACY_SENTINEL"])
        self.authenticated_connections.add(self.current_connection)
        return None

    def login_with_public_key(self, user, key, **_options):
        return self.login(user, **_options)

    def execute_command(self, command, **options):
        if self.current_connection not in self.authenticated_connections:
            raise RuntimeError("connection is not authenticated")
        self.command = command
        self.command_timeout = options.get("timeout")
        self.return_stderr = options.get("return_stderr")
        sentinel = os.environ["ROBOT_PRIVACY_SENTINEL"]
        if self.mode == "query-exception":
            raise RuntimeError(sentinel)
        if self.mode == "elapsed-timeout":
            timeout = int(options["timeout"].removesuffix("s"))
            time.sleep(timeout + 0.01)
            raise TimeoutError(sentinel)
        if self.mode == "query-failure":
            return sentinel, 23
        if self.mode == "invalid-response":
            return self._invalid_output(sentinel), 0
        return self._valid_output(), 0

    def collect_stub_diagnostic_summary(
            self, trigger, primary_exit_code=None, reconnect_timeout="10s",
            remote_timeout="25s", command_timeout="30s"):
        self.diagnostic_calls += 1
        self.diagnostic_trigger = trigger
        self.diagnostic_primary_rc = primary_exit_code
        collector = (
            greenboot_diagnostics.
            collect_public_greenboot_diagnostic_summary_using_library
        )
        return collector(
            trigger,
            "DiagnosticSSH",
            primary_exit_code,
            reconnect_timeout=reconnect_timeout,
            remote_timeout=remote_timeout,
            command_timeout=command_timeout,
        )

    def raise_sensitive_diagnostic_exception(self, *_args, **_options):
        raise RuntimeError(os.environ["ROBOT_PRIVACY_SENTINEL"])

    def return_sensitive_malformed_diagnostic_result(
            self, *_args, **_options):
        sentinel = os.environ["ROBOT_PRIVACY_SENTINEL"]
        return {
            "saved": True,
            "filename": sentinel,
            "collection_result": sentinel,
            "error_category": sentinel,
        }

    def simulate_clock_restore_cleanup(self):
        if self.current_connection != 1:
            raise RuntimeError("prior connection was not restored")
        if self.current_connection not in self.authenticated_connections:
            raise RuntimeError("prior connection is not authenticated")
        self.clock_restore_calls += 1

    @staticmethod
    def _valid_output():
        return "\n".join((
            "ActiveEnterTimestampMonotonic=123456",
            "ActiveState=failed",
            "ExecMainStatus=7",
            "InactiveEnterTimestampMonotonic=234567",
            "Result=exit-code",
            "SubState=failed",
        ))

    @staticmethod
    def _invalid_output(sentinel):
        return "\n".join((
            "ActiveEnterTimestampMonotonic=123456",
            f"ActiveState={sentinel}",
            "ExecMainStatus=9999",
            "InactiveEnterTimestampMonotonic=234567",
            f"Result={sentinel}",
            f"SubState={sentinel}",
            f"UnexpectedProperty={sentinel}",
        ))

    def get_diagnostic_ssh_observations(self):
        return {
            "reconnect_timeout": self.reconnect_timeout,
            "command_timeout": self.command_timeout,
            "command": self.command,
            "return_stderr": self.return_stderr,
            "calls": self.diagnostic_calls,
            "trigger": self.diagnostic_trigger,
            "primary_rc": self.diagnostic_primary_rc,
            "current_connection": self.current_connection,
            "current_connection_authenticated": (
                self.current_connection in self.authenticated_connections
            ),
            "clock_restore_calls": self.clock_restore_calls,
        }

    def get_healthcheck_observations(self):
        return {
            "calls": self.healthcheck_calls,
            "command": self.healthcheck_command,
            "options": self.healthcheck_options,
        }
