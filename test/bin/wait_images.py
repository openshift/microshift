#!/usr/bin/env python3
#
# This script is run by build_images.sh after all of the builds have
# been enqueued. It waits for the jobs identified by the UUIDs
# provided as input to either fail or complete.

import argparse
import json
import logging
import subprocess
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


def main(build_ids):
    ignore_ids = set()
    known_ids = set(build_ids)
    found_ids = set()
    while build_ids:
        logging.info(f'Waiting for {build_ids}')
        for job in flattened_status():
            job_id = job["id"]
            found_ids.add(job_id)
            status_text = f'{job_id} {job["compose_type"]} for {job["blueprint"]} - {job["queue_status"]}'
            if job_id in build_ids:
                logging.info(status_text)
                if job["queue_status"] in {"FAILED", "FINISHED"}:
                    # After a job fails or finishes, stop reporting its status.
                    build_ids.remove(job_id)
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
            build_ids.remove(i)
        if build_ids:
            time.sleep(30)


if __name__ == '__main__':
    cli_parser = argparse.ArgumentParser(add_help=True)
    cli_parser.add_argument(
        'build_id',
        nargs='+',
        help="a build id (UUID) from composer",
    )
    args = cli_parser.parse_args()
    main(args.build_id)
