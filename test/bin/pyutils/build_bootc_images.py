#!/usr/bin/env python3

import argparse
import concurrent.futures
import glob
import os
import platform
import re
import shutil
import sys
import traceback

import common

# Global environment variables
#
# Note: Global variables for RPM versions and repos are
# initialized in set_rpm_version_info_vars function
SCRIPTDIR = common.get_env_var('SCRIPTDIR')
BOOTC_IMAGE_DIR = common.get_env_var('BOOTC_IMAGE_DIR')
IMAGEDIR = common.get_env_var('IMAGEDIR')
UNAME_M = common.get_env_var('UNAME_M')
CONTAINER_LIST = common.get_env_var('CONTAINER_LIST')
LOCAL_REPO = common.get_env_var('LOCAL_REPO')
BASE_REPO = common.get_env_var('BASE_REPO')
NEXT_REPO = common.get_env_var('NEXT_REPO')
HOME_DIR = common.get_env_var("HOME")
PULL_SECRET = common.get_env_var('PULL_SECRET', f"{HOME_DIR}/.pull-secret.json")


def find_latest_rpm(repo_path, version=""):
    rpms = glob.glob(f"{repo_path}/**/microshift-release-info-{version}*.rpm", recursive=True)
    if not rpms:
        raise Exception(f"Failed to find 'microshift-release-info-{version}*' RPM in {repo_path}")
    rpms.sort()
    return rpms[-1]


def is_rhocp_available(ver):
    # Equivalent to `uname -m`
    architecture = platform.machine()
    repository = f"rhocp-4.{ver}-for-rhel-9-{architecture}-rpms"

    try:
        # Run the dnf command to check for cri-o in the specified repository
        repo_info = common.run_command_in_shell(f"sudo dnf repository-packages {repository} info cri-o")
        common.print_msg(repo_info)
        return True
    except Exception:
        return False


def set_rpm_version_info_vars():
    global SOURCE_VERSION
    global MINOR_VERSION
    global PREVIOUS_MINOR_VERSION
    global YMINUS2_MINOR_VERSION
    global FAKE_NEXT_MINOR_VERSION
    global SOURCE_VERSION_BASE
    global CURRENT_RELEASE_VERSION
    global CURRENT_RELEASE_REPO
    global PREVIOUS_RELEASE_VERSION
    global PREVIOUS_RELEASE_REPO
    global RHOCP_MINOR_Y
    global RHOCP_MINOR_Y1
    global RHOCP_MINOR_Y2
    global YMINUS2_RELEASE_VERSION
    global YMINUS2_RELEASE_REPO

    release_info_rpm = find_latest_rpm(LOCAL_REPO)
    release_info_rpm_base = find_latest_rpm(BASE_REPO)

    SOURCE_VERSION = common.run_command_in_shell(f"rpm -q --queryformat '%{{version}}' {release_info_rpm}")
    MINOR_VERSION = SOURCE_VERSION.split('.')[1]
    PREVIOUS_MINOR_VERSION = str(int(MINOR_VERSION) - 1)
    YMINUS2_MINOR_VERSION = str(int(MINOR_VERSION) - 2)
    FAKE_NEXT_MINOR_VERSION = str(int(MINOR_VERSION) + 1)
    SOURCE_VERSION_BASE = common.run_command_in_shell(f"rpm -q --queryformat '%{{version}}' {release_info_rpm_base}")

    current_version_repo = common.run_command_in_shell(f"source {SCRIPTDIR}/get_rel_version_repo.sh; get_rel_version_repo {MINOR_VERSION}")
    CURRENT_RELEASE_VERSION, CURRENT_RELEASE_REPO = current_version_repo.split(',')

    previous_version_repo = common.run_command_in_shell(f"source {SCRIPTDIR}/get_rel_version_repo.sh; get_rel_version_repo {PREVIOUS_MINOR_VERSION}")
    PREVIOUS_RELEASE_VERSION, PREVIOUS_RELEASE_REPO = previous_version_repo.split(',')

    if is_rhocp_available(MINOR_VERSION):
        RHOCP_MINOR_Y = MINOR_VERSION
    if is_rhocp_available(PREVIOUS_MINOR_VERSION):
        RHOCP_MINOR_Y1 = PREVIOUS_MINOR_VERSION

    # For Y-2, there will always be a real repository, so we can always
    # set the template variable for enabling that package source and use
    # the well-known name of that repo instead of figuring out the URL.
    yminus2_version_repo = common.run_command_in_shell(f"source {SCRIPTDIR}/get_rel_version_repo.sh; get_rel_version_repo {YMINUS2_MINOR_VERSION}")
    YMINUS2_RELEASE_VERSION, _ = yminus2_version_repo.split(',')
    YMINUS2_RELEASE_REPO = common.run_command_in_shell(f"source {SCRIPTDIR}/get_rel_version_repo.sh; get_ocp_repo_name_for_version {YMINUS2_MINOR_VERSION}")
    RHOCP_MINOR_Y2 = YMINUS2_MINOR_VERSION


