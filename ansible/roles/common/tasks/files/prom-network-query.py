#!/usr/bin/env python3
"""Query Prometheus for network transfer during a measurement window.

Usage:
  prom-network-query.py [options] <start_epoch> <end_epoch>
  prom-network-query.py [options] <hours>
  prom-network-query.py [options]

Options:
  --prometheus URL        Prometheus base URL (default: http://zen3:9091)
  --instance HOST:PORT    Prometheus instance label to filter by (e.g. microshift:9100)
  --device IFACE          Network device (default: enp5s0)
  --step STEP             Range query step (default: 15s)

Uses increase() to get accurate totals that handle counter resets natively.
Also shows raw counter time series for visibility into resets.
"""

import argparse
import http.client
import ipaddress
import json
import socket
import sys
import urllib.parse
from dataclasses import dataclass
from datetime import datetime, timezone

DEFAULT_PROMETHEUS = "http://zen3:9091"
DEFAULT_DEVICE = "enp5s0"
DEFAULT_STEP = "15s"


class PromQueryError(RuntimeError):
    """Raised when a Prometheus query fails or returns ambiguous data."""


@dataclass(frozen=True)
class PrometheusEndpoint:
    """Normalized Prometheus endpoint details."""

    scheme: str
    host: str
    port: int
    base_path: str = ""

    @property
    def display_host(self):
        if ":" in self.host and not self.host.startswith("["):
            return f"[{self.host}]"
        return self.host

    @property
    def base_url(self):
        return f"{self.scheme}://{self.display_host}:{self.port}{self.base_path}"


def _is_local_or_private_ip(value):
    ip = ipaddress.ip_address(value)
    return ip.is_private or ip.is_loopback or ip.is_link_local


def _validate_prometheus_host(host):
    """Allow only loopback/link-local/private Prometheus endpoints."""
    try:
        if not _is_local_or_private_ip(host):
            raise argparse.ArgumentTypeError(
                f"--prometheus host must be loopback/link-local/private, got {host!r}"
            )
        return
    except ValueError:
        pass

    try:
        infos = socket.getaddrinfo(host, None, type=socket.SOCK_STREAM)
    except socket.gaierror as exc:
        raise argparse.ArgumentTypeError(
            f"--prometheus host {host!r} could not be resolved: {exc}"
        ) from exc

    disallowed = set()
    for _, _, _, _, sockaddr in infos:
        resolved = sockaddr[0].split("%", 1)[0]  # strip IPv6 scope suffix if present
        try:
            if not _is_local_or_private_ip(resolved):
                disallowed.add(resolved)
        except ValueError:
            disallowed.add(resolved)

    if disallowed:
        raise argparse.ArgumentTypeError(
            f"--prometheus host {host!r} resolves to non-private address(es): {', '.join(sorted(disallowed))}"
        )


def _validate_prometheus_url(value):
    """Validate and normalize --prometheus URL to prevent SSRF."""
    parsed = urllib.parse.urlparse(value)
    if parsed.scheme not in ("http", "https"):
        raise argparse.ArgumentTypeError(
            f"--prometheus must use http or https, got: {parsed.scheme!r}"
        )
    if parsed.username or parsed.password:
        raise argparse.ArgumentTypeError("--prometheus must not include credentials")
    if parsed.query or parsed.fragment or parsed.params:
        raise argparse.ArgumentTypeError(
            "--prometheus must not include query, fragment, or params"
        )
    if parsed.hostname is None:
        raise argparse.ArgumentTypeError("--prometheus must include a hostname")

    try:
        port = parsed.port
    except ValueError as exc:
        raise argparse.ArgumentTypeError(f"--prometheus has invalid port: {exc}") from exc

    if port is None:
        port = 443 if parsed.scheme == "https" else 80

    base_path = parsed.path.rstrip("/")
    _validate_prometheus_host(parsed.hostname)
    return PrometheusEndpoint(parsed.scheme, parsed.hostname, port, base_path)


