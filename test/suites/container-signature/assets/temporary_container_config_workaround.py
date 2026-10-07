#!/usr/bin/env python3
"""Temporarily materialize container configuration for legacy CRI-O.

This is a test-only workaround for OCPBUGS-129494.  Remove it, its unit tests,
and the signed-image-pull A/B setup when that issue is resolved:
https://redhat.atlassian.net/browse/OCPBUGS-129494

The affected runtime reads one rootful storage.conf but does not apply the
rootful storage.conf.d files.  This helper materializes the documented merged
configuration in /etc/containers/storage.conf and copies missing vendor
signature-discovery YAML files into /etc.  It never changes policy.json.

Loader and schema references:
https://raw.githubusercontent.com/containers/container-libs/main/common/docs/containers-config.5.md
https://raw.githubusercontent.com/containers/storage/main/docs/containers-storage.conf.5.md
"""

from __future__ import annotations

import argparse
import copy
import datetime
import errno
import hashlib
import json
import math
import os
from pathlib import Path
import re
import shlex
import shutil
import stat
import sys
import tempfile
from typing import Any

try:
    import tomllib
except ImportError:  # pragma: no cover - exercised by the RHEL 9 controller
    import tomli as tomllib


SYSTEM_STORAGE = Path("/usr/share/containers/storage.conf")
ADMIN_STORAGE = Path("/etc/containers/storage.conf")
SYSTEM_STORAGE_DROPINS = Path("/usr/share/containers/storage.conf.d")
SYSTEM_STORAGE_ROOTFUL_DROPINS = Path(
    "/usr/share/containers/storage.rootful.conf.d"
)
ADMIN_STORAGE_DROPINS = Path("/etc/containers/storage.conf.d")
ADMIN_STORAGE_ROOTFUL_DROPINS = Path("/etc/containers/storage.rootful.conf.d")
SYSTEM_SIGNATURES = Path("/usr/share/containers/registries.d")
ADMIN_SIGNATURES = Path("/etc/containers/registries.d")
DEFAULT_STATE_DIR = Path("/var/tmp/microshift-signed-image-pull-workaround")
STATE_FILE = "state.json"
STORAGE_BACKUP = "storage.conf.original"
BARE_KEY = re.compile(r"^[A-Za-z0-9_-]+$")
SHA256 = re.compile(r"^[0-9a-f]{64}$")
MAX_OWNER_ID = 2**32 - 2
STATE_KEYS = {
    "version",
    "status",
    "storage",
    "created_signature_files",
    "signature_directory_originally_exists",
}
STORAGE_KEYS = {
    "path",
    "original_exists",
    "backup",
    "backup_sha256",
    "mode",
    "uid",
    "gid",
    "generated_sha256",
    "inputs",
}
SIGNATURE_KEYS = {"path", "sha256"}


class WorkaroundError(RuntimeError):
    """A safe apply or restore operation could not be completed."""


def rooted(root: Path, path: Path) -> Path:
    """Return an absolute configuration path underneath root."""
    if not path.is_absolute():
        raise WorkaroundError(f"configuration path must be absolute: {path}")
    relative = path.relative_to("/")
    if ".." in relative.parts:
        raise WorkaroundError(f"configuration path escapes filesystem root: {path}")
    return root / relative


def require_regular_file(path: Path) -> os.stat_result:
    """Return lstat data for a regular, non-symlink file."""
    try:
        metadata = path.lstat()
    except FileNotFoundError as exc:
        raise WorkaroundError(f"required file is missing: {path}") from exc
    if stat.S_ISLNK(metadata.st_mode):
        raise WorkaroundError(f"refusing to follow symlink: {path}")
    if not stat.S_ISREG(metadata.st_mode):
        raise WorkaroundError(f"expected a regular file: {path}")
    return metadata


def reject_symlink_if_present(path: Path) -> None:
    try:
        metadata = path.lstat()
    except FileNotFoundError:
        return
    if stat.S_ISLNK(metadata.st_mode):
        raise WorkaroundError(f"refusing to replace symlink: {path}")
    if not stat.S_ISREG(metadata.st_mode):
        raise WorkaroundError(f"expected a regular file or absence: {path}")


def ensure_directory(path: Path, mode: int = 0o755) -> None:
    """Create path without traversing or accepting symlink components."""
    missing: list[Path] = []
    current = path
    while True:
        try:
            metadata = current.lstat()
        except FileNotFoundError:
            missing.append(current)
            parent = current.parent
            if parent == current:
                raise WorkaroundError(f"cannot find existing parent for {path}")
            current = parent
            continue
        if stat.S_ISLNK(metadata.st_mode):
            raise WorkaroundError(f"refusing to traverse symlink: {current}")
        if not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(f"expected a directory: {current}")
        break
    for directory in reversed(missing):
        try:
            directory.mkdir(mode=mode)
        except FileExistsError:
            metadata = directory.lstat()
            if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(
                metadata.st_mode
            ):
                raise WorkaroundError(
                    f"unsafe path appeared while creating {directory}"
                )


