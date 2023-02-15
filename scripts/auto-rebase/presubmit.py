#!/usr/bin/env python3

import os
import sys
import glob
import yaml
from functools import reduce

try:
    from yaml import CLoader as Loader
except ImportError:
    from yaml import Loader

ASSETS_DIR = "assets/"
STAGING_DIR = "_output/staging/"
RECIPE_FILEPATH = "./scripts/auto-rebase/assets.yaml"


def build_assets_filelist_from_asset_dir(dir, prefix=""):
    dir_path = os.path.join(prefix, dir['dir'])
    return ([os.path.join(dir_path, f['file']) for f in dir.get('files', [])] +
            reduce(lambda x, y: x+y,
                   [build_assets_filelist_from_asset_dir(subdir, dir_path) for subdir in dir.get('dirs', [])],
                   []))


def build_assets_filelist_from_recipe(recipe):
    return reduce(lambda x, y: x+[y] if type(y) == str else x+y,
                  [build_assets_filelist_from_asset_dir(asset) if 'dir' in asset else asset['file'] for asset in recipe['assets']],
                  [])


def main():
    if not os.path.isdir(ASSETS_DIR):
        print(f"ERROR: Expected to run in root directory of microshift repository but was in {os.getcwd()}")
        sys.exit(1)

    recipe = yaml.load(open(RECIPE_FILEPATH).read(), Loader=Loader)

    assets_filelist = set(build_assets_filelist_from_recipe(recipe))
    realfiles = set([f.replace('assets/', '') for f in glob.glob('assets/**/*.*', recursive=True)])

    missing_in_recipe = realfiles - assets_filelist
    superfluous_in_recipe = assets_filelist - realfiles

    if missing_in_recipe:
        print("ERROR: Detected files in assets/ that are not present in assets.yaml:\n\t -", '\n\t - '.join(missing_in_recipe))

    if superfluous_in_recipe:
        print("ERROR: Found unnecessary files in assets.yaml that do not exist in assets/:\n\t -", '\n\t - '.join(superfluous_in_recipe))

    if missing_in_recipe or superfluous_in_recipe:
        print("\nFiles in assets.yaml:\n\t -", '\n\t - '.join(sorted(assets_filelist)))
        print("\nFiles in assets/:\n\t -", '\n\t - '.join(sorted(realfiles)))
        print("\nFAILURE")
        sys.exit(1)

    print("SUCCESS")


if __name__ == "__main__":
    main()
