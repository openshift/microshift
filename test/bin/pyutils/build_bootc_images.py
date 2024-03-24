#!/usr/bin/env python3

import argparse
import os
import shutil
import sys
import traceback

import common

BOOTC_IMAGE_DIR = common.get_env_var('BOOTC_IMAGE_DIR')
IMAGEDIR = common.get_env_var('IMAGEDIR')

def process_containerfiles(groupdir, dry_run=False):
    for containerfile in os.listdir(groupdir):
        if not containerfile.endswith(".containerfile"):
            continue

        cf_path = os.path.join(groupdir, containerfile)
        cf_outname = os.path.splitext(os.path.basename(containerfile))[0]
        cf_outdir = os.path.join(BOOTC_IMAGE_DIR, cf_outname)

        os.makedirs(BOOTC_IMAGE_DIR, exist_ok=True)

        if os.path.exists(cf_outdir):
            print(f"{cf_outdir} already exists")
            if common.should_skip(cf_outname):
                common.record_junit(groupdir, cf_path, "containerfile", "SKIPPED")
                continue

        print(f"Processing {cf_path}")
        common.run_command(
            ["podman", "build", "-t", cf_outname, "-f", cf_path,
            os.path.join(IMAGEDIR, "rpm-repos")], dry_run)
        if os.path.exists(cf_outdir):
            shutil.rmtree(cf_outdir)
        common.run_command(["podman", "save", "--format", "oci-dir", "-o", cf_outdir, cf_outname], dry_run)

def main():
    # Parse command line arguments
    parser = argparse.ArgumentParser(description="Process container files with Podman.")
    parser.add_argument("-d", "--dry-run", action="store_true", help="Dry run: skip executing Podman commands.")
    dirgroup = parser.add_mutually_exclusive_group(required=True)
    dirgroup.add_argument("-l", "--layer-dir", type=str, help="Path to the layer directory to process.")
    dirgroup.add_argument("-g", "--group-dir", type=str, help="Path to the group directory to process.")

    args = parser.parse_args()

    # Run the directory processing
    try:
        # Process individual group directory
        if args.group_dir:
            process_containerfiles(args.group_dir, args.dry_run)
            return
        # Process a layer with group directories
        for item in os.listdir(args.layer_dir):
            item_path = os.path.join(args.layer_dir, item)       
            # Check if this item is a directory
            if os.path.isdir(item_path):
                process_containerfiles(item_path, args.dry_run)
    except Exception as e:
        print(f"An error occurred: {e}")
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    main()