def require_no_symlink_components(root: Path, path: Path) -> os.stat_result:
    """Return final lstat data after rejecting symlinks below root."""
    try:
        relative = path.relative_to(root)
    except ValueError as exc:
        raise WorkaroundError(f"path is outside test root: {path}") from exc
    current = root
    metadata = current.lstat()
    if stat.S_ISLNK(metadata.st_mode):
        raise WorkaroundError(f"refusing to traverse symlink: {current}")
    for index, part in enumerate(relative.parts):
        current /= part
        try:
            metadata = current.lstat()
        except FileNotFoundError as exc:
            raise WorkaroundError(f"required path is missing: {current}") from exc
        if stat.S_ISLNK(metadata.st_mode):
            raise WorkaroundError(f"refusing to traverse symlink: {current}")
        if index < len(relative.parts) - 1 and not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(f"expected a directory: {current}")
    return metadata


def fsync_directory(path: Path) -> None:
    descriptor = os.open(path, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)


def atomic_write(
    path: Path,
    content: bytes,
    *,
    mode: int,
    uid: int,
    gid: int,
) -> None:
    """Write a validated file atomically without following a target symlink."""
    reject_symlink_if_present(path)
    ensure_directory(path.parent)
    descriptor, temporary_name = tempfile.mkstemp(
        prefix=f".{path.name}.", suffix=".tmp", dir=path.parent
    )
    temporary = Path(temporary_name)
    try:
        os.fchmod(descriptor, mode & 0o777)
        os.fchown(descriptor, uid, gid)
        with os.fdopen(descriptor, "wb", closefd=True) as stream:
            descriptor = -1
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
        fsync_directory(path.parent)
    except Exception:
        if descriptor >= 0:
            os.close(descriptor)
        try:
            temporary.unlink()
        except FileNotFoundError:
            pass
        raise


def atomic_create(
    path: Path,
    content: bytes,
    *,
    mode: int,
    uid: int,
    gid: int,
) -> None:
    """Atomically create path while refusing to replace a concurrent file."""
    reject_symlink_if_present(path)
    if path.exists():
        raise WorkaroundError(f"refusing to replace existing file: {path}")
    ensure_directory(path.parent)
    descriptor, temporary_name = tempfile.mkstemp(
        prefix=f".{path.name}.", suffix=".tmp", dir=path.parent
    )
    temporary = Path(temporary_name)
    try:
        os.fchmod(descriptor, mode & 0o777)
        os.fchown(descriptor, uid, gid)
        with os.fdopen(descriptor, "wb", closefd=True) as stream:
            descriptor = -1
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        os.link(temporary, path, follow_symlinks=False)
        temporary.unlink()
        fsync_directory(path.parent)
    except Exception:
        if descriptor >= 0:
            os.close(descriptor)
        try:
            temporary.unlink()
        except FileNotFoundError:
            pass
        raise


def parse_toml(path: Path) -> dict[str, Any]:
    require_regular_file(path)
    try:
        with path.open("rb") as stream:
            return tomllib.load(stream)
    except (OSError, tomllib.TOMLDecodeError) as exc:
        raise WorkaroundError(f"cannot parse TOML file {path}: {exc}") from exc


def recursive_merge(base: dict[str, Any], later: dict[str, Any]) -> None:
    """Merge tables recursively; later arrays and scalar values replace."""
    for key, value in later.items():
        if (
            key in base
            and isinstance(base[key], dict)
            and isinstance(value, dict)
        ):
            recursive_merge(base[key], value)
        else:
            base[key] = copy.deepcopy(value)


def toml_basic_string(value: str) -> str:
    """Encode a TOML basic string, including JSON's one missed control."""
    return json.dumps(value, ensure_ascii=False).replace("\x7f", "\\u007f")


def toml_key(key: str) -> str:
    return key if BARE_KEY.fullmatch(key) else toml_basic_string(key)


