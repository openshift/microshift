#!/usr/bin/env python3
"""Generate Disk IO Performance graph from pcp2json-derived data.

Reads a JSON file with arrays: timestamps, bi, bo, await.
Produces a line chart of Disk Read OPS, Disk Write OPS, and Disk Await.
"""

import argparse
import json
import sys
from datetime import datetime

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt  # noqa: E402
import matplotlib.dates as mdates  # noqa: E402

FMT = "%Y-%m-%d %H:%M:%S"


def main():
    parser = argparse.ArgumentParser(description="Plot disk IO from JSON data")
    parser.add_argument("json_file", help="Path to JSON file with timestamps/bi/bo/await")
    parser.add_argument("-o", "--output", default="/data/disk_io_performance.png",
                        help="Output PNG file path")
    parser.add_argument("--timezone", default="UTC",
                        help="Timezone label for the chart (default: UTC)")
    args = parser.parse_args()

    with open(args.json_file) as f:
        data = json.load(f)

    timestamps = []
    bi_values = []
    bo_values = []
    await_values = []

    for i, ts_str in enumerate(data["timestamps"]):
        try:
            ts = datetime.strptime(ts_str, FMT)
        except ValueError:
            continue
        timestamps.append(ts)
        bi_values.append(float(data["bi"][i]))
        bo_values.append(float(data["bo"][i]))
        await_values.append(float(data["await"][i]))

    if not timestamps:
        print("ERROR: No data points found in JSON", file=sys.stderr)
        sys.exit(1)

    print(f"Loaded {len(timestamps)} data points")

    # Find peak values
    bi_peak_idx = max(range(len(bi_values)), key=lambda i: bi_values[i])
    bo_peak_idx = max(range(len(bo_values)), key=lambda i: bo_values[i])
    q_peak_idx = max(range(len(await_values)), key=lambda i: await_values[i])
    bi_peak_ts = timestamps[bi_peak_idx].strftime("%H:%M:%S")
    bo_peak_ts = timestamps[bo_peak_idx].strftime("%H:%M:%S")
    q_peak_ts = timestamps[q_peak_idx].strftime("%H:%M:%S")
    bi_peak_val = bi_values[bi_peak_idx]
    bo_peak_val = bo_values[bo_peak_idx]
    q_peak_val = await_values[q_peak_idx]

    print(f"Peak bi (read):  {bi_peak_val:,.0f} blk/s at {bi_peak_ts}")
    print(f"Peak bo (write): {bo_peak_val:,.0f} blk/s at {bo_peak_ts}")
    print(f"Peak await:      {q_peak_val:,.1f} ms at {q_peak_ts}")

    fig, (ax, ax_tbl) = plt.subplots(
        2, 1, figsize=(16, 7),
        gridspec_kw={"height_ratios": [8, 1]},
    )

    ax.plot(timestamps, bi_values, label="Disk Read OPS (bi)", color="tab:blue",
            linewidth=0.8, alpha=0.9)
    ax.plot(timestamps, bo_values, label="Disk Write OPS (bo)", color="tab:red",
            linewidth=0.8, alpha=0.9)

    ax.set_xlabel(f"Time ({args.timezone})")
    ax.set_ylabel("Block I/O (blocks/s)")
    ax.set_title("Disk I/O Performance (15-second intervals)")
    ax.grid(True, alpha=0.3)

    # Disk await on secondary Y axis
    ax2 = ax.twinx()
    ax2.plot(timestamps, await_values, label="Disk Await (ms)", color="tab:green",
             linewidth=0.8, alpha=0.7, linestyle="--")
    ax2.set_ylabel("Disk Await (ms)")

    # Combined legend
    lines1, labels1 = ax.get_legend_handles_labels()
    lines2, labels2 = ax2.get_legend_handles_labels()
    ax.legend(lines1 + lines2, labels1 + labels2, loc="upper right")

    ax.xaxis.set_major_formatter(mdates.DateFormatter("%H:%M"))
    ax.xaxis.set_major_locator(mdates.AutoDateLocator())
    plt.setp(ax.xaxis.get_majorticklabels(), rotation=30, ha="right")

    # Small table showing peak values
    ax_tbl.axis("off")
    table = ax_tbl.table(
        cellText=[
            ["Disk Read (bi)", bi_peak_ts, f"{bi_peak_val:,.0f}"],
            ["Disk Write (bo)", bo_peak_ts, f"{bo_peak_val:,.0f}"],
            ["Disk Await (ms)", q_peak_ts, f"{q_peak_val:,.1f}"],
        ],
        colLabels=["Type", "Time", "Peak"],
        cellColours=[["#dbe9f6"] * 3, ["#f6dbdb"] * 3, ["#dbf6db"] * 3],
        colColours=["#cccccc"] * 3,
        loc="center",
        cellLoc="center",
    )
    table.auto_set_font_size(False)
    table.set_fontsize(9)
    table.scale(0.4, 1.2)

    fig.text(0.99, 0.01, f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
             ha="right", va="bottom", fontsize=7, color="gray")

    plt.tight_layout()
    plt.savefig(args.output, dpi=150)
    print(f"Saved chart to {args.output}")


if __name__ == "__main__":
    main()
