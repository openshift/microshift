"""Collect allowlisted, publish-safe greenboot health-check failure summaries.

MicroShift CI strips output.xml / rf-debug.log from public PR artifacts and
there is no approved private sink for raw health-check output, so this module
deliberately trades generality for a hard privacy guarantee: it never returns,
logs, or writes raw command output. Instead it opens an isolated SSH
connection, runs a single `systemctl show` of fixed allowlisted properties,
validates every value against an allowlist, and writes only a fixed-schema
key=value summary.

This per-keyword redaction is intentionally bespoke. A shared "publish-safe
diagnostic" utility, or an approved private artifact destination, would be the
more general fix; until one exists, keep new diagnostics inside this allowlist
model rather than logging raw output elsewhere.
"""
import os
import re
import time
import uuid

from robot.api import logger
from robot.libraries.BuiltIn import BuiltIn


_UNIT_NAME = "greenboot-healthcheck.service"
# The two triggers differ in how the collected service_* fields relate to the
# failure that prompted collection:
#   greenboot-final-wait - we were waiting on greenboot-healthcheck.service
#       itself, so its live state (ActiveState/SubState/Result/ExecMainStatus)
#       is the direct cause of the failure.
#   direct-healthcheck - an ad-hoc `microshift healthcheck` CLI invocation just
#       failed. greenboot-healthcheck.service runs the *same* check but at boot
#       (packaging/greenboot/microshift-running-check.sh), so the service_*
#       fields reflect that earlier boot-time run, NOT the invocation that just
#       failed. Use primary_exit_code for the current failure; read service_*
#       only as "was the same check healthy at boot" (i.e. is this a regression
#       since boot). The raw -v=2 output naming the component is intentionally
#       not published - there is no approved private artifact destination.
_TRIGGERS = {"direct-healthcheck", "greenboot-final-wait"}
_ACTIVE_STATES = {
    "active",
    "activating",
    "deactivating",
    "failed",
    "inactive",
    "maintenance",
    "refreshing",
    "reloading",
}
_SUB_STATES = {
    "auto-restart",
    "auto-restart-queued",
    "cleaning",
    "condition",
    "dead",
    "exited",
    "failed",
    "final-sigkill",
    "final-sigterm",
    "final-watchdog",
    "reload",
    "reload-notify",
    "reload-signal",
    "running",
    "start",
    "start-post",
    "start-pre",
    "stop",
    "stop-post",
    "stop-sigkill",
    "stop-sigterm",
    "stop-watchdog",
}
_RESULTS = {
    "assert",
    "condition",
    "core-dump",
    "exit-code",
    "exec-condition",
    "oom-kill",
    "protocol",
    "resources",
    "signal",
    "skipped",
    "start-limit-hit",
    "success",
    "timeout",
    "watchdog",
}
# systemd property name -> (validation rule, summary field name). Keeping the
# rule and the emitted field in one entry removes the hand-sync hazard of two
# dicts keyed by the same names.
_PROPERTIES = {
    "ActiveEnterTimestampMonotonic": ("unsigned", "active_enter_timestamp_monotonic_us"),
    "ActiveState": (_ACTIVE_STATES, "service_active_state"),
    "ExecMainStatus": ("exit_code", "exec_main_status"),
    "InactiveEnterTimestampMonotonic": ("unsigned", "inactive_enter_timestamp_monotonic_us"),
    "Result": (_RESULTS, "service_result"),
    "SubState": (_SUB_STATES, "service_sub_state"),
}
_PUBLIC_COLLECTION_RESULTS = {"complete", "failed", "partial"}
_PUBLIC_ERROR_CATEGORIES = {
    "artifact_write_failed",
    "collection_failed",
    "connection_restore_failed",
    "invalid_response",
    "none",
    "query_failed",
    "reconnect_failed",
}
_PUBLIC_FILENAME = re.compile(
    r"greenboot-healthcheck-summary-[0-9a-f]{32}\.txt"
)