def toml_value(value: Any) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, int):
        return str(value)
    if isinstance(value, float):
        if math.isnan(value):
            return "-nan" if math.copysign(1.0, value) < 0 else "nan"
        if math.isinf(value):
            return "+inf" if value > 0 else "-inf"
        return repr(value)
    if isinstance(value, str):
        return toml_basic_string(value)
    if isinstance(value, (datetime.datetime, datetime.date, datetime.time)):
        return value.isoformat()
    if isinstance(value, list):
        return "[" + ", ".join(toml_value(item) for item in value) + "]"
    if isinstance(value, dict):
        entries = (
            f"{toml_key(str(key))} = {toml_value(item)}"
            for key, item in value.items()
        )
        return "{ " + ", ".join(entries) + " }"
    raise WorkaroundError(f"unsupported TOML value type: {type(value).__name__}")


def is_array_of_tables(value: Any) -> bool:
    return isinstance(value, list) and bool(value) and all(
        isinstance(item, dict) for item in value
    )


def emit_table(
    lines: list[str],
    table: dict[str, Any],
    path: tuple[str, ...],
    *,
    emit_header: bool,
) -> None:
    if emit_header:
        if lines and lines[-1] != "":
            lines.append("")
        lines.append("[" + ".".join(toml_key(key) for key in path) + "]")

    for key, value in table.items():
        if not isinstance(value, dict) and not is_array_of_tables(value):
            lines.append(f"{toml_key(str(key))} = {toml_value(value)}")

    for key, value in table.items():
        child_path = (*path, str(key))
        if isinstance(value, dict):
            emit_table(lines, value, child_path, emit_header=True)
        elif is_array_of_tables(value):
            for item in value:
                if lines and lines[-1] != "":
                    lines.append("")
                lines.append(
                    "[[" + ".".join(toml_key(part) for part in child_path) + "]]"
                )
                emit_table(lines, item, child_path, emit_header=False)


def toml_semantically_equal(left: Any, right: Any) -> bool:
    if isinstance(left, float) and isinstance(right, float):
        if math.isnan(left) and math.isnan(right):
            return math.copysign(1.0, left) == math.copysign(1.0, right)
    if isinstance(left, dict) and isinstance(right, dict):
        return left.keys() == right.keys() and all(
            toml_semantically_equal(left[key], right[key]) for key in left
        )
    if isinstance(left, list) and isinstance(right, list):
        return len(left) == len(right) and all(
            toml_semantically_equal(left_item, right_item)
            for left_item, right_item in zip(left, right)
        )
    return type(left) is type(right) and left == right


def serialize_toml(document: dict[str, Any]) -> bytes:
    """Serialize every value type produced by tomllib and verify semantics."""
    lines: list[str] = []
    emit_table(lines, document, (), emit_header=False)
    encoded = ("\n".join(lines).rstrip() + "\n").encode()
    try:
        reparsed = tomllib.loads(encoded.decode())
    except tomllib.TOMLDecodeError as exc:
        raise WorkaroundError(f"generated TOML failed validation: {exc}") from exc
    if not toml_semantically_equal(reparsed, document):
        raise WorkaroundError("generated TOML changed configuration semantics")
    return encoded


def select_storage_inputs(root: Path) -> list[Path]:
    """Select one main file and rootful drop-ins in documented precedence."""
    admin_main = rooted(root, ADMIN_STORAGE)
    system_main = rooted(root, SYSTEM_STORAGE)
    if admin_main.exists() or admin_main.is_symlink():
        main = admin_main
    else:
        main = system_main
    require_regular_file(main)

    selected: dict[str, Path] = {}
    # Sorting happens globally only after duplicates are resolved.  This list
    # is the documented low-to-high precedence order for UID 0.
    for directory in (
        SYSTEM_STORAGE_DROPINS,
        SYSTEM_STORAGE_ROOTFUL_DROPINS,
        ADMIN_STORAGE_DROPINS,
        ADMIN_STORAGE_ROOTFUL_DROPINS,
    ):
        candidate_dir = rooted(root, directory)
        if not candidate_dir.exists():
            if candidate_dir.is_symlink():
                raise WorkaroundError(
                    f"unsafe storage drop-in directory: {candidate_dir}"
                )
            continue
        metadata = candidate_dir.lstat()
        if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(f"unsafe storage drop-in directory: {candidate_dir}")
        for candidate in candidate_dir.glob("*.conf"):
            require_regular_file(candidate)
            selected[candidate.name] = candidate
    return [main, *(selected[name] for name in sorted(selected))]


def merge_storage(inputs: list[Path]) -> tuple[dict[str, Any], bytes]:
    merged: dict[str, Any] = {}
    for path in inputs:
        recursive_merge(merged, parse_toml(path))
    return merged, serialize_toml(merged)


def sha256(content: bytes) -> str:
    return hashlib.sha256(content).hexdigest()


