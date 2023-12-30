#!/bin/bash
set -euo pipefail

BCKSUFFIX=".tmpfs.bck"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} <create <tmpfs_size_in_GB> | cleanup>

    -h   Show this help.

create:  1. Rename the '${HOME}' directory to '${HOME}${BCKSUFFIX}'.
         2. Create an in-memory tmpfs file system of the specified size at
            '${HOME}' and copy the backed up data.
         3. Rename '/var/tmp' and '/var/cache' directories to have '${BCKSUFFIX}'
         suffix and soft-link these directories to '${HOME}'.

cleanup: Undo the settings made by 'create' command.
         DATA LOSS WARNING: The tmpfs changes in '${HOME}', including the
         contents of 'tmp' and 'cache' directories, are lost.
EOF
}

is_tmpfs() {
    local -r fs_path="$1"
    local -r fs_type="$(df --output=fstype "${fs_path}" | tail -1)"

    if [ "${fs_type}" = "tmpfs" ] ; then
        return 0
    fi
    return 1
}

print_tmpfs_status() {
    echo "=================================="
    echo "Status of the updated file systems"
    echo "=================================="

    df -T "${HOME}/"
    ls -ld /var/tmp /var/cache
}

free_tmpfs_usage() {
    local -r tmp_root="$1"
    local -r tmp_pids=$(sudo lsof "${tmp_root}" | awk '{print $2}' | sort -un | grep -vw PID | xargs)

    if [ -n "${tmp_pids}" ] ; then
        echo "Killing processes that use '${tmp_root}': ${tmp_pids}"
        # shellcheck disable=SC2086
        sudo kill -9 ${tmp_pids}
    fi
}

action_create() {
    local -r tmp_size="$1"
    local -r ram_size="$(free -g | awk '/^Mem:/ {print $2}')"

    local -r tmp_root="${HOME}"
    local -r bck_root="${HOME}${BCKSUFFIX}"

    # Make sure the specified sizes make sense
    if [ "${tmp_size}" -lt 1 ] || [ "${tmp_size}" -ge "${ram_size}" ] ; then
        echo "Cannot allocate tmpfs of ${tmp_size}GB size using ${ram_size}GB RAM"
        exit 1
    fi

    # Backup the microshift directory if it already exists
    if [ -d "${tmp_root}" ] ; then
        if is_tmpfs "${tmp_root}" ; then
            echo "The '${tmp_root}' directory is already mounted on tmpfs"
            exit 1
        fi
        if [ -d "${bck_root}" ] ; then
            echo "The '${bck_root}' directory already exists"
        else
            sudo mv "${tmp_root}" "${bck_root}"
        fi
    fi

    # Create the tmpfs file system
    sudo mkdir -p "${tmp_root}"
    sudo mount -t tmpfs -o size="${tmp_size}G" tmpfs "${tmp_root}"
    sudo chown "$(whoami)." "${tmp_root}"

    # Copy the backup data into the newly create mount
    if [ -d "${bck_root}" ] ; then
        echo "Copying '${bck_root}' contents to '${tmp_root}'"
        sudo rsync -ar "${bck_root}/" "${tmp_root}/"
    fi

    # Relocate the cache directories
    for d in /var/cache /var/tmp ; do
        local base_dir
        base_dir=$(basename "${d}")
        sudo mkdir -p "${tmp_root}/${base_dir}"

        echo "Moving '${d}' to '${d}${BCKSUFFIX}'"
        sudo bash -c "mv \"${d}\" \"${d}${BCKSUFFIX}\" ; ln -s \"${tmp_root}/${base_dir}\" \"${d}\""
    done
    # Make sure the tmp directory has right permissions
    sudo chmod 1777 "${tmp_root}/tmp"

    print_tmpfs_status
}

action_cleanup() {
    local -r tmp_root="${HOME}"
    local -r bck_root="${HOME}${BCKSUFFIX}"

    if [ ! -d "${tmp_root}" ] ; then
        echo "The '${tmp_root}' directory does not exist"
        exit 1
    fi

    # Verify that the directory is mounted on tmpfs
    if ! is_tmpfs "${tmp_root}" ; then
        echo "The '${tmp_root}' directory is not mounted on tmpfs"
        exit 1
    fi

    # Restore the cache contents from backup if they exist
    for d in /var/cache /var/tmp ; do
        if [ ! -L "${d}" ] ; then
            echo "Skipping '${d}' because it is not a symbolic link"
            continue
        fi
        sudo rm -f "${d}"

        local bck_dir="${d}${BCKSUFFIX}"
        if [ ! -d "${bck_dir}" ] ; then
            echo "The '${bck_dir}' directory does not exists"
        else
            sudo mv "${bck_dir}" "${d}"
        fi
    done

    # Unmount the tmpfs parition and delete the top-level directory
    free_tmpfs_usage "${tmp_root}"
    sudo umount --force "${tmp_root}"
    sudo rmdir "${tmp_root}"

    # Restore the home directory from backup if it exists
    if [ ! -d "${bck_root}" ] ; then
        echo "The '${bck_root}' directory does not exists"
    else
        sudo mv "${bck_root}" "${tmp_root}"
    fi

    print_tmpfs_status
}

if [ $# -ne 1 ] && [ $# -ne 2 ]; then
    usage
    exit 1
fi

case "${1}" in
    create)
        [ $# -ne 2 ] && usage && exit 1
        action_create "$2"
        ;;
    cleanup)
        action_cleanup
        ;;
    -h)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
