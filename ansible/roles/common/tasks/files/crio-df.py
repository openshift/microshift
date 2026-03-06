#!/usr/bin/env python3
"""
CRI-O Storage Disk Usage Reporter
A replacement for 'podman system df -v' that correctly calculates SharedSize for CRI-O storage.
Also calculates compressed (download) sizes from locally stored registry manifests.
"""

import json
import os
import sys
from collections import defaultdict
from datetime import datetime, timezone

# Configuration
STORAGE_ROOT = "/var/lib/containers/storage"
IMAGES_DIR = f"{STORAGE_ROOT}/overlay-images"
IMAGES_JSON = f"{IMAGES_DIR}/images.json"
LAYERS_JSON = f"{STORAGE_ROOT}/overlay-layers/layers.json"
VOLATILE_LAYERS_JSON = f"{STORAGE_ROOT}/overlay-layers/volatile-layers.json"


def load_json_file(filepath):
    """Load and parse a JSON file."""
    try:
        with open(filepath, 'r') as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return []
    except Exception as e:
        print(f"Warning: Could not load {filepath}: {e}", file=sys.stderr)
        return []


def format_size(bytes_value):
    """Format bytes as human-readable string using decimal (SI) units to match podman."""
    if bytes_value == 0:
        return "0B"

    units = ['B', 'KB', 'MB', 'GB', 'TB']
    unit_index = 0
    size = float(bytes_value)

    # Use 1000 (decimal) instead of 1024 (binary) to match podman's output
    while size >= 1000 and unit_index < len(units) - 1:
        size /= 1000
        unit_index += 1

    # Format with appropriate precision
    if unit_index == 0:  # Bytes
        return f"{int(size)}{units[unit_index]}"
    elif size >= 100:
        return f"{size:.0f}{units[unit_index]}"
    elif size >= 10:
        return f"{size:.1f}{units[unit_index]}"
    else:
        return f"{size:.2f}{units[unit_index]}"


def format_time_ago(timestamp):
    """Format timestamp as 'X days/weeks/months ago'."""
    if not timestamp:
        return "Unknown"

    try:
        # Parse the timestamp
        if isinstance(timestamp, str):
            # Handle ISO format with timezone
            created = datetime.fromisoformat(timestamp.replace('Z', '+00:00'))
        else:
            created = datetime.fromtimestamp(timestamp, tz=timezone.utc)

        now = datetime.now(timezone.utc)
        diff = now - created

        days = diff.days
        if days == 0:
            hours = diff.seconds // 3600
            if hours == 0:
                return "Just now"
            return f"{hours} hour{'s' if hours > 1 else ''} ago"
        elif days == 1:
            return "Yesterday"
        elif days < 7:
            return f"{days} days ago"
        elif days < 30:
            weeks = days // 7
            return f"{weeks} week{'s' if weeks > 1 else ''} ago"
        elif days < 365:
            months = days // 30
            return f"{months} month{'s' if months > 1 else ''} ago"
        else:
            years = days // 365
            return f"{years} year{'s' if years > 1 else ''} ago"
    except Exception:
        return "Unknown"


def walk_image_layers(image, layers_by_id):
    """Walk the layer chain from TopLayer up through parents."""
    layers_walked = []
    visited = set()

    # Start from TopLayer
    layer_id = image.get("layer")

    # Walk up the parent chain
    while layer_id and layer_id not in visited:
        visited.add(layer_id)

        layer = layers_by_id.get(layer_id)
        if not layer:
            break

        layers_walked.append(layer_id)
        layer_id = layer.get("parent")

    return layers_walked


def get_image_display_name(image):
    """Get the best display name for an image."""
    names = image.get("names", [])
    if not names:
        return ("<none>", "<none>")

    # Use the first name
    name = names[0]

    # Remove digest if present
    if "@sha256:" in name:
        name = name.split("@")[0]

    # Split into repository and tag
    if ":" in name:
        parts = name.rsplit(":", 1)
        # If the right side still contains '/', this ':' belongs to a registry
        # host:port segment and the image is untagged.
        if "/" in parts[1]:
            return (name, "latest")
        return (parts[0], parts[1])
    else:
        return (name, "latest")