def build_policy_backup_command(policy_path: str | Path, backup_dir: str | Path) -> str:
    """Build the remote shell command used to atomically back up policy.json."""
    script = r"""
set -eu
policy=$1
backup_dir=$2
backup=$backup_dir/policy.json
partial=$backup_dir/.policy.json.partial
if test -L "$policy"; then
    echo "Refusing to back up policy symlink" >&2
    exit 1
elif test -f "$policy"; then
    test ! -e "$backup" && test ! -L "$backup"
    test ! -e "$partial" && test ! -L "$partial"
    trap 'rm -f -- "$partial"' EXIT
    cp --archive --no-dereference -- "$policy" "$partial"
    test -f "$partial" && test ! -L "$partial"
    mv -T -- "$partial" "$backup"
    trap - EXIT
    test -f "$backup" && test ! -L "$backup"
    echo present
elif test -e "$policy"; then
    echo "Policy path is not a regular file" >&2
    exit 1
else
    echo absent
fi
""".strip()
    return " ".join(
        (
            "bash -c",
            shlex.quote(script),
            "policy-backup",
            shlex.quote(str(policy_path)),
            shlex.quote(str(backup_dir)),
        )
    )


def baseline_failure_is_signature_only(output: str, cause: str) -> bool:
    """Accept only the exact cause through CRI-O's known crictl wrappers."""
    lines = [line for line in output.splitlines() if line.strip()]
    if not lines:
        return False
    signature_prefix = "SignatureValidationFailed: "
    if not cause.startswith(signature_prefix):
        return False
    source_rejection = cause.removeprefix(signature_prefix)
    observed_description = (
        f"{signature_prefix}unable to pull image or OCI artifact: pull image "
        f"err: copying system image from manifest list: {source_rejection}; "
        "artifact err: image reference: reference is a container image, not an "
        "OCI artifact"
    )
    exact_cause = re.escape(cause)
    exact_observed_description = re.escape(observed_description)
    expected_message = re.compile(
        r"pulling image: rpc error: code = Unknown desc = "
        r"(?:verifying signatures: )?" + exact_cause
    )
    compact_log = re.compile(r"(?:FATA|ERRO)\[\d+\]\s+(?P<message>.+)")
    timestamped_log = re.compile(
        r'time="[^"]+" level=error msg="(?P<message>[^"]+)"'
    )
    pull_image_log = re.compile(
        r'E\d{4} \d{2}:\d{2}:\d{2}\.\d+\s+\d+\s+'
        r'remote_image\.go:\d+\] "PullImage from image service failed" '
        r'err="rpc error: code = Unknown desc = '
        + exact_observed_description
        + r'" image="[^"\s]+"'
    )
    fatal_log = re.compile(
        r'time="\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z" '
        r'level=fatal msg="pulling image: '
        + exact_observed_description
        + r'"'
    )
    for line in lines:
        record = line.strip()
        if record == cause:
            continue
        if pull_image_log.fullmatch(record) or fatal_log.fullmatch(record):
            continue
        match = compact_log.fullmatch(record) or timestamped_log.fullmatch(record)
        if match is None or expected_message.fullmatch(match.group("message")) is None:
            return False
    return True


def classify_baseline_result(
    return_code: int, stdout: str, stderr: str, cause: str
) -> str:
    """Classify a baseline pull without accepting mixed failure causes."""
    output = "\n".join((stdout, stderr))
    if int(return_code) == 0:
        if not stdout.strip():
            raise WorkaroundError("successful baseline pull returned no image reference")
        return "pass"
    if baseline_failure_is_signature_only(output, cause):
        return "signature_failure"
    raise WorkaroundError(
        "baseline pull failed for a cause other than the exact signature rejection"
    )


def state_path(state_dir: Path) -> Path:
    return state_dir / STATE_FILE


def write_state(state_dir: Path, state: dict[str, Any]) -> None:
    ensure_directory(state_dir, mode=0o700)
    encoded = (json.dumps(state, indent=2, sort_keys=True) + "\n").encode()
    atomic_write(
        state_path(state_dir),
        encoded,
        mode=0o600,
        uid=os.geteuid(),
        gid=os.getegid(),
    )


def require_recovery_directory(state_dir: Path, root: Path) -> None:
    metadata = require_no_symlink_components(root, state_dir)
    if not stat.S_ISDIR(metadata.st_mode):
        raise WorkaroundError(f"unsafe recovery state directory: {state_dir}")
    if metadata.st_uid != os.geteuid() or metadata.st_gid != os.getegid():
        raise WorkaroundError(f"unexpected recovery state owner: {state_dir}")
    if stat.S_IMODE(metadata.st_mode) != 0o700:
        raise WorkaroundError(f"unsafe recovery state permissions: {state_dir}")