def _safe_integer(value, maximum=None):
    text = str(value)
    if not re.fullmatch(r"[0-9]+", text):
        return None
    number = int(text)
    if maximum is not None and number > maximum:
        return None
    return str(number)


def _safe_timeout(value):
    text = str(value)
    if not re.fullmatch(r"[1-9][0-9]*s", text):
        raise ValueError("invalid timeout")
    return text


def _safe_port(value):
    if value in (None, ""):
        return 22
    number = _safe_integer(value, 65535)
    if number is None or int(number) == 0:
        raise ValueError("invalid port")
    return int(number)


def _parse_properties(output):
    values = {}
    invalid = False
    for line in output.splitlines():
        if "=" not in line:
            invalid = True
            continue
        key, value = line.split("=", 1)
        if key not in _PROPERTIES or key in values:
            invalid = True
            continue
        rule = _PROPERTIES[key][0]
        if rule == "unsigned":
            safe_value = _safe_integer(value)
        elif rule == "exit_code":
            safe_value = _safe_integer(value, 255)
        else:
            safe_value = value if value in rule else None
        if safe_value is None:
            invalid = True
            values[key] = "unavailable"
        else:
            values[key] = safe_value

    if set(values) != set(_PROPERTIES):
        invalid = True
    fields = {
        field: values.get(prop, "unavailable")
        for prop, (_rule, field) in _PROPERTIES.items()
    }
    return fields, invalid


def _query_service_on_isolated_connection(
        ssh, built_in, reconnect_timeout, remote_timeout, command_timeout):
    host = built_in.get_variable_value("${USHIFT_HOST}")
    user = built_in.get_variable_value("${USHIFT_USER}")
    port = _safe_port(built_in.get_variable_value("${SSH_PORT}"))
    key = built_in.get_variable_value("${SSH_PRIV_KEY}")

    previous_connection = None
    diagnostic_connection = None
    output = None
    return_code = None
    error_category = "none"
    try:
        try:
            previous_connection = ssh.get_connection().index
        except Exception:
            previous_connection = None

        try:
            diagnostic_connection = ssh.open_connection(
                host,
                port=port,
                timeout=reconnect_timeout,
            )
            if key:
                ssh.login_with_public_key(
                    user,
                    key,
                    keep_alive_interval=30,
                )
            else:
                ssh.login(
                    user,
                    allow_agent=True,
                    keep_alive_interval=30,
                )
        except Exception:
            error_category = "reconnect_failed"
        else:
            try:
                output, return_code = _query_service(
                    ssh,
                    remote_timeout,
                    command_timeout,
                )
            except Exception:
                error_category = "query_failed"
    finally:
        if diagnostic_connection is not None:
            try:
                ssh.close_connection()
            except Exception:
                error_category = "connection_restore_failed"
        if previous_connection is not None:
            try:
                ssh.switch_connection(previous_connection)
            except Exception:
                error_category = "connection_restore_failed"

    return output, return_code, error_category


def _query_service(ssh, remote_timeout, command_timeout):
    properties = " ".join(
        f"--property={name}" for name in _PROPERTIES
    )
    command = (
        "timeout --signal=TERM --kill-after=5s "
        f"{remote_timeout} systemctl show --no-pager "
        f"{properties} {_UNIT_NAME}"
    )
    output, return_code = ssh.execute_command(
        command,
        sudo=True,
        return_stdout=True,
        return_stderr=False,
        return_rc=True,
        timeout=command_timeout,
    )
    return output, return_code


def _safe_public_result(result):
    try:
        saved = result["saved"]
        filename = result["filename"]
        collection_result = result["collection_result"]
        error_category = result["error_category"]
    except (KeyError, TypeError):
        return None

    if type(saved) is not bool:
        return None
    if collection_result not in _PUBLIC_COLLECTION_RESULTS:
        return None
    if error_category not in _PUBLIC_ERROR_CATEGORIES:
        return None
    expected_result = (
        "complete" if error_category == "none" else
        "partial" if error_category == "invalid_response" else
        "failed"
    )
    if collection_result != expected_result:
        return None
    if saved:
        if error_category == "artifact_write_failed":
            return None
        if not isinstance(filename, str):
            return None
        if _PUBLIC_FILENAME.fullmatch(filename) is None:
            return None
    else:
        if error_category != "artifact_write_failed":
            return None
        if filename != "unavailable":
            return None

    return {
        "saved": saved,
        "filename": filename,
        "collection_result": collection_result,
        "error_category": error_category,
    }


