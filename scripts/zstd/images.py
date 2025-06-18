#!/usr/bin/env python3

import argparse
import json
import os
import subprocess
import sys
import re
import urllib.request
import urllib.error

def run_command(command, description, stream_output=False):
    try:
        print(f"  -> Executing: {description}...")
        print(f"     Command: {' '.join(command)}")
        if stream_output:
            popen_kwargs = {
                "stdout": None,
                "stderr": None
            }
        else:
            popen_kwargs = {
                "stdout": subprocess.PIPE,
                "stderr": subprocess.PIPE,
                "text": True,
                "encoding": 'utf-8'
            }
        process = subprocess.Popen(command, **popen_kwargs)
        stdout_data, stderr_data = process.communicate()
        if process.returncode != 0:
            raise subprocess.CalledProcessError(
                returncode=process.returncode,
                cmd=command,
                output=stdout_data,
                stderr=stderr_data
            )
        print(f"  ✅ Success: {description}.")
        return True

    except subprocess.CalledProcessError as e:
        capture = not stream_output
        print(f"  ❌ Error: Failed to {description.lower()}.", file=sys.stderr)
        print(f"     Exit Code: {e.returncode}", file=sys.stderr)

        if capture:
            if e.output:
                print(f"     STDOUT:\n{e.output}", file=sys.stderr)
            if e.stderr:
                print(f"     STDERR:\n{e.stderr}", file=sys.stderr)
        return False

    except FileNotFoundError:
        print(f"  ❌ Error: '{command[0]}' command not found. Please ensure it's installed and in your PATH.", file=sys.stderr)
        sys.exit(1)

def get_image_data(source):
    if source.lower().startswith(('http://', 'https://')):
        print(f"-> Attempting to fetch JSON from URL: {source}")
        try:
            with urllib.request.urlopen(source) as response:
                if response.status != 200:
                    print(f"Error: Failed to fetch from URL. Status code: {response.status}", file=sys.stderr)
                    sys.exit(1)
                response_text = response.read().decode('utf-8')
                print("  ✅ Success: JSON data fetched from URL.")
                return json.loads(response_text)
        except urllib.error.URLError as e:
            print(f"Error: Failed to open URL '{source}'. Reason: {e.reason}", file=sys.stderr)
            sys.exit(1)
        except json.JSONDecodeError:
            print(f"Error: Failed to decode JSON from URL '{source}'. Please check the content.", file=sys.stderr)
            sys.exit(1)
    else:
        print(f"-> Attempting to read JSON from local file: {source}")
        try:
            with open(source, 'r') as f:
                print("  ✅ Success: JSON data read from local file.")
                return json.load(f)
        except FileNotFoundError:
            print(f"Error: The file '{source}' was not found.", file=sys.stderr)
            sys.exit(1)
        except json.JSONDecodeError:
            print(f"Error: Failed to decode JSON from '{source}'. Please check the file format.", file=sys.stderr)
            sys.exit(1)

def get_podman_storage_size():
    path = f"{os.environ['HOME']}/.local/share/containers/storage"
    try:
        process = subprocess.Popen(
            ["du", "-ks", path],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            encoding='utf-8'
        )
        stdout_data, _ = process.communicate()
        size_kb = int(stdout_data.split()[0])
        return size_kb / 1024  # Convert KB to MB
    except (ValueError, IndexError) as e:
        print(f"  ⚠️ Warning: Could not parse size of path '{path}'. Error: {e}", file=sys.stdout)
        return None