def require_recovery_file(path: Path, root: Path) -> None:
    metadata = require_no_symlink_components(root, path)
    if not stat.S_ISREG(metadata.st_mode):
        raise WorkaroundError(f"expected a regular recovery file: {path}")
    if metadata.st_uid != os.geteuid() or metadata.st_gid != os.getegid():
        raise WorkaroundError(f"unexpected recovery file owner: {path}")
    if stat.S_IMODE(metadata.st_mode) != 0o600:
        raise WorkaroundError(f"unsafe recovery file permissions: {path}")


def require_exact_keys(value: dict[str, Any], expected: set[str], name: str) -> None:
    if set(value) != expected:
        raise WorkaroundError(f"invalid {name} keys in recovery state")


def valid_owner_id(value: Any) -> bool:
    return (
        isinstance(value, int)
        and not isinstance(value, bool)
        and 0 <= value <= MAX_OWNER_ID
    )


def valid_storage_input(path: str, *, main: bool) -> bool:
    candidate = Path(path)
    if main:
        return candidate in {SYSTEM_STORAGE, ADMIN_STORAGE}
    return (
        candidate.is_absolute()
        and candidate.parent
        in {
            SYSTEM_STORAGE_DROPINS,
            SYSTEM_STORAGE_ROOTFUL_DROPINS,
            ADMIN_STORAGE_DROPINS,
            ADMIN_STORAGE_ROOTFUL_DROPINS,
        }
        and candidate.name == path.rsplit("/", 1)[-1]
        and candidate.suffix == ".conf"
    )


def vendor_signature_basenames(root: Path) -> set[str]:
    source_dir = rooted(root, SYSTEM_SIGNATURES)
    try:
        metadata = source_dir.lstat()
    except FileNotFoundError as exc:
        raise WorkaroundError(
            f"vendor signature directory is missing: {source_dir}"
        ) from exc
    if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
        raise WorkaroundError(f"unsafe vendor signature directory: {source_dir}")
    sources = sorted(source_dir.glob("*.yaml"))
    if not sources:
        raise WorkaroundError(f"no vendor signature YAML files found in {source_dir}")
    for source in sources:
        require_regular_file(source)
    return {source.name for source in sources}


def validate_recovery_state(state: Any, root: Path) -> dict[str, Any]:
    if not isinstance(state, dict):
        raise WorkaroundError("recovery state must be a JSON object")
    require_exact_keys(state, STATE_KEYS, "top-level")
    if type(state["version"]) is not int or state["version"] != 1:
        raise WorkaroundError("unsupported recovery state version")
    if state["status"] not in {"prepared", "applied"}:
        raise WorkaroundError("invalid recovery state status")
    if type(state["signature_directory_originally_exists"]) is not bool:
        raise WorkaroundError("invalid signature directory state")

    storage = state["storage"]
    if not isinstance(storage, dict):
        raise WorkaroundError("invalid storage recovery state")
    require_exact_keys(storage, STORAGE_KEYS, "storage")
    if storage["path"] != str(ADMIN_STORAGE):
        raise WorkaroundError("unexpected storage target in recovery state")
    if type(storage["original_exists"]) is not bool:
        raise WorkaroundError("invalid original_exists in recovery state")
    original_exists = storage["original_exists"]
    expected_backup = STORAGE_BACKUP if original_exists else None
    if storage["backup"] != expected_backup:
        raise WorkaroundError("unexpected storage backup in recovery state")
    if original_exists:
        if not isinstance(storage["backup_sha256"], str) or not SHA256.fullmatch(
            storage["backup_sha256"]
        ):
            raise WorkaroundError("invalid storage backup digest")
    elif storage["backup_sha256"] is not None:
        raise WorkaroundError("unexpected digest for absent storage backup")
    if not isinstance(storage["generated_sha256"], str) or not SHA256.fullmatch(
        storage["generated_sha256"]
    ):
        raise WorkaroundError("invalid generated storage digest")
    if (
        type(storage["mode"]) is not int
        or not 0 <= storage["mode"] <= 0o777
    ):
        raise WorkaroundError("invalid storage mode in recovery state")
    if not valid_owner_id(storage["uid"]) or not valid_owner_id(storage["gid"]):
        raise WorkaroundError("invalid storage owner in recovery state")
    inputs = storage["inputs"]
    if (
        not isinstance(inputs, list)
        or not inputs
        or not all(isinstance(item, str) for item in inputs)
        or inputs[0]
        != str(ADMIN_STORAGE if original_exists else SYSTEM_STORAGE)
        or not valid_storage_input(inputs[0], main=True)
        or not all(valid_storage_input(item, main=False) for item in inputs[1:])
        or len(inputs) != len(set(inputs))
    ):
        raise WorkaroundError("invalid storage inputs in recovery state")

    entries = state["created_signature_files"]
    if not isinstance(entries, list):
        raise WorkaroundError("invalid signature recovery state")
    vendor_basenames = vendor_signature_basenames(root)
    seen_paths: set[str] = set()
    for entry in entries:
        if not isinstance(entry, dict):
            raise WorkaroundError("invalid signature entry in recovery state")
        require_exact_keys(entry, SIGNATURE_KEYS, "signature entry")
        path = entry["path"]
        if not isinstance(path, str):
            raise WorkaroundError("invalid signature path in recovery state")
        candidate = Path(path)
        if (
            not candidate.is_absolute()
            or candidate.parent != ADMIN_SIGNATURES
            or candidate.name not in vendor_basenames
            or path in seen_paths
        ):
            raise WorkaroundError("unexpected signature target in recovery state")
        seen_paths.add(path)
        if not isinstance(entry["sha256"], str) or not SHA256.fullmatch(
            entry["sha256"]
        ):
            raise WorkaroundError("invalid signature digest in recovery state")
    return state