def _resolve_library_keyword(built_in, keyword_name):
    library_name, separator, method_name = str(keyword_name).partition(".")
    if not separator or not library_name or not method_name:
        raise ValueError("invalid diagnostic keyword")
    library = built_in.get_library_instance(library_name)
    method = method_name.lower().replace(" ", "_")
    return getattr(library, method)


def run_public_greenboot_diagnostics_safely(
        collector_keyword, trigger, primary_exit_code=None,
        reconnect_timeout="10s", remote_timeout="25s",
        command_timeout="30s"):
    """Run diagnostics without exposing collector failures or return values."""
    built_in = BuiltIn()
    old_level = None
    result = None
    try:
        # Suppress logging around the collector call itself. This guards any
        # collector keyword, including ones that do not set their own level
        # (e.g. a stub that logs before failing). The production collector also
        # sets NONE internally; the two layers are intentional defense in depth.
        old_level = built_in.set_log_level("NONE")
        collector = _resolve_library_keyword(built_in, collector_keyword)
        untrusted_result = collector(
            trigger,
            primary_exit_code,
            reconnect_timeout=reconnect_timeout,
            remote_timeout=remote_timeout,
            command_timeout=command_timeout,
        )
        result = _safe_public_result(untrusted_result)
    except Exception:
        result = None
    finally:
        if old_level is not None:
            try:
                built_in.set_log_level(old_level)
            except Exception:
                # Never leave the suite stuck at NONE: fall back to a level that
                # keeps logging on so later keywords stay diagnosable.
                result = None
                try:
                    built_in.set_log_level("INFO")
                except Exception:
                    pass

    try:
        if result is not None and result["saved"]:
            logger.info(
                "Public greenboot diagnostic summary saved as "
                f"{result['filename']} "
                f"(collection_result={result['collection_result']}, "
                f"error_category={result['error_category']})"
            )
        else:
            error_category = (
                result["error_category"]
                if result is not None else "collection_failed"
            )
            logger.warn(
                "Public greenboot diagnostic summary unavailable "
                f"(error_category={error_category})"
            )
    except Exception:
        pass


def _summary_text(fields):
    order = (
        "schema_version",
        "trigger",
        "collection_result",
        "error_category",
        "collection_elapsed_ms",
        "primary_exit_code",
        "query_exit_code",
        "service_active_state",
        "service_sub_state",
        "service_result",
        "exec_main_status",
        "active_enter_timestamp_monotonic_us",
        "inactive_enter_timestamp_monotonic_us",
    )
    return "".join(f"{key}={fields[key]}\n" for key in order)


def _write_summary(output_dir, summary):
    for _ in range(10):
        filename = (
            "greenboot-healthcheck-summary-"
            f"{uuid.uuid4().hex}.txt"
        )
        path = os.path.join(output_dir, filename)
        try:
            # Create at 0600 atomically: no window at the default umask, and no
            # separate chmod that could fail after the file already exists.
            fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        except FileExistsError:
            continue
        try:
            with os.fdopen(fd, "w", encoding="utf-8") as stream:
                stream.write(summary)
        except Exception:
            # Keep saved=False consistent with no artifact on disk.
            try:
                os.unlink(path)
            except OSError:
                pass
            raise
        return filename
    raise OSError("unable to allocate diagnostic filename")