def process_images(source, target, base_repo):
    data_from = get_image_data(source).get("images", {})
    data_to = get_image_data(target).get("images", {})

    intersection = set(data_from.keys()) & set(data_to.keys())

    if not intersection:
        print("Warning: No common images found in the 'images' section of both files.")
        return

    print(f"\n🚀 Starting image processing")
    print("-" * 50)

    for name in intersection:
        print(f"\nProcessing image '{name}'")
        source_image = data_from[name]
        target_image = data_to[name]

        image_parts = source_image.split("/")
        chunked_source_image = f"{base_repo}/{re.sub(r'@sha256', '', image_parts[-1])}"
        image_parts = target_image.split("/")
        chunked_target_image = f"{base_repo}/{re.sub(r'@sha256', '', image_parts[-1])}"

        # Step 1: Pull, re-tag, push source image
        run_command(["podman", "rmi", "-f", source_image], f"Clean up source image")
        run_command(["podman", "rmi", "-f", chunked_source_image], f"Clean up chunked source image")
        result = run_command(["podman", "pull", source_image], f"Pull source image {source_image}")
        if not result:
            print(f"  ❌ Failed to pull source image. Continuing with next image")
            continue
        result = run_command(["podman", "tag", source_image, chunked_source_image], f"Tag source image as {chunked_source_image}")
        if not result:
            print(f"  ❌ Failed to tag source image. Continuing with next image")
            continue
        result = run_command(["podman", "push", "--compression-format", "zstd:chunked", chunked_source_image], f"Push chunked source image {chunked_source_image}")
        if not result:
            print(f"  ❌ Failed to push chunked source image. Continuing with next image")
            continue

        # Step 2: Pull, re-tag, push, clean target image.
        run_command(["podman", "rmi", "-f", target_image], f"Clean up target image")
        run_command(["podman", "rmi", "-f", chunked_target_image], f"Clean up chunked target image")
        prev_size = get_podman_storage_size()
        result = run_command(["podman", "pull", target_image], f"Pull target image {target_image}")
        if not result:
            print(f"  ❌ Failed to pull target image. Continuing with next image")
            continue
        new_size = get_podman_storage_size()
        print(f"  💾 Successfully pulled target image. Added size: {new_size-prev_size:.2f} MB")
        result = run_command(["podman", "tag", target_image, chunked_target_image], f"Tag target image as {chunked_target_image}")
        if not result:
            print(f"  ❌ Failed to tag target image. Continuing with next image")
            continue
        result = run_command(["podman", "push", "--compression-format", "zstd:chunked", chunked_target_image], f"Push chunked target image {chunked_target_image}")
        if not result:
            print(f"  ❌ Failed to push chunked target image. Continuing with next image")
            continue
        run_command(["podman", "rmi", "-f", source_image], f"Clean up source image")
        run_command(["podman", "rmi", "-f", chunked_source_image], f"Clean up chunked source image")
        run_command(["podman", "rmi", "-f", target_image], f"Clean up target image")
        run_command(["podman", "rmi", "-f", chunked_target_image], f"Clean up chunked target image")

        # Step 3: Pull chunked source and target images.
        result = run_command(["podman", "pull", chunked_source_image], f"Pull chunked source image")
        if not result:
            print(f"  ❌ Failed to pull chunked source image. Continuing with next image")
            continue
        prev_size = get_podman_storage_size()
        result = run_command(["podman", "pull", chunked_target_image], f"Pull chunked target image", stream_output=True)
        if not result:
            print(f"  ❌ Failed to pull chunked target image. Continuing with next image")
            continue
        new_size = get_podman_storage_size()
        print(f"  💾 Successfully pulled chunked target image. Added size: {new_size-prev_size:.2f} MB")

    print("=" * 50)

def main():
    """
    Parses command-line arguments and initiates the image processing.
    """
    parser = argparse.ArgumentParser(
        description="""A script to simulate an upgrade using zstd:chunked compressed pulls using Podman.
        The script takes two JSON files as input, one representing the source and the other the target images.
        It pulls all images, re-tagging them to push them to another repo using zstd:chunked compression.
        Afterwards, each image is pulled again to verify the reduced network usage. Also shows storage size changes.""",
        formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument(
        "source",
        help="The file containing the base for the upgrade simulation (can be a URL)."
    )
    parser.add_argument(
        "target",
        help="The file containing the target for the upgrade simulation (can be a URL)."
    )
    parser.add_argument(
        "repo",
        help="The base repository to use for the images (e.g., quay.io/username)."
    )
    args = parser.parse_args()
    process_images(args.source, args.target, args.repo)

if __name__ == "__main__":
    main()