def read_state(state_dir: Path, root: Path) -> dict[str, Any]:
    require_recovery_directory(state_dir, root)
    path = state_path(state_dir)
    require_recovery_file(path, root)
    try:
        state = json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise WorkaroundError(f"cannot read recovery state {path}: {exc}") from exc
    try:
        return validate_recovery_state(state, root)
    except WorkaroundError as exc:
        raise WorkaroundError(f"invalid recovery state in {path}: {exc}") from exc


def validate_recovery_backup(
    state: dict[str, Any], state_dir: Path, root: Path
) -> bytes | None:
    """Validate and read the storage backup named by recovery state."""
    storage = state["storage"]
    if not storage["original_exists"]:
        return None
    backup = state_dir / STORAGE_BACKUP
    require_recovery_file(backup, root)
    original = backup.read_bytes()
    if sha256(original) != storage["backup_sha256"]:
        raise WorkaroundError(f"storage backup failed integrity check: {backup}")
    return original


def relative_to_root(root: Path, path: Path) -> str:
    try:
        return "/" + str(path.relative_to(root))
    except ValueError as exc:
        raise WorkaroundError(f"path is outside test root: {path}") from exc


def signature_copies(root: Path) -> tuple[list[tuple[Path, Path, bytes]], bool]:
    source_dir = rooted(root, SYSTEM_SIGNATURES)
    basenames = vendor_signature_basenames(root)

    destination_dir = rooted(root, ADMIN_SIGNATURES)
    originally_exists = destination_dir.exists() or destination_dir.is_symlink()
    if originally_exists:
        metadata = destination_dir.lstat()
        if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(
                f"unsafe admin signature directory: {destination_dir}"
            )
    copies: list[tuple[Path, Path, bytes]] = []
    for basename in sorted(basenames):
        source = source_dir / basename
        destination = destination_dir / source.name
        if destination.exists() or destination.is_symlink():
            require_regular_file(destination)
            continue
        copies.append((source, destination, source.read_bytes()))
    return copies, originally_exists


def cleanup_uncommitted_state(state_dir: Path) -> None:
    """Remove recovery artifacts made before the first live mutation."""
    if not state_dir.exists() and not state_dir.is_symlink():
        return
    metadata = state_dir.lstat()
    if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
        raise WorkaroundError(f"unsafe incomplete recovery directory: {state_dir}")
    shutil.rmtree(state_dir)
    fsync_directory(state_dir.parent)


