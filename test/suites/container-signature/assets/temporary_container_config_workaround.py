#!/usr/bin/env python3
"""Materialize rootful container configuration for the COS 10 test VM.

TEMPORARY until OCPBUGS-129494 is resolved. Remove this helper and the A/B
steps in signed-image-pull.robot after the fix is available:
https://redhat.atlassian.net/browse/OCPBUGS-129494

CRI-O 1.35 reads the files below /etc/containers but does not consume the
rootful storage drop-ins used by the COS 10 image. This one-shot test helper
copies only missing vendor files and writes one semantically merged
/etc/containers/storage.conf. It is intended only for a disposable CI VM.
"""

from __future__ import annotations

import argparse
import copy
import datetime
import json
import math
import os
from pathlib import Path
import stat
import sys
import tempfile
from typing import Any
import tomllib


SYSTEM_CONTAINERS = Path("/usr/share/containers")
ADMIN_CONTAINERS = Path("/etc/containers")
SYSTEM_STORAGE = SYSTEM_CONTAINERS / "storage.conf"
ADMIN_STORAGE = ADMIN_CONTAINERS / "storage.conf"
STORAGE_DROPINS = (
    SYSTEM_CONTAINERS / "storage.conf.d",
    SYSTEM_CONTAINERS / "storage.rootful.conf.d",
    ADMIN_CONTAINERS / "storage.conf.d",
    ADMIN_CONTAINERS / "storage.rootful.conf.d",
)
VENDOR_FILES = ("policy.json", "registries.conf")
VENDOR_DIRECTORIES = (
    ("registries.conf.d", "*.conf"),
    ("registries.d", "*.yaml"),
)
BARE_KEY = frozenset(
    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
)


class WorkaroundError(RuntimeError):
    """The temporary configuration could not be applied or restored safely."""


def rooted(root: Path, path: Path) -> Path:
    if not path.is_absolute() or ".." in path.parts:
        raise WorkaroundError(f"invalid absolute configuration path: {path}")
    return root / path.relative_to("/")


def require_regular(path: Path) -> os.stat_result:
    try:
        metadata = path.lstat()
    except FileNotFoundError as error:
        raise WorkaroundError(f"required file is missing: {path}") from error
    if stat.S_ISLNK(metadata.st_mode):
        raise WorkaroundError(f"refusing to follow symlink: {path}")
    if not stat.S_ISREG(metadata.st_mode):
        raise WorkaroundError(f"expected a regular file: {path}")
    return metadata


def reject_unsafe_target(path: Path) -> None:
    try:
        metadata = path.lstat()
    except FileNotFoundError:
        return
    if stat.S_ISLNK(metadata.st_mode):
        raise WorkaroundError(f"refusing to replace symlink: {path}")
    if not stat.S_ISREG(metadata.st_mode):
        raise WorkaroundError(f"expected a regular file or absence: {path}")


def ensure_directory(path: Path, mode: int = 0o755) -> None:
    ancestry: list[Path] = []
    current = path
    while True:
        ancestry.append(current)
        if current.parent == current:
            break
        current = current.parent
    for directory in reversed(ancestry):
        try:
            metadata = directory.lstat()
        except FileNotFoundError:
            directory.mkdir(mode=mode)
            continue
        if stat.S_ISLNK(metadata.st_mode):
            raise WorkaroundError(f"refusing to traverse symlink: {directory}")
        if not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(f"expected a directory: {directory}")


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
    replace: bool,
) -> None:
    reject_unsafe_target(path)
    if not replace and path.exists():
        raise FileExistsError(f"refusing to overwrite existing file: {path}")
    ensure_directory(path.parent)
    descriptor, temporary_name = tempfile.mkstemp(
        prefix=f".{path.name}.", suffix=".tmp", dir=path.parent
    )
    temporary = Path(temporary_name)
    try:
        os.fchmod(descriptor, mode & 0o777)
        os.fchown(descriptor, uid, gid)
        with os.fdopen(descriptor, "wb") as stream:
            descriptor = -1
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        if replace:
            os.replace(temporary, path)
        else:
            os.link(temporary, path, follow_symlinks=False)
            temporary.unlink()
        fsync_directory(path.parent)
    except Exception:
        if descriptor >= 0:
            os.close(descriptor)
        temporary.unlink(missing_ok=True)
        raise