def get_container_images(path, version):
    # Find the last microshift-release-info RPM with the specified version
    release_info_rpm = find_latest_rpm(path, version)
    # Extract list of image URIs and join them with a comma
    cpio_cmd = f"rpm2cpio '{release_info_rpm}' | cpio -i --to-stdout '*release-{UNAME_M}.json' 2> /dev/null"
    jq_cmd = "jq -r '[.images[]] | join(\",\")'"
    return common.run_command_in_shell(f"{cpio_cmd} | {jq_cmd}")


def extract_container_images(version, repo_spec, outfile, dry_run=False):
    common.print_msg(f"Extracting images from {version}")
    # Create and change the directory for extracting RPMs
    image_path = common.create_dir(f"{IMAGEDIR}/release-info-rpms")
    common.pushd(str(image_path))

    repo_name = common.basename(repo_spec)
    dnf_options = []

    if re.match(r'^https://.*', repo_spec):
        # If the spec is a URL, set up the arguments to point to that location.
        dnf_options.extend(["--repofrompath", f"{repo_name},{repo_spec}", "--repo", repo_name])
    elif re.match(r'^/.*', repo_spec):
        # If the spec is a path, set up the arguments to point to that path.
        dnf_options.extend(["--repofrompath", f"{repo_name},{repo_spec}", "--repo", repo_name])
    elif repo_spec:
        # If the spec is a name, assume it is already known to the
        # system through normal configuration. The repo does not need
        # to be enabled in order for dnf to download a package from it.
        dnf_options.extend(["--repo", repo_spec])

    # Construct and execute the dnf download command
    dnf_command = ["sudo", "dnf", "download"] + dnf_options + [f"microshift-release-info-{version}"]
    if common.run_command(dnf_command, dry_run) is not None:
        images_output = get_container_images(str(image_path), version)
        with open(outfile, "a") as f:
            f.write(images_output.replace(',', '\n'))
            f.write('\n')

        # Cleanup RPM files
        rpm_list = list(map(str, image_path.glob("microshift-release-info-*.rpm")))
        common.run_command(["sudo", "rm", "-f"] + rpm_list, dry_run)
    # Restore the current directory
    common.popd()


def process_containerfile(groupdir, containerfile, dry_run):
    cf_path = os.path.join(groupdir, containerfile)
    cf_outname = os.path.splitext(containerfile)[0]
    cf_outdir = os.path.join(BOOTC_IMAGE_DIR, cf_outname)
    cf_logfile = os.path.join(BOOTC_IMAGE_DIR, f"{cf_outname}.log")

    os.makedirs(BOOTC_IMAGE_DIR, exist_ok=True)

    if os.path.exists(cf_outdir):
        common.print_msg(f"{cf_outdir} already exists")
        if common.should_skip(cf_outname):
            common.record_junit(groupdir, cf_path, "containerfile", "SKIPPED")
            return

    common.print_msg(f"Processing {containerfile} with logs in {cf_logfile}")
    # Redirect the output to the log file
    with open(cf_logfile, 'w') as logfile:
        # Run the container build command
        build_args = [
            "podman", "build",
            "--authfile", PULL_SECRET,
            "-t", cf_outname, "-f", cf_path,
            os.path.join(IMAGEDIR, "rpm-repos")
        ]
        common.run_command_in_shell(build_args, dry_run, logfile, logfile)

        # Run the container export command
        if os.path.exists(cf_outdir):
            shutil.rmtree(cf_outdir)
        save_args = [
            "podman", "save",
            "--format", "oci-dir",
            "-o", cf_outdir, cf_outname
        ]
        common.run_command_in_shell(save_args, dry_run, logfile, logfile)