def apply(root: Path, state_dir: Path) -> None:
    if state_dir.exists() or state_dir.is_symlink():
        metadata = state_dir.lstat()
        if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(f"unsafe recovery state directory: {state_dir}")
        if not state_path(state_dir).exists() and not state_path(state_dir).is_symlink():
            raise WorkaroundError(
                f"recovery directory exists without valid state: {state_dir}"
            )
        state = read_state(state_dir, root)
        validate_recovery_backup(state, state_dir, root)
        if state["status"] != "applied":
            raise WorkaroundError(
                f"incomplete prior operation requires restore: {state_dir}"
            )
        target = rooted(root, Path(state["storage"]["path"]))
        require_regular_file(target)
        if sha256(target.read_bytes()) != state["storage"]["generated_sha256"]:
            raise WorkaroundError(
                f"applied storage configuration changed unexpectedly: {target}"
            )
        for entry in state["created_signature_files"]:
            path = rooted(root, Path(entry["path"]))
            require_regular_file(path)
            if sha256(path.read_bytes()) != entry["sha256"]:
                raise WorkaroundError(
                    f"applied signature configuration changed unexpectedly: {path}"
                )
        print("Temporary container configuration workaround is already applied")
        return

    target = rooted(root, ADMIN_STORAGE)
    reject_symlink_if_present(target)
    original_exists = target.exists()
    original_metadata = require_regular_file(target) if original_exists else None

    # Validate every input and the serialized result before creating recovery
    # state or touching live configuration.
    inputs = select_storage_inputs(root)
    _, storage_content = merge_storage(inputs)
    copies, signature_directory_originally_exists = signature_copies(root)

    original = target.read_bytes() if original_metadata is not None else None
    source_metadata = require_regular_file(inputs[0])
    output_metadata = original_metadata or source_metadata
    state: dict[str, Any] = {
        "version": 1,
        "status": "prepared",
        "signature_directory_originally_exists": (
            signature_directory_originally_exists
        ),
        "storage": {
            "path": str(ADMIN_STORAGE),
            "original_exists": original_exists,
            "backup": STORAGE_BACKUP if original_exists else None,
            "backup_sha256": sha256(original) if original is not None else None,
            "mode": stat.S_IMODE(output_metadata.st_mode),
            "uid": output_metadata.st_uid,
            "gid": output_metadata.st_gid,
            "generated_sha256": sha256(storage_content),
            "inputs": [relative_to_root(root, path) for path in inputs],
        },
        "created_signature_files": [
            {
                "path": relative_to_root(root, destination),
                "sha256": sha256(content),
            }
            for _, destination, content in copies
        ],
    }

    try:
        ensure_directory(state_dir, mode=0o700)
        backup = state_dir / STORAGE_BACKUP
        if original is not None:
            atomic_write(
                backup,
                original,
                mode=0o600,
                uid=os.geteuid(),
                gid=os.getegid(),
            )
        write_state(state_dir, state)
        state = read_state(state_dir, root)
    except Exception as preparation_error:
        try:
            cleanup_uncommitted_state(state_dir)
        except Exception as cleanup_error:
            raise WorkaroundError(
                f"recovery preparation failed ({preparation_error}) and cleanup "
                f"failed ({cleanup_error})"
            ) from preparation_error
        raise

    try:
        validate_recovery_backup(state, state_dir, root)
    except Exception as exc:
        raise WorkaroundError(
            f"apply preflight failed ({exc}). Recovery data retained in "
            f"{state_dir}"
        ) from exc

    try:
        write_storage = atomic_write if original_exists else atomic_create
        write_storage(
            target,
            storage_content,
            mode=state["storage"]["mode"],
            uid=state["storage"]["uid"],
            gid=state["storage"]["gid"],
        )
        for _, destination, content in copies:
            atomic_create(
                destination,
                content,
                mode=0o644,
                uid=0 if root == Path("/") else os.geteuid(),
                gid=0 if root == Path("/") else os.getegid(),
            )
        state["status"] = "applied"
        write_state(state_dir, state)
    except Exception as apply_error:
        try:
            restore(root, state_dir, remove_state=False)
        except Exception as restore_error:
            raise WorkaroundError(
                f"apply failed ({apply_error}). Rollback also failed "
                f"({restore_error}). Recovery data retained in {state_dir}"
            ) from apply_error
        raise WorkaroundError(
            f"apply failed and was rolled back. Recovery data retained in "
            f"{state_dir}: {apply_error}"
        ) from apply_error

    print(
        f"Applied temporary workaround using {len(inputs)} storage TOML files "
        f"and {len(copies)} missing signature registry files"
    )