def parse_toml(path: Path) -> dict[str, Any]:
    require_regular(path)
    try:
        with path.open("rb") as stream:
            document = tomllib.load(stream)
    except (OSError, tomllib.TOMLDecodeError) as error:
        raise WorkaroundError(f"cannot parse TOML file {path}: {error}") from error
    if not isinstance(document, dict):
        raise WorkaroundError(f"TOML document is not a table: {path}")
    return document


def recursive_merge(base: dict[str, Any], later: dict[str, Any]) -> None:
    """Merge tables recursively; later arrays and scalar values replace."""
    for key, value in later.items():
        if isinstance(base.get(key), dict) and isinstance(value, dict):
            recursive_merge(base[key], value)
        else:
            base[key] = copy.deepcopy(value)


def toml_string(value: str) -> str:
    return json.dumps(value, ensure_ascii=False).replace("\x7f", "\\u007f")


def toml_key(value: str) -> str:
    if value and all(character in BARE_KEY for character in value):
        return value
    return toml_string(value)


def toml_value(value: Any) -> str:
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, int):
        return str(value)
    if isinstance(value, float):
        if not math.isfinite(value):
            sign = "-" if math.copysign(1.0, value) < 0 else "+"
            return sign + ("nan" if math.isnan(value) else "inf")
        return repr(value)
    if isinstance(value, str):
        return toml_string(value)
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


def array_of_tables(value: Any) -> bool:
    return isinstance(value, list) and bool(value) and all(
        isinstance(item, dict) for item in value
    )


def emit_table(
    lines: list[str],
    table: dict[str, Any],
    path: tuple[str, ...],
    *,
    header: bool,
) -> None:
    if header:
        if lines and lines[-1]:
            lines.append("")
        lines.append("[" + ".".join(toml_key(part) for part in path) + "]")
    for key, value in table.items():
        if not isinstance(value, dict) and not array_of_tables(value):
            lines.append(f"{toml_key(str(key))} = {toml_value(value)}")
    for key, value in table.items():
        child_path = (*path, str(key))
        if isinstance(value, dict):
            emit_table(lines, value, child_path, header=True)
        elif array_of_tables(value):
            for item in value:
                if lines and lines[-1]:
                    lines.append("")
                lines.append(
                    "[[" + ".".join(toml_key(part) for part in child_path) + "]]"
                )
                emit_table(lines, item, child_path, header=False)


def serialize_toml(document: dict[str, Any]) -> bytes:
    lines: list[str] = []
    emit_table(lines, document, (), header=False)
    content = ("\n".join(lines).rstrip() + "\n").encode()
    try:
        reparsed = tomllib.loads(content.decode())
    except tomllib.TOMLDecodeError as error:
        raise WorkaroundError(f"generated storage TOML is invalid: {error}") from error
    verified: list[str] = []
    emit_table(verified, reparsed, (), header=False)
    if ("\n".join(verified).rstrip() + "\n").encode() != content:
        raise WorkaroundError("generated storage TOML changed configuration semantics")
    return content


