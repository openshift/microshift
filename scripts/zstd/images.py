#!/usr/bin/env python3

import argparse
import json
import os
import subprocess
import sys
import re
import urllib.request
import urllib.error
import time
import threading
import signal
from collections import defaultdict
from typing import Dict, Any, Optional
import psutil

class LiveMonitor:
    def __init__(self, command_name: str, interval: float = 1.0):
        self.command_name = command_name
        self.interval = interval
        self.monitoring = False
        self.start_time = None
        self.metrics = defaultdict(list)
        self.process_pid = None
        self.monitor_thread = None
        self.prev_process_io = None

    def get_system_metrics(self):
        try:
            # For some reason the process cpu is not reported correctly. Use the system value instead as we are interested in deltas.
            cpu_percent = psutil.cpu_percent(interval=0)
            current_time = time.time()

            metrics = {
                "timestamp": current_time,
                "available": True
            }

            if self.process_pid:
                try:
                    process = psutil.Process(self.process_pid)
                    memory_info = process.memory_info()
                    process_metrics = {
                        "cpu_percent": cpu_percent,
                        "memory_rss_mb": memory_info.rss / (1024 * 1024),
                        "memory_vms_mb": memory_info.vms / (1024 * 1024),
                        "num_threads": process.num_threads(),
                        "status": process.status()
                    }

                    try:
                        process_io = process.io_counters()
                        if process_io:
                            process_io_metrics = {
                                "read_count": process_io.read_count,
                                "write_count": process_io.write_count,
                                "read_bytes": process_io.read_bytes,
                                "write_bytes": process_io.write_bytes,
                                "read_iops": 0,
                                "write_iops": 0,
                                "total_iops": 0
                            }

                            if self.prev_process_io:
                                time_diff = current_time - self.prev_process_io["timestamp"]
                                if time_diff > 0:
                                    read_iops = (process_io.read_count - self.prev_process_io["read_count"]) / time_diff
                                    write_iops = (process_io.write_count - self.prev_process_io["write_count"]) / time_diff
                                    process_io_metrics["read_iops"] = read_iops
                                    process_io_metrics["write_iops"] = write_iops
                                    process_io_metrics["total_iops"] = read_iops + write_iops


                            process_metrics["disk"] = process_io_metrics

                            self.prev_process_io = {
                                "timestamp": current_time,
                                "read_count": process_io.read_count,
                                "write_count": process_io.write_count
                            }
                    except (AttributeError, OSError):
                        process_metrics["disk"] = {"available": False}

                    metrics["process"] = process_metrics
                except (psutil.NoSuchProcess, psutil.AccessDenied):
                    metrics["process"] = {"available": False}

            return metrics

        except Exception as e:
            return {"timestamp": time.time(), "available": False, "error": str(e)}

    def _monitor_loop(self):
        while self.monitoring:
            metrics = self.get_system_metrics()
            self.metrics[time.time() - self.start_time].append(metrics)
            time.sleep(self.interval)

    def start_monitoring(self, process_pid: Optional[int] = None):
        self.process_pid = process_pid
        self.start_time = time.time()
        self.monitoring = True
        self.storage_size = get_podman_storage_size()
        self.monitor_thread = threading.Thread(target=self._monitor_loop, daemon=True)
        self.monitor_thread.start()

    def stop_monitoring(self):
        self.monitoring = False
        if self.monitor_thread:
            self.monitor_thread.join(timeout=1.0)
        new_size = get_podman_storage_size()

        total_time = time.time() - self.start_time if self.start_time else 0

        memory_avg = 0
        memory_peak = 0
        cpu_avg = 0
        cpu_peak = 0
        disk_read_iops_avg = 0
        disk_write_iops_avg = 0
        disk_total_iops_avg = 0
        disk_read_iops_peak = 0
        disk_write_iops_peak = 0
        disk_total_iops_peak = 0

        if self.metrics:
            all_metrics = []
            for time_point, metric_list in self.metrics.items():
                all_metrics.extend(metric_list)

            if not all_metrics:
                return

            available_metrics = [m for m in all_metrics if m.get("available", False)]
            if not available_metrics:
                return

            proc_metrics = [m for m in available_metrics if "process" in m and m["process"].get("available", True)]
            if proc_metrics:
                proc_mem_values = [m["process"]["memory_rss_mb"] for m in proc_metrics]
                proc_cpu_values = [m["process"]["cpu_percent"] for m in proc_metrics]
                memory_avg = sum(proc_mem_values)/len(proc_mem_values)
                memory_peak = max(proc_mem_values)
                cpu_avg = sum(proc_cpu_values)/len(proc_cpu_values)
                cpu_peak = max(proc_cpu_values)

                proc_disk_metrics = [m for m in proc_metrics if "disk" in m["process"] and m["process"]["disk"].get("available", True)]
                if proc_disk_metrics:
                    proc_disk_read_iops = [m["process"]["disk"]["read_iops"] for m in proc_disk_metrics if "read_iops" in m["process"]["disk"]]
                    proc_disk_write_iops = [m["process"]["disk"]["write_iops"] for m in proc_disk_metrics if "write_iops" in m["process"]["disk"]]
                    proc_disk_total_iops = [m["process"]["disk"]["total_iops"] for m in proc_disk_metrics if "total_iops" in m["process"]["disk"]]

                    if proc_disk_read_iops:
                        disk_read_iops_avg = sum(proc_disk_read_iops) / len(proc_disk_read_iops)
                        disk_read_iops_peak = max(proc_disk_read_iops)
                    if proc_disk_write_iops:
                        disk_write_iops_avg = sum(proc_disk_write_iops) / len(proc_disk_write_iops)
                        disk_write_iops_peak = max(proc_disk_write_iops)
                    if proc_disk_total_iops:
                        disk_total_iops_avg = sum(proc_disk_total_iops) / len(proc_disk_total_iops)
                        disk_total_iops_peak = max(proc_disk_total_iops)

        return {
            "duration": total_time,
            "memory": {
                "avg": memory_avg,
                "peak": memory_peak
            },
            "cpu": {
                "avg": cpu_avg,
                "peak": cpu_peak
            },
            "disk": {
                "read_iops_avg": disk_read_iops_avg,
                "write_iops_avg": disk_write_iops_avg,
                "total_iops_avg": disk_total_iops_avg,
                "read_iops_peak": disk_read_iops_peak,
                "write_iops_peak": disk_write_iops_peak,
                "total_iops_peak": disk_total_iops_peak
            },
            "storage": new_size - self.storage_size,
            "command": self.command_name
        }