def preflight_restore(
    root: Path, state_dir: Path
) -> tuple[dict[str, Any], Path, str, bytes | None, list[Path], bool]:
    """Validate every recovery input and live target before any mutation."""
    state = read_state(state_dir, root)
    storage = state["storage"]
    target = rooted(root, Path(storage["path"]))
    storage_action = "none"
    original = validate_recovery_backup(state, state_dir, root)
    if storage["original_exists"]:
        target_metadata = require_regular_file(target)
        current_digest = sha256(target.read_bytes())
        if current_digest == storage["generated_sha256"]:
            storage_action = "write"
        elif current_digest != storage["backup_sha256"]:
            raise WorkaroundError(
                "generated storage configuration changed after apply; "
                "refusing to overwrite it"
            )
        elif (
            stat.S_IMODE(target_metadata.st_mode) != storage["mode"]
            or target_metadata.st_uid != storage["uid"]
            or target_metadata.st_gid != storage["gid"]
        ):
            storage_action = "write"
    elif target.exists() or target.is_symlink():
        require_regular_file(target)
        if sha256(target.read_bytes()) != storage["generated_sha256"]:
            raise WorkaroundError(
                "generated storage configuration changed after apply; "
                "refusing to remove it"
            )
        storage_action = "unlink"

    signature_removals: list[Path] = []
    for entry in state["created_signature_files"]:
        path = rooted(root, Path(entry["path"]))
        if not path.exists() and not path.is_symlink():
            continue
        require_regular_file(path)
        if sha256(path.read_bytes()) != entry["sha256"]:
            raise WorkaroundError(
                f"created file changed after apply, refusing to remove it: {path}"
            )
        signature_removals.append(path)
    remove_signature_directory = not state[
        "signature_directory_originally_exists"
    ]
    if remove_signature_directory:
        destination_dir = rooted(root, ADMIN_SIGNATURES)
        if destination_dir.exists() or destination_dir.is_symlink():
            metadata = destination_dir.lstat()
            if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
                raise WorkaroundError(
                    f"unsafe admin signature directory: {destination_dir}"
                )
    return (
        state,
        target,
        storage_action,
        original,
        signature_removals,
        remove_signature_directory,
    )


def restore(root: Path, state_dir: Path, *, remove_state: bool = True) -> None:
    try:
        (
            state,
            target,
            storage_action,
            original,
            signature_removals,
            remove_signature_directory,
        ) = preflight_restore(root, state_dir)
    except Exception as exc:
        raise WorkaroundError(
            f"restore preflight failed ({exc}). Recovery data retained in "
            f"{state_dir}"
        ) from exc
    storage = state["storage"]

    try:
        if storage_action == "write":
            if original is None:  # Defensive: guaranteed by preflight_restore.
                raise WorkaroundError("original storage content is unavailable")
            atomic_write(
                target,
                original,
                mode=storage["mode"],
                uid=storage["uid"],
                gid=storage["gid"],
            )
        elif storage_action == "unlink":
            target.unlink()
            fsync_directory(target.parent)

        if storage["original_exists"]:
            restored_metadata = require_regular_file(target)
            if (
                sha256(target.read_bytes()) != storage["backup_sha256"]
                or stat.S_IMODE(restored_metadata.st_mode) != storage["mode"]
                or restored_metadata.st_uid != storage["uid"]
                or restored_metadata.st_gid != storage["gid"]
            ):
                raise WorkaroundError(
                    f"restored storage content or metadata is incorrect: {target}"
                )
        elif target.exists() or target.is_symlink():
            raise WorkaroundError(f"generated storage target still exists: {target}")

        for path in signature_removals:
            path.unlink()
            fsync_directory(path.parent)
        if remove_signature_directory:
            destination_dir = rooted(root, ADMIN_SIGNATURES)
            try:
                destination_dir.rmdir()
            except FileNotFoundError:
                pass
            except OSError as exc:
                if exc.errno not in {errno.ENOTEMPTY, errno.EEXIST}:
                    raise
            else:
                fsync_directory(destination_dir.parent)
    except Exception as exc:
        raise WorkaroundError(
            f"restore failed ({exc}). Recovery data retained in {state_dir}"
        ) from exc

    if remove_state:
        try:
            shutil.rmtree(state_dir)
            fsync_directory(state_dir.parent)
        except Exception as exc:
            raise WorkaroundError(
                f"configuration restored but recovery cleanup failed ({exc}). "
                f"Recovery data retained in {state_dir}"
            ) from exc
    print("Restored container configuration changed by the temporary workaround")


def parse_args(arguments: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("apply", "restore"))
    parser.add_argument(
        "--root",
        type=Path,
        default=Path("/"),
        help="alternate filesystem root for unit tests",
    )
    parser.add_argument(
        "--state-dir",
        type=Path,
        default=DEFAULT_STATE_DIR,
        help="absolute recovery directory inside --root",
    )
    parser.add_argument(
        "--keep-state",
        action="store_true",
        help="restore files but retain recovery state after a failed apply",
    )
    return parser.parse_args(arguments)


def main(arguments: list[str] | None = None) -> int:
    args = parse_args(arguments if arguments is not None else sys.argv[1:])
    root = args.root.resolve()
    if not root.is_dir():
        raise WorkaroundError(f"filesystem root is not a directory: {root}")
    state_dir = rooted(root, args.state_dir)
    if args.action == "apply":
        if args.keep_state:
            raise WorkaroundError("--keep-state is only valid with restore")
        apply(root, state_dir)
    else:
        restore(root, state_dir, remove_state=not args.keep_state)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except WorkaroundError as error:
        print(f"error: {error}", file=sys.stderr)
        raise SystemExit(1) from error