def _collect(trigger, primary_exit_code, ssh_library_name,
             reconnect_timeout, remote_timeout, command_timeout):
    built_in = BuiltIn()
    started = time.monotonic_ns()
    fields = {
        "schema_version": "1",
        "trigger": trigger if trigger in _TRIGGERS else "unavailable",
        "collection_result": "failed",
        "error_category": "none",
        "primary_exit_code": "unavailable",
        "query_exit_code": "unavailable",
    }
    fields.update({field: "unavailable"
                   for _rule, field in _PROPERTIES.values()})

    # Suppress logging around the SSH login/exec, where SSHLibrary would log the
    # server MOTD at INFO. This is the guard for callers that reach the collector
    # directly (bypassing run_public_greenboot_diagnostics_safely); it is
    # exercised by the "Login Banner Is Suppressed By The Log Level Guard" test.
    old_level = built_in.set_log_level("NONE")
    try:
        reconnect_timeout = _safe_timeout(reconnect_timeout)
        remote_timeout = _safe_timeout(remote_timeout)
        command_timeout = _safe_timeout(command_timeout)
        if trigger not in _TRIGGERS:
            fields["error_category"] = "invalid_response"
        if primary_exit_code is not None:
            safe_exit_code = _safe_integer(primary_exit_code, 255)
            if safe_exit_code is None:
                fields["error_category"] = "invalid_response"
            else:
                fields["primary_exit_code"] = safe_exit_code

        ssh = built_in.get_library_instance(ssh_library_name)
        output, return_code, query_error = (
            _query_service_on_isolated_connection(
                ssh,
                built_in,
                reconnect_timeout,
                remote_timeout,
                command_timeout,
            )
        )
        if query_error != "none":
            fields["error_category"] = query_error
        else:
            safe_return_code = _safe_integer(return_code, 255)
            if safe_return_code is None:
                fields["error_category"] = "invalid_response"
            else:
                fields["query_exit_code"] = safe_return_code
                if safe_return_code != "0":
                    fields["error_category"] = "query_failed"
                else:
                    parsed, invalid = _parse_properties(output)
                    fields.update(parsed)
                    if invalid:
                        fields["error_category"] = "invalid_response"

        if fields["error_category"] == "none":
            fields["collection_result"] = "complete"
        elif fields["error_category"] == "invalid_response":
            fields["collection_result"] = "partial"
    except Exception:
        fields["error_category"] = "collection_failed"
        fields["collection_result"] = "failed"
    finally:
        elapsed = (time.monotonic_ns() - started) // 1_000_000
        fields["collection_elapsed_ms"] = str(max(0, elapsed))

    try:
        output_dir = built_in.get_variable_value("${OUTPUTDIR}")
        filename = _write_summary(output_dir, _summary_text(fields))
        result = {
            "saved": True,
            "filename": filename,
            "collection_result": fields["collection_result"],
            "error_category": fields["error_category"],
        }
    except Exception:
        result = {
            "saved": False,
            "filename": "unavailable",
            "collection_result": "failed",
            "error_category": "artifact_write_failed",
        }
    finally:
        built_in.set_log_level(old_level)
    return result


def collect_public_greenboot_diagnostic_summary(
        trigger, primary_exit_code=None, reconnect_timeout="10s",
        remote_timeout="25s", command_timeout="30s"):
    """Collect and save only validated, allowlisted public diagnostics."""
    return _collect(
        trigger,
        primary_exit_code,
        "SSHLibrary",
        reconnect_timeout,
        remote_timeout,
        command_timeout,
    )


def _collect_public_greenboot_diagnostic_summary_using_library(
        trigger, ssh_library_name, primary_exit_code=None,
        reconnect_timeout="10s", remote_timeout="25s",
        command_timeout="30s"):
    """Test seam for exercising the collector without a live host.

    Underscore-prefixed so Robot does not expose it as a keyword: it is a
    Python-level injection point called only by the unit test library, not part
    of this security-sensitive module's public keyword surface.
    """
    return _collect(
        trigger,
        primary_exit_code,
        ssh_library_name,
        reconnect_timeout,
        remote_timeout,
        command_timeout,
    )