def get_compressed_sizes(images):
    """Read registry manifests from local storage to get compressed layer sizes."""
    all_layers = {}          # digest -> compressed size
    layer_usage = defaultdict(int)
    per_image = []

    for image in images:
        image_id = image["id"]
        manifest_path = os.path.join(IMAGES_DIR, image_id, "manifest")
        if not os.path.exists(manifest_path):
            continue

        manifest = load_json_file(manifest_path)
        if not manifest:
            continue

        layers = manifest.get("layers", [])
        image_compressed = 0
        image_layers = []

        for layer in layers:
            digest = layer.get("digest", "")
            size = layer.get("size", 0)
            image_compressed += size
            image_layers.append({"digest": digest, "size": size})
            all_layers[digest] = size
            layer_usage[digest] += 1

        repo, _ = get_image_display_name(image)
        per_image.append({
            "id": image_id[:12],
            "name": repo,
            "compressed": image_compressed,
            "layers": image_layers,
        })

    shared_digests = {d for d, c in layer_usage.items() if c > 1}
    return all_layers, layer_usage, shared_digests, per_image


def main(verbose=False):
    """Main function to display CRI-O storage disk usage."""

    # Check if running as root
    if os.geteuid() != 0:
        print("Error: This script must be run as root (sudo)", file=sys.stderr)
        sys.exit(1)

    # Load data
    images = load_json_file(IMAGES_JSON)
    layers = load_json_file(LAYERS_JSON)
    volatile_layers = load_json_file(VOLATILE_LAYERS_JSON)

    if not images:
        print("No images found in CRI-O storage")
        return

    # Create layer lookup map
    layers_by_id = {layer["id"]: layer for layer in layers}

    # Count layer usage across all images
    layer_count = defaultdict(int)
    image_layers = {}

    for image in images:
        image_id = image["id"]
        walked_layers = walk_image_layers(image, layers_by_id)
        image_layers[image_id] = walked_layers

        for layer_id in walked_layers:
            layer_count[layer_id] += 1

    # Calculate total storage first (each layer counted exactly once)
    total_size = 0
    for layer_id in layer_count.keys():
        layer = layers_by_id[layer_id]
        diff_size = layer.get("diff-size")
        size = diff_size if diff_size is not None else (layer.get("uncompress_size", 0) or 0)
        total_size += size

    # Map image top layers to image IDs for container counting
    layer_to_image = {}
    for img in images:
        layer_to_image[img.get("layer")] = img["id"]

    # Count containers per image using volatile layers (matching podman's method)
    containers_per_image = defaultdict(int)
    for container in volatile_layers:
        parent = container.get("parent")
        if parent in layer_to_image:
            image_id = layer_to_image[parent]
            containers_per_image[image_id] += 1

    # Calculate sizes for each image
    image_data = []
    total_reclaimable = 0

    for image in images:
        image_id = image["id"]
        walked_layers = image_layers[image_id]

        shared_size = 0
        unique_size = 0

        for layer_id in walked_layers:
            layer = layers_by_id[layer_id]
            diff_size = layer.get("diff-size")
            size = diff_size if diff_size is not None else (layer.get("uncompress_size", 0) or 0)

            if layer_count[layer_id] > 1:
                shared_size += size
            else:
                unique_size += size

        image_size = shared_size + unique_size

        # Count containers using this image (from volatile layers)
        container_count = containers_per_image.get(image_id, 0)

        # Get image metadata
        repo, tag = get_image_display_name(image)
        created = format_time_ago(image.get("created"))

        image_data.append({
            "id": image_id[:12],
            "repository": repo,
            "tag": tag,
            "created": created,
            "size": image_size,
            "shared_size": shared_size,
            "unique_size": unique_size,
            "containers": container_count,
            "reclaimable": unique_size if container_count == 0 else 0
        })

        if container_count == 0:
            total_reclaimable += unique_size

    # Calculate reclaimable percentage
    reclaimable_pct = (total_reclaimable / total_size * 100) if total_size > 0 else 0

    # Calculate compressed download sizes
    comp_layers, _, comp_shared, comp_per_image = get_compressed_sizes(images)
    comp_deduplicated = sum(comp_layers.values())

    if verbose:
        # Detailed output similar to 'podman system df -v'
        print("Images space usage:\n")
        print(f"{'REPOSITORY':<55} {'TAG':<12} {'IMAGE ID':<12} {'CREATED':<10} {'SIZE':<12} {'SHARED SIZE':<12} {'UNIQUE SIZE':<12} {'CONTAINERS'}")

        for img in sorted(image_data, key=lambda x: x["size"], reverse=True):
            print(f"{img['repository']:<55} {img['tag']:<12} {img['id']:<12} {img['created']:<10} "
                  f"{format_size(img['size']):<12} {format_size(img['shared_size']):<12} "
                  f"{format_size(img['unique_size']):<12} {img['containers']}")

        print("\nContainers space usage:\n")
        print(f"{'CONTAINER ID':<12} {'IMAGE':<35} {'COMMAND':<20} {'LOCAL VOLUMES':<15} {'SIZE':<12} {'CREATED':<12} {'STATUS':<12} {'NAMES'}")

        # Container layers from volatile-layers.json
        for container in volatile_layers[:10]:  # Limit to first 10
            container_id = container.get("id", "")[:12]
            parent = container.get("parent", "")
            image = layer_to_image.get(parent, "")[:12] if parent in layer_to_image else "N/A"
            print(f"{container_id:<12} {image:<35} {'N/A':<20} {'0':<15} {'N/A':<12} {'N/A':<12} {'N/A':<12} {'N/A'}")

        print("\nLocal Volumes space usage:\n")
        print(f"{'VOLUME NAME':<30} {'LINKS':<10} {'SIZE'}")
        # No volume info in CRI-O context

        # Compressed download sizes
        print("\nCompressed download sizes:\n")
        for img in sorted(comp_per_image, key=lambda x: x["compressed"], reverse=True):
            print(f"{img['id']}  {format_size(img['compressed']):>10}  {img['name']}")
            for layer in img["layers"]:
                marker = " *" if layer["digest"] in comp_shared else ""
                print(f"  {layer['digest'][:19]}...  {format_size(layer['size']):>10}{marker}")

    else:
        # Summary output similar to 'podman system df'
        print(f"{'TYPE':<15} {'TOTAL':<12} {'ACTIVE':<12} {'SIZE':<15} {'RECLAIMABLE'}")
        print(f"{'Images':<15} {len(images):<12} {len([i for i in image_data if i['containers'] > 0]):<12} "
              f"{format_size(total_size):<15} {format_size(total_reclaimable)} ({reclaimable_pct:.0f}%)")
        print(f"{'Containers':<15} {len(volatile_layers):<12} {'0':<12} {'0B':<15} {'0B (0%)'}")
        print(f"{'Local Volumes':<15} {'0':<12} {'0':<12} {'0B':<15} {'0B (0%)'}")

    # Print summary statistics
    if verbose:
        print("\n" + "="*80)
        print("Storage Summary:")
        print(f"  Total Images: {len(images)}")
        print(f"  Images with containers: {len([i for i in image_data if i['containers'] > 0])}")
        print(f"  Total unique layers: {len(layer_count)}")
        print(f"  Shared layers (used by >1 image): {len([lid for lid, c in layer_count.items() if c > 1])}")
        print(f"  Total storage used: {format_size(total_size)}")
        print(f"  Reclaimable space: {format_size(total_reclaimable)} ({reclaimable_pct:.0f}%)")
        print(f"  Compressed download size: {format_size(comp_deduplicated)}")
        print(f"  Compression ratio: {total_size / comp_deduplicated:.1f}:1" if comp_deduplicated > 0 else "")


if __name__ == "__main__":
    # Check for verbose flag
    verbose = "-v" in sys.argv or "--verbose" in sys.argv
    main(verbose)