def run_command(command, description, stream_output=False, enable_live_monitoring=True, monitor_interval=1.0, quiet=True):
    try:
        if not quiet:
            print(f"  -> Executing: {description}...")
            print(f"     Command: {' '.join(command)}")

        monitor = None
        monitor_results = None
        if enable_live_monitoring:
            monitor = LiveMonitor(description, interval=monitor_interval)

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

        if monitor:
            monitor.start_monitoring(process.pid)

        stdout_data, stderr_data = process.communicate()

        if monitor:
            monitor_results = monitor.stop_monitoring()

        if process.returncode != 0:
            raise subprocess.CalledProcessError(
                returncode=process.returncode,
                cmd=command,
                output=stdout_data,
                stderr=stderr_data
            )

        if not quiet:
            print(f"  ✅ Success: {description}.")
        return True, monitor_results

    except subprocess.CalledProcessError as e:
        capture = not stream_output
        print(f"  ❌ Error: Failed to {description.lower()}.", file=sys.stderr)
        print(f"     Exit Code: {e.returncode}", file=sys.stderr)

        if capture:
            if e.output:
                print(f"     STDOUT:\n{e.output}", file=sys.stderr)
            if e.stderr:
                print(f"     STDERR:\n{e.stderr}", file=sys.stderr)
        return False, None

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

    total_images = len(intersection)
    processed_images = 0

    for name in intersection:
        processed_images += 1
        print(f"\nProcessing image '{name}' ({processed_images}/{total_images})")
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
        Afterwards, each image is pulled again to verify the reduced network usage. Also shows storage size changes.

        Keep in mind that you need specific configuration in storage.conf to take full advantage of zstd:chunked.
        You need to configure:
        enable_partial_images=true
        use_hard_links=true
        """,
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