def parse_window_args(argv):
    parser = argparse.ArgumentParser(
        description="Query Prometheus for RX/TX transfer over a measurement window."
    )
    parser.add_argument(
        "window",
        nargs="*",
        help="Either <start_epoch> <end_epoch>, or <hours>, or empty for last 4h.",
    )
    parser.add_argument(
        "--prometheus",
        default=DEFAULT_PROMETHEUS,
        type=_validate_prometheus_url,
        help=f"Prometheus base URL (default: {DEFAULT_PROMETHEUS})",
    )
    parser.add_argument(
        "--device",
        default=DEFAULT_DEVICE,
        help=f"Network device label value (default: {DEFAULT_DEVICE})",
    )
    parser.add_argument(
        "--instance",
        default=None,
        help="Prometheus instance label to filter by (e.g. microshift:9100)",
    )
    parser.add_argument(
        "--step",
        default=DEFAULT_STEP,
        help=f"Range query step for raw series output (default: {DEFAULT_STEP})",
    )
    args = parser.parse_args(argv)

    if len(args.window) == 2:
        try:
            start = int(args.window[0])
            end = int(args.window[1])
        except ValueError as exc:
            parser.error(f"start/end must be integers: {exc}")
    elif len(args.window) == 1:
        try:
            hours = int(args.window[0])
        except ValueError as exc:
            parser.error(f"hours must be an integer: {exc}")
        now = int(datetime.now(timezone.utc).timestamp())
        start = now - hours * 3600
        end = now
    elif len(args.window) == 0:
        now = int(datetime.now(timezone.utc).timestamp())
        start = now - 4 * 3600
        end = now
    else:
        parser.error("Provide either <start_epoch> <end_epoch>, <hours>, or no positional args.")

    if end <= start:
        parser.error("end_epoch must be greater than start_epoch.")

    return args, start, end


def prom_query(prometheus, endpoint, params):
    qs = urllib.parse.urlencode(params)
    api_path = f"{prometheus.base_path}/api/v1/{endpoint}" if prometheus.base_path else f"/api/v1/{endpoint}"
    request_path = f"{api_path}?{qs}"
    request_url = f"{prometheus.base_url}{request_path}"
    connection_class = http.client.HTTPSConnection if prometheus.scheme == "https" else http.client.HTTPConnection
    connection = connection_class(prometheus.host, prometheus.port, timeout=30)

    try:
        connection.request("GET", request_path, headers={"Accept": "application/json"})
        response = connection.getresponse()
        body = response.read()
    except (http.client.HTTPException, OSError, TimeoutError) as exc:
        raise PromQueryError(f"request failed for {request_url}: {exc}") from exc
    finally:
        connection.close()

    if response.status >= 400:
        response_body = body.decode("utf-8", errors="replace").strip()
        if response_body:
            raise PromQueryError(
                f"HTTP {response.status} from {request_url}: {response_body}"
            )
        raise PromQueryError(
            f"HTTP {response.status} from {request_url}: {response.reason}"
        )

    try:
        data = json.loads(body)
    except json.JSONDecodeError as exc:
        raise PromQueryError(f"invalid JSON from {request_url}: {exc}") from exc

    if data.get("status") != "success":
        error_type = data.get("errorType", "unknown_error")
        error_msg = data.get("error", "no details")
        raise PromQueryError(
            f"Prometheus API error ({endpoint}): {error_type}: {error_msg}"
        )

    return data


def query_range(prometheus, query, start, end, step):
    return prom_query(prometheus, "query_range", {
        "query": query, "start": start, "end": end, "step": step,
    })


def query_instant(prometheus, query, time):
    return prom_query(prometheus, "query", {"query": query, "time": time})


def fmt_ts(epoch):
    return datetime.fromtimestamp(epoch).strftime("%Y-%m-%d %H:%M:%S")


def fmt_bytes(b):
    if b >= 1024 * 1024 * 1024:
        return f"{b / 1024 / 1024 / 1024:.2f} GiB"
    if b >= 1024 * 1024:
        return f"{b / 1024 / 1024:.1f} MiB"
    if b >= 1024:
        return f"{b / 1024:.1f} KiB"
    return f"{b} B"


def label_selector(device, instance=None):
    """Build PromQL label selector string."""
    labels = f'device="{device}"'
    if instance:
        labels += f',instance="{instance}"'
    return labels


def get_single_series(result, context):
    """Return exactly one series; fail if response is ambiguous."""
    if not result:
        return None
    if len(result) > 1:
        raise PromQueryError(
            f"{context}: expected 1 series, got {len(result)}; "
            "tighten labels (--instance/--device)."
        )
    return result[0]


