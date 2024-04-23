#!/usr/bin/env python3

import os
import pathlib
import sys
import subprocess
from typing import List


PUSHD_DIR_STACK = []


def should_skip(file):
    # Implement your skipping logic here
    return False


def record_junit(groupdir, containerfile, filetype, status):
    # Implement your recording logic here
    pass


def print_msg(msg: str, file=sys.stderr):
    print(msg, file=file)


def get_env_var(var_name: str, def_val: str = None):
    """Get an environment variable or exit if not set."""
    value = os.environ.get(var_name)
    if value:
        return value
    if def_val:
        return def_val
    print_msg(f"Error: {var_name} environment variable not set.")
    sys.exit(1)


def run_command(command: List[str], dry_run: bool):
    """Run the command or print the command line depending on the dry run argument"""
    if dry_run:
        print_msg(f"[DRY RUN] {' '.join(command)}")
        return None

    print_msg(f"[RUN] {' '.join(command)}")
    return subprocess.run(command, check=True)


def run_command_in_shell(command: List[str], dry_run: bool = False,
                         stdout=subprocess.PIPE, stderr=sys.stderr):
    """Run the command through shell and return its standard output"""
    """If output file descriptors are specified, the appropriate output is redirected"""
    # Convert command to a string if necessary
    if isinstance(command, list):
        command = ' '.join(command)
    if dry_run:
        print_msg(f"[DRY RUN] {command}")
        return ""

    print_msg(f"[SHELL] {command}")
    # Run the command and return its output
    result = subprocess.run(
        command,
        check=True, shell=True, text=True,
        env=os.environ.copy(),
        stdout=stdout, stderr=stderr)
    return result.stdout.strip() if result.stdout else ""


def create_dir(dir: str):
    """Attempt recursive directory creation ignoring errors if it already exists"""
    path = pathlib.Path(dir)
    path.mkdir(parents=True, exist_ok=True)
    return path


def pushd(dir: str):
    """Change directory saving the current directory in the stack"""
    global PUSHD_DIR_STACK
    PUSHD_DIR_STACK.append(os.getcwd())
    os.chdir(dir)


def popd():
    """Restore the current directory from the stack"""
    if not PUSHD_DIR_STACK:
        raise Exception("Directory stack is empty")
    dir = PUSHD_DIR_STACK.pop()
    os.chdir(dir)


def delete_file(file_path: str):
    """Attempt file deletion ignoring errors when a file does not exist"""
    try:
        os.remove(file_path)
    except FileNotFoundError:
        pass


def basename(path: str):
    """Return a base name of the path"""
    return pathlib.Path(path).name