def process_image_bootc(groupdir, bootcfile, dry_run):
    None


def process_group(groupdir, dry_run=False):
    futures = []
    # Parallel processing loop
    with concurrent.futures.ProcessPoolExecutor() as executor:
        # Scan group directory contents sorted by length and then alphabetically
        for file in sorted(os.listdir(groupdir), key=lambda i: (len(i), i)):
            if file.endswith(".containerfile"):
                futures += [executor.submit(process_containerfile, groupdir, file, dry_run)]
            elif file.endswith(".image-bootc"):
                futures += [executor.submit(process_image_bootc, groupdir, file, dry_run)]
            else:
                common.print_msg(f"Skipping unknown file {file}")

    try:
        # Wait for the parallel tasks to complete
        for f in concurrent.futures.as_completed(futures):
            common.print_msg(f"Task {f} completed")
            # Result function generates an exception depending on the task state
            f.result()
    except Exception:
        # Cancel all pending tasks and propagate the exception
        for f in futures:
            if not f.done():
                f.cancel()
                common.print_msg(f"Task {f} cancelled")
        raise


def main():
    # Parse command line arguments
    parser = argparse.ArgumentParser(description="Process container files with Podman.")
    parser.add_argument("-d", "--dry-run", action="store_true", help="Dry run: skip executing Podman commands.")
    dirgroup = parser.add_mutually_exclusive_group(required=True)
    dirgroup.add_argument("-l", "--layer-dir", type=str, help="Path to the layer directory to process.")
    dirgroup.add_argument("-g", "--group-dir", type=str, help="Path to the group directory to process.")

    args = parser.parse_args()
    # Convert input directories to absolute paths
    if args.group_dir:
        args.group_dir = os.path.abspath(args.group_dir)
    if args.layer_dir:
        args.layer_dir = os.path.abspath(args.layer_dir)

    try:
        # Make sure the local RPM repository exists
        if not os.path.isdir(LOCAL_REPO):
            raise Exception("Run create_local_repo.sh before building images")

        # Determine versions of RPM packages
        set_rpm_version_info_vars()
        # Prepare container lists for mirroring registries
        common.delete_file(CONTAINER_LIST)
        extract_container_images(SOURCE_VERSION, LOCAL_REPO, CONTAINER_LIST, args.dry_run)
        # The following images are specific to layers that use fake rpms built from source
        extract_container_images(f"4.{FAKE_NEXT_MINOR_VERSION}.*", NEXT_REPO, CONTAINER_LIST, args.dry_run)
        extract_container_images(PREVIOUS_RELEASE_VERSION, PREVIOUS_RELEASE_REPO, CONTAINER_LIST, args.dry_run)
        extract_container_images(YMINUS2_RELEASE_VERSION, YMINUS2_RELEASE_REPO, CONTAINER_LIST, args.dry_run)

        # Process individual group directory
        if args.group_dir:
            process_group(args.group_dir, args.dry_run)
        else:
            # Process layer directory contents sorted by length and then alphabetically
            for item in sorted(os.listdir(args.layer_dir), key=lambda i: (len(i), i)):
                item_path = os.path.join(args.layer_dir, item)
                # Check if this item is a directory
                if os.path.isdir(item_path):
                    process_group(item_path, args.dry_run)
        # Success message
        common.print_msg("Build complete")
    except Exception as e:
        common.print_msg(f"An error occurred: {e}")
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