def get_increase(metric, start, end, prometheus, device, instance=None):
    """Use increase() to get accurate total accounting for counter resets."""
    window = end - start
    sel = label_selector(device, instance)
    # Deduplicate label variants (job/pod/namespace/etc.) for the same host/device.
    query = f"max by (instance, device) (increase({metric}{{{sel}}}[{window}s]))"
    data = query_instant(prometheus, query, end)
    series = get_single_series(data["data"]["result"], f"{metric} increase()")
    if series is None:
        return None
    try:
        return float(series["value"][1])
    except (KeyError, IndexError, TypeError, ValueError) as exc:
        raise PromQueryError(f"{metric} increase(): unexpected response shape") from exc


def get_time_series(start, end, prometheus, device, step, instance=None):
    """Get raw counter values to show resets."""
    sel = label_selector(device, instance)
    # Deduplicate label variants (job/pod/namespace/etc.) for the same host/device.
    query = f"max by (instance, device) (node_network_receive_bytes_total{{{sel}}})"
    data = query_range(prometheus, query, start, end, step)
    series = get_single_series(data["data"]["result"], "raw receive counter")
    if series is None:
        return []
    try:
        return series["values"]
    except (KeyError, TypeError) as exc:
        raise PromQueryError("raw receive counter: unexpected response shape") from exc


def main():
    args, start, end = parse_window_args(sys.argv[1:])

    duration = end - start
    print(f"Measurement window: {fmt_ts(start)} -> {fmt_ts(end)} ({duration}s)")
    print(f"Prometheus: {args.prometheus.base_url}")
    print(f"Instance: {args.instance or '(all)'}")
    print(f"Device: {args.device}")
    print()

    try:
        # Get increase() — the accurate total from Prometheus
        rx_increase = get_increase(
            "node_network_receive_bytes_total", start, end,
            args.prometheus, args.device, args.instance,
        )
        tx_increase = get_increase(
            "node_network_transmit_bytes_total", start, end,
            args.prometheus, args.device, args.instance,
        )
    except PromQueryError as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1

    print("=== Network transfer (Prometheus increase) ===")
    if rx_increase is not None:
        print(f"  RX: {fmt_bytes(rx_increase)} ({int(rx_increase):,} bytes)")
    else:
        print("  RX: no data")
    if tx_increase is not None:
        print(f"  TX: {fmt_bytes(tx_increase)} ({int(tx_increase):,} bytes)")
    else:
        print("  TX: no data")
    if rx_increase is not None and tx_increase is not None:
        print(f"  Total: {fmt_bytes(rx_increase + tx_increase)}")
    print()

    try:
        # Get raw time series for visibility
        values = get_time_series(
            start, end, args.prometheus, args.device, args.step, args.instance
        )
    except PromQueryError as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1

    if not values:
        print("No time series data available.")
        return 0

    print(f"=== Raw counter time series ({len(values)} samples, step={args.step}) ===")
    print(f"{'timestamp':<24} {'rx_bytes':>16} {'rx_MB':>10} {'delta_MB':>10}")
    print("-" * 64)

    prev_val = None
    resets = []
    for ts, val in values:
        val = int(float(val))
        dt = fmt_ts(ts)
        mb = val / 1024 / 1024
        delta = ""
        if prev_val is not None:
            diff = val - prev_val
            delta = f"{diff / 1024 / 1024:>10.2f}"
            if diff < 0:
                delta += " *** RESET ***"
                resets.append((ts, prev_val, val))
        print(f"{dt:<24} {val:>16} {mb:>10.2f} {delta}")
        prev_val = val

    if resets:
        print()
        print(f"Counter resets: {len(resets)}")
        for ts, before, after in resets:
            print(f"  {fmt_ts(ts)}: {before:,} -> {after:,} "
                  f"(lost {fmt_bytes(before - after)})")

    # Output JSON summary for machine parsing
    print()
    summary = {
        "start_epoch": start,
        "end_epoch": end,
        "duration_seconds": duration,
        "device": args.device,
        "rx_bytes": int(rx_increase) if rx_increase is not None else 0,
        "tx_bytes": int(tx_increase) if tx_increase is not None else 0,
        "counter_resets": len(resets),
        "samples": len(values),
    }
    print(json.dumps(summary))
    return 0


if __name__ == "__main__":
    sys.exit(main())
