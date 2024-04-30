#!/usr/bin/env python3
#
# This script is run by build_images.sh after all of the builds have
# been enqueued. It waits for the jobs identified by the UUIDs
# provided as input to either fail or complete.

import argparse
import json
import logging
import os
import shutil
import subprocess
import sys
import time

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s: %(message)s',
)

STATUS_COMMAND = ['sudo', 'composer-cli', 'compose', 'status', '--json']


# Convert the weird queue json data structure to a sequence of jobs.
# $ sudo composer-cli compose status --json
# [
#     {
#         "method": "GET",
#         "path": "/compose/queue",
#         "status": 200,
#         "body": {
#             "new": [
#                 {
#                     "blueprint": "rhel-9.2",
#                     "compose_type": "image-installer",
#                     "id": "7bdb78bf-c4b6-4f19-80df-073d38fade56",
#                     "image_size": 0,
#                     "job_created": 1687466891.2393456,
#                     "queue_status": "WAITING",
#                     "version": "0.0.1"
#                 }
#             ],
#             "run": [
#                 {
#                     "blueprint": "rhel-9.2",
#                     "compose_type": "edge-commit",
#                     "id": "4aa19f32-54e3-42ce-a4ba-cf038a3df91c",
#                     "image_size": 0,
#                     "job_created": 1687466889.5383687,
#                     "job_started": 1687466889.5480232,
#                     "queue_status": "RUNNING",
#                     "version": "0.0.1"
#                 }
#             ]
#         }
#     },
#     {
#         "method": "GET",
#         "path": "/compose/finished",
#         "status": 200,
#         "body": {
#             "finished": []
#         }
#     },
#     {
#         "method": "GET",
#         "path": "/compose/failed",
#         "status": 200,
#         "body": {
#             "failed": []
#         }
#     }
# ]
def flattened_status():
    result = subprocess.run(STATUS_COMMAND, stdout=subprocess.PIPE)
    if result.returncode != 0:
        raise SystemError(f'Status command returned {result.returncode}')
    status = json.loads(result.stdout)
    for result_set in status:
        for state in result_set["body"]:
            for job in result_set["body"][state]:
                yield job


def restart_job(cmd):
    result = subprocess.run(cmd, shell=True, text=True, stdout=subprocess.PIPE)
    if result.returncode != 0:
        return ""
    # Osbuild emits a warning on stdout about functionality we don't use in our blueprints,
    # so we need to skip the lines we don't find interesting. Example output:
    # > Warning: Please note that user customizations on "edge-commit" image type are deprecated and will be removed in the near future
    # >
    # > Compose 2c6a1ba2-4a18-49b8-a0f1-d103de1bd93a added to the queue
    # Split output with \n and select line starting with "Compose", then split that line with space and select 2nd word.
    line_with_id = next(line for line in result.stdout.split("\n") if line.startswith("Compose"))
    return line_with_id.split(" ")[1]


def copy_build_metadata(old_id, new_id):
    imagedir = os.environ['IMAGEDIR']
    shutil.copy(f'{imagedir}/builds/{old_id}.build', f'{imagedir}/builds/{new_id}.build')


def main(build_ids):
    ignore_ids = set()
    known_ids = set(build_ids.keys())
    found_ids = set()
    # IDs that the script will print out after waiting.
    # If any job failed, its ID will be replaced with ID of retry job
    finished_ids = set()

    while build_ids:
        logging.info(f'Waiting for {list(build_ids.keys())}')
        for job in flattened_status():
            job_id = job["id"]
            found_ids.add(job_id)
            status_text = f'{job_id} {job["compose_type"]} for {job["blueprint"]} - {job["queue_status"]}'
            if job_id in build_ids:
                logging.info(status_text)

                if job["queue_status"] == "FAILED":
                    cmd = build_ids[job_id]

                    # After a job fails, stop reporting its status.
                    del build_ids[job_id]

                    if cmd != "":
                        logging.info(f'Job {job_id} failed - restarting once ({cmd})')
                        new_id = restart_job(cmd)
                        if new_id == "":
                            # Failed to restart job, print it at the end so the caller can handle it.
                            finished_ids.add(job_id)
                            continue

                        logging.info(f'Job {job_id} restarted as {new_id}')
                        copy_build_metadata(job_id, new_id)
                        # Adding empty cmd means it won't be retried anymore - if fails again, it'll be final.
                        build_ids[new_id] = ""
                        known_ids.add(new_id)
                        found_ids.add(new_id)
                    else:
                        logging.error(f'Job {job_id} failed - not restarting anymore (empty cmd)')
                        finished_ids.add(job_id)

                elif job["queue_status"] == "FINISHED":
                    # After a job finishes, stop reporting its status.
                    del build_ids[job_id]
                    finished_ids.add(job_id)

            elif job_id not in ignore_ids and job_id not in known_ids:
                # Report any unknown jobs one time, then ignore them.
                logging.info(f'{status_text} (unknown job)')
                ignore_ids.add(job_id)

        to_ignore = []
        for build_id in build_ids:
            if build_id not in found_ids:
                logging.info(f'{build_id} is not a known build, ignoring')
                to_ignore.append(build_id)
        for i in to_ignore:
            del build_ids[i]

        if build_ids:
            time.sleep(30)

    # Print to stdout list of builds that caller script should handle.
    # It might be different from the list this script received initially, if any of the build had to be restarted.
    print(' '.join(finished_ids))


def pair(arg):
    return tuple(arg.split(','))


if __name__ == '__main__':
    cli_parser = argparse.ArgumentParser(add_help=True)
    cli_parser.add_argument(
        'id_cmd',
        type=pair,
        nargs='+',
        help="""a build id (UUID) from composer and a command to retry build in case of failure (separated with a comma), e.g.:
        'UUID,sudo composer-cli compose start-ostree --parent rhel-9.2 --url http://IP:8080/repo --ref rhel-9.2-microshift-source rhel-9.2-microshift-source edge-commit'""",
    )
    args = cli_parser.parse_args()

    # Make sure it exists as soon as possible, instead of waiting until it's needed.
    imagedir = os.getenv('IMAGEDIR')
    if imagedir is None:
        sys.exit('Script requires IMAGEDIR env var to be set')

    input = dict(args.id_cmd)
    logging.info(f'Received arguments: {input}')
    main(input)