def storage_inputs(root: Path) -> list[Path]:
    admin_main = rooted(root, ADMIN_STORAGE)
    system_main = rooted(root, SYSTEM_STORAGE)
    main = admin_main if admin_main.exists() or admin_main.is_symlink() else system_main
    require_regular(main)

    selected: dict[str, Path] = {}
    # Low-to-high rootful precedence. Resolve duplicate basenames first, then
    # apply the selected files in one global lexicographic order.
    for directory in STORAGE_DROPINS:
        candidate_directory = rooted(root, directory)
        if not candidate_directory.exists():
            if candidate_directory.is_symlink():
                raise WorkaroundError(
                    f"unsafe storage drop-in directory: {candidate_directory}"
                )
            continue
        metadata = candidate_directory.lstat()
        if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(
                f"unsafe storage drop-in directory: {candidate_directory}"
            )
        for candidate in candidate_directory.glob("*.conf"):
            require_regular(candidate)
            selected[candidate.name] = candidate
    return [main, *(selected[name] for name in sorted(selected))]


def merged_storage(root: Path) -> tuple[list[Path], bytes]:
    inputs = storage_inputs(root)
    merged: dict[str, Any] = {}
    for path in inputs:
        recursive_merge(merged, parse_toml(path))
    return inputs, serialize_toml(merged)


def vendor_destinations(root: Path) -> dict[Path, tuple[Path, bytes, int]]:
    destinations: dict[Path, tuple[Path, bytes, int]] = {}
    for name in VENDOR_FILES:
        source = rooted(root, SYSTEM_CONTAINERS / name)
        if source.exists() or source.is_symlink():
            metadata = require_regular(source)
            destination = rooted(root, ADMIN_CONTAINERS / name)
            destinations[destination] = (
                source,
                source.read_bytes(),
                stat.S_IMODE(metadata.st_mode),
            )
    for directory_name, pattern in VENDOR_DIRECTORIES:
        source_directory = rooted(root, SYSTEM_CONTAINERS / directory_name)
        if not source_directory.exists():
            if source_directory.is_symlink():
                raise WorkaroundError(
                    f"unsafe vendor directory: {source_directory}"
                )
            continue
        metadata = source_directory.lstat()
        if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
            raise WorkaroundError(f"unsafe vendor directory: {source_directory}")
        for source in sorted(source_directory.glob(pattern)):
            file_metadata = require_regular(source)
            destination = rooted(
                root, ADMIN_CONTAINERS / directory_name / source.name
            )
            destinations[destination] = (
                source,
                source.read_bytes(),
                stat.S_IMODE(file_metadata.st_mode),
            )
    return destinations


def apply(root: Path) -> None:
    target = rooted(root, ADMIN_STORAGE)
    reject_unsafe_target(target)
    original_metadata = require_regular(target) if target.exists() else None

    # Validate and read every input before copying vendor data or touching
    # /etc/containers.
    inputs, storage_content = merged_storage(root)
    destinations = vendor_destinations(root)
    copies = []
    for destination, source_data in destinations.items():
        reject_unsafe_target(destination)
        if not destination.exists():
            copies.append((destination, *source_data))
    owner = 0 if root == Path("/") else os.geteuid()
    group = 0 if root == Path("/") else os.getegid()
    for destination, _source, content, mode in copies:
        atomic_write(
            destination,
            content,
            mode=mode,
            uid=owner,
            gid=group,
            replace=False,
        )
    output_metadata = original_metadata or require_regular(inputs[0])
    atomic_write(
        target,
        storage_content,
        mode=stat.S_IMODE(output_metadata.st_mode),
        uid=output_metadata.st_uid,
        gid=output_metadata.st_gid,
        replace=original_metadata is not None,
    )
    print(
        f"Applied temporary workaround using {len(inputs)} storage files and "
        f"{len(copies)} missing vendor files"
    )


def parse_args(arguments: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--root",
        type=Path,
        default=Path("/"),
        help="alternate root used only for isolated validation",
    )
    return parser.parse_args(arguments)


def main(arguments: list[str] | None = None) -> int:
    args = parse_args(arguments if arguments is not None else sys.argv[1:])
    root = args.root.absolute()
    metadata = root.lstat()
    if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
        raise WorkaroundError(f"unsafe filesystem root: {root}")
    apply(root)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, WorkaroundError) as error:
        print(f"error: {error}", file=sys.stderr)
        raise SystemExit(1) from error
