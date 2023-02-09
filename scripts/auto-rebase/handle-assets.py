#!/usr/bin/env python3

import os
import sys
import yaml
import shutil
import logging
import subprocess

try:
    from yaml import CLoader as Loader, CDumper as Dumper
except ImportError:
    from yaml import Loader, Dumper

logging.basicConfig(level=logging.DEBUG, format='%(asctime)-27s %(levelname)-10s %(message)s')

ASSETS_DIR = "assets/"
STAGING_DIR = "_output/staging/"


def merge_paths(pathl, pathr):
    if pathr.startswith("/"):
        return pathr[1:]
    return os.path.join(pathl, pathr)


def run_command(args=[]):
    if not args:
        logging.error("run_command() received empty args")
        sys.exit(1)

    logging.debug(f"Executing '{' '.join(args)}'")
    result = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)

    if result.returncode != 0:
        logging.error(f"Command '{' '.join(args)}' returned {result.returncode}. Output: {result.stdout}")
        sys.exit(1)

    return result.returncode == 0


def git_restore(path):
    path = os.path.join(ASSETS_DIR, path)
    logging.info(f"Restoring {path}")
    return run_command(["git", "restore", path])


def copy(src, dst):
    src = os.path.join(STAGING_DIR, src)
    dst = os.path.join(ASSETS_DIR, dst)
    logging.debug(f"Copying {dst} <- {src}")
    shutil.copyfile(src, dst)


def clear_dir(path):
    path = os.path.join(ASSETS_DIR, path)
    logging.info(f"Clearing directory {path}")
    shutil.rmtree(path)
    os.makedirs(path)


def should_be_ignored(asset, dst):
    if 'ignore' in asset:
        reason = asset['ignore']
        if not reason:
            logging.error(f"{dst} is missing a reason why it's ignored")
            sys.exit(1)
        logging.warning(f"Ignoring {dst} because {reason}")
        return True
    return False


def handle_file(file, dst_dir="", src_prefix=""):
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


def handle_dir(dir, dst_dir="", src_prefix=""):
    dst = merge_paths(dst_dir, dir['dir'])
    new_src_prefix = merge_paths(src_prefix, dir['src'] if "src" in dir else "")

    if should_be_ignored(dir, dst):
        return

    if dir.get('no_clean', False):
        logging.info(f"Not clearing dir {dst}")
    else:
        clear_dir(dst)

    for f in dir.get('files', []):
        handle_file(f, dst, new_src_prefix)

    for d in dir.get('dirs', []):
        handle_dir(d, dst, new_src_prefix)


def main():
    if not os.path.isdir(ASSETS_DIR):
        logging.error(f"Expected to run in root directory of microshift repository but was in {os.getcwd()}")
        sys.exit(1)

    if not os.path.isdir(STAGING_DIR):
        logging.error(f"{STAGING_DIR} does not exist")
        sys.exit(1)

    recipe_filepath = "./scripts/auto-rebase/assets.yaml"
    recipe = yaml.load(open(recipe_filepath).read(), Loader=Loader)
    for asset in recipe['assets']:
        if "dir" in asset:
            handle_dir(asset)

        if 'file' in asset:
            handle_file(asset)


if __name__ == "__main__":
    main()
