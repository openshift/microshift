#!/usr/bin/env python3

import os
import sys
import subprocess
from typing import List

def should_skip(file):
    # Implement your skipping logic here
    return False

def record_junit(groupdir, containerfile, filetype, status):
    # Implement your recording logic here
    pass

def get_env_var(var_name: str):
    """Get an environment variable or exit if not set."""
    value = os.environ.get(var_name)
    if value is None:
        print(f"Error: {var_name} environment variable not set.")
        sys.exit(1)
    return value

def run_command(command: List[str], dry_run: bool):
    """Run the command or print the command line depending on the dry run argument"""
    if dry_run:
        print(f"[Dry Run] {" ".join(command)}")
    else:
        subprocess.run(command, check=True)
