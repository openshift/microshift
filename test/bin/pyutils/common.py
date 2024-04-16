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


def get_env_var(var_name: str):
    """Get an environment variable or exit if not set."""
    value = os.environ.get(var_name)
    if value is None:
        print_msg(f"Error: {var_name} environment variable not set.")
        sys.exit(1)
    return value


def run_command(command: List[str], dry_run: bool):
    """Run the command or print the command line depending on the dry run argument"""
    if dry_run:
        print_msg(f"[DRY RUN] {' '.join(command)}")
        return None

    print_msg(f"[RUN] {' '.join(command)}")
    return subprocess.run(command, check=True)


def run_command_in_shell(command: str):
    """Run the command through shell and return its output"""
    print_msg(f"[SHELL] {command}")
    result = subprocess.run(
        command,
        check=True, shell=True, text=True,
        env=os.environ.copy(), stdout=subprocess.PIPE)
    return result.stdout.strip()


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


def read_file(file_path: str):
    """Read the file contents and return them to the caller"""
    with open(file_path, 'r') as file:
        content = file.read()
    return content


def basename(path: str):
    """Return a base name of the path"""
    return pathlib.Path(path).name
