#!/usr/bin/env python3
"""
This script updates assets based on a YAML recipe (`assets.yaml`)
The recipe specifies what files and directories should be copied, ignored, and restored.

File: handle_assets.py
"""

import argparse
import logging
import os
import shutil
import subprocess
import sys

import yaml

try:
    from yaml import CLoader as Loader
except ImportError:
    from yaml import Loader

logging.basicConfig(level=logging.DEBUG, format='%(asctime)-27s %(levelname)-10s %(message)s')

ASSETS_DIR = "assets/"
STAGING_DIR = "_output/staging/"


def merge_paths(pathl, pathr):
    """
    Merge two paths depending upon following conditions:
    - If `pathr` is absolute (starts with `/`), then discard the leading `/` and return rest `pathr`.
    - If `pathr` is relative, then return `pathl/pathr`.
    """
    if pathr.startswith("/"):
        return pathr[1:]
    return os.path.join(pathl, pathr)


def run_command(args=None):
    """Run a command with the given args and return True if successful."""
    if args is None:
        args = []
    if not args:
        logging.error("run_command() received empty args")
        sys.exit(1)

    logging.debug(f"Executing '{' '.join(args)}'")
    result = subprocess.run(
        args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True, check=False)

    if result.returncode != 0:
        logging.error(f"Command '{' '.join(args)}' returned {result.returncode}. Output: {result.stdout}")
        sys.exit(1)

    return result.returncode == 0


def git_restore(path):
    """Restore a file from git."""
    path = os.path.join(ASSETS_DIR, path)
    logging.info(f"Restoring {path}")
    return run_command(["git", "restore", path])


def copy(src, dst):
    """Copy a file from the source path to the destination path."""
    src = os.path.join(STAGING_DIR, src)
    dst = os.path.join(ASSETS_DIR, dst)
    logging.debug(f"Copying {dst} <- {src}")
    shutil.copyfile(src, dst)


def clear_dir(path):
    """Clear the contents of a directory."""
    path = os.path.join(ASSETS_DIR, path)
    if os.path.isdir(path):
        logging.info(f"Clearing directory {path}")
        shutil.rmtree(path)
    os.makedirs(path)


def should_be_ignored(asset, dst):
    """Check if an asset should be ignored based on its 'ignore' field."""
    if 'ignore' in asset:
        reason = asset['ignore']
        if not reason:
            logging.error(f"{dst} is missing a reason why it's ignored")
            sys.exit(1)
        logging.warning(f"Ignoring {dst} because {reason}")
        return True
    return False


def handle_file(file, dst_dir="", src_prefix=""):
    """Handle a file by copying, restoring or ignoring it."""
    name = file['file']
    dst = merge_paths(dst_dir, name)

    if should_be_ignored(file, dst):
        return

    if 'git_restore' in file:
        git_restore(dst)
        return

    src = src_prefix
    if 'src' in file:
        src = merge_paths(src, file['src'])
    if os.path.extsep not in os.path.basename(src):
        src = merge_paths(src, name)
    copy(src, dst)


def handle_dir(dir_, dst_dir="", src_prefix=""):
    """"Recursively handle a directory, its files and subdirectories."""
    dst = merge_paths(dst_dir, dir_['dir'])
    new_src_prefix = merge_paths(src_prefix, dir_['src'] if "src" in dir_ else "")

    if should_be_ignored(dir_, dst):
        return

    if dir_.get('no_clean', False):
        logging.info(f"Not clearing dir {dst}")
    else:
        clear_dir(dst)

    for file in dir_.get('files', []):
        handle_file(file, dst, new_src_prefix)

    for sub_dir in dir_.get('dirs', []):
        handle_dir(sub_dir, dst, new_src_prefix)


def main():
    """Main function for handling assets based on the recipe file."""

    parser = argparse.ArgumentParser()
    parser.add_argument("asset_file", action="store")
    args = parser.parse_args()

    if not os.path.isdir(ASSETS_DIR):
        logging.error(f"Expected to run in root directory of microshift repository but was in {os.getcwd()}")
        sys.exit(1)

    if not os.path.isdir(STAGING_DIR):
        logging.error(f"{STAGING_DIR} does not exist")
        sys.exit(1)

    with open(args.asset_file, encoding='utf-8') as recipe_file:
        recipe = yaml.load(recipe_file.read(), Loader=Loader)

    for asset in recipe['assets']:
        if "dir" in asset:
            handle_dir(asset)

        if 'file' in asset:
            handle_file(asset)


if __name__ == "__main__":
    main()
