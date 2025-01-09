#!/usr/bin/env python3

"""
This script verifies all the assets for auto-rebase.
It checks that all files in the assets directory are listed in the assets.yaml file, and vice versa.

File: presubmit.py
"""

import glob
import os
import sys
from functools import reduce

import yaml

# pylint: disable=R0801
try:
    from yaml import CLoader as Loader
except ImportError:
    from yaml import Loader

ASSETS_DIR = "assets/"
STAGING_DIR = "_output/staging/"
RECIPE_FILEPATHS = [
    "./scripts/auto-rebase/assets.yaml",
    "./scripts/auto-rebase/lvms_assets.yaml",
    "./scripts/auto-rebase/ossm_assets.yaml",
    "./scripts/auto-rebase/assets_rhoai.yaml",
]


def build_assets_filelist_from_asset_dir(asset_dir, prefix=""):
    """Recursively builds a list of assets filepaths from an asset directory."""
    dir_path = os.path.join(prefix, asset_dir['dir'])
    return ([os.path.join(dir_path, f['file']) for f in asset_dir.get('files', [])] +
            reduce(lambda x, y: x+y,
                   [build_assets_filelist_from_asset_dir(subdir, dir_path) for subdir in asset_dir.get('dirs', [])],
                   []))


def build_assets_filelist_from_recipe(recipe):
    """Builds a list of assets filepaths from a recipe file."""
    return reduce(lambda x, y: x+[y] if isinstance(y, str) else x+y,
                  [build_assets_filelist_from_asset_dir(asset) if 'dir' in asset else asset['file'] for asset in recipe['assets']],
                  [])


def check_assets_dir_against_instructions(recipe):
    """Reports errors if the contents of the assets directory do not match the instruction file.

    Returns boolean indicating whether an issue was found.
    """
    assets_filelist = set(build_assets_filelist_from_recipe(recipe))
    realfiles = {f.replace('assets/', '') for f in glob.glob('assets/**/*', recursive=True) if not os.path.isdir(f)}

    missing_in_recipe = realfiles - assets_filelist
    superfluous_in_recipe = assets_filelist - realfiles

    if missing_in_recipe:
        print("ERROR: Detected files in assets/ that are not present in assets.yaml:\n\t -", '\n\t - '.join(missing_in_recipe))

    if superfluous_in_recipe:
        print("ERROR: Found unnecessary files in assets.yaml that do not exist in assets/:\n\t -", '\n\t - '.join(superfluous_in_recipe))

    if missing_in_recipe or superfluous_in_recipe:
        print("\nFiles in assets.yaml:\n\t -", '\n\t - '.join(sorted(assets_filelist)))
        print("\nFiles in assets/:\n\t -", '\n\t - '.join(sorted(realfiles)))
        return True
    return False


def _check_for_redundant_instructions(path, instructions):
    next_path = path + [instructions.get('dir', '')]
    if 'dirs' in instructions:
        # Use a list comprehension instead of a generator expression
        # to ensure all of the sub-entries are checked.
        return any([_check_for_redundant_instructions(next_path, d)
                    for d in instructions['dirs']])
    # Evaluate the files at this level
    have_error = False
    filenames = {}
    for entry in instructions.get('files', []):
        if entry['file'] in filenames:
            existing_path, existing_entry = filenames[entry['file']]
            print("ERROR: found multiple instructions for {}".format(entry['file']))
            print("       {}:".format(' -> '.join(existing_path)))
            print("       {}".format(existing_entry))
            print("       AND")
            print("       {}:".format(' -> '.join(next_path)))
            print("       {}".format(entry))
            print("")
            have_error = True
        filenames[entry['file']] = (next_path, entry)
    return have_error


def check_for_redundant_instructions(recipe):
    """Look for assets that appear in the recipe file multiple times."""
    return any([
        _check_for_redundant_instructions([], asset)
        for asset in recipe['assets']
    ])


def main():
    """Main function for checking assets against an asset recipe."""
    if not os.path.isdir(ASSETS_DIR):
        print(f"ERROR: Expected to run in root directory of microshift repository but was in {os.getcwd()}")
        sys.exit(1)

    # Merge the data from all of the asset files into one set so we
    # can ensure that everything in the assets directory is mentioned
    # in at least one input file.
    recipe = {"assets": []}
    for filename in RECIPE_FILEPATHS:
        with open(filename, encoding='utf-8') as recipe_file:
            r = yaml.load(recipe_file.read(), Loader=Loader)
            recipe["assets"].extend(r["assets"])

    found_error = any([
        check_assets_dir_against_instructions(recipe),
        check_for_redundant_instructions(recipe),
    ])

    if found_error:
        print("\nFAILURE")
        sys.exit(1)

    print("SUCCESS")


if __name__ == "__main__":
    main()
