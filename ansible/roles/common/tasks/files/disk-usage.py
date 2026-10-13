#!/usr/bin/env python3
"""Capture allocated bytes and the existing MiB display from one du traversal."""

import glob
import json
import shlex
import subprocess
import sys
from datetime import datetime, timezone


def measure(request):
    paths = []
    for pattern in shlex.split(request):
        paths.extend(sorted(glob.glob(pattern)) or [pattern])
    result = subprocess.run(
        ["du", "-B1", "-t", "1M", "--max-depth=1", "--null", "--", *paths],
        capture_output=True, check=False,
    )
    entries = []
    text = []
    for row in result.stdout.split(b"\0"):
        if not row:
            continue
        size, path = row.split(b"\t", 1)
        size = int(size)
        path = path.decode(errors="surrogateescape")
        entries.append({"path": path, "allocated_bytes": size})
        text.append(f"{(size + 1048575) // 1048576}M\t{path}")
    return {
        "rc": result.returncode,
        "stdout": "\n".join(text),
        "stderr_lines": result.stderr.decode(errors="replace").splitlines(),
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "request": request,
        "paths": entries,
    }


if __name__ == "__main__":
    print(json.dumps(measure(sys.argv[1])))
