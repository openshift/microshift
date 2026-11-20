import re
import sys
import requests
import libipv6

_CERT_LABEL_RE = re.compile(r'(\w+)="([^"]*)"')
_CERT_SERIES = {
    "microshift_certificate_not_before_timestamp_seconds": "not_before",
    "microshift_certificate_not_after_timestamp_seconds": "not_after",
    "microshift_certificate_status": "status",
}


def _run_prometheus_query(host: str, port: int, query: str) -> requests.Response:
    """Executes the given query against the prometheus instance.
    Fails if the response is empty."""
    base_url = f"http://{host}:{port}/api/v1/query"
    encoded_query = requests.utils.quote(query)
    url = f"{base_url}?query={encoded_query}"
    _log(f"Querying Prometheus Server: {url}")
    response = requests.get(url)
    _log(f"Response: {response.status_code} {response.text}")
    if response.status_code != 200:
        raise Exception(f"Prometheus query failed with status code {response.status_code}")
    if response.json().get("status") != "success":
        raise Exception(f"Prometheus query failed with status: {response.json().get('status')}")
    return response


def _add_hostname_filter(query: str, add_hostname_filter: bool = True) -> str:
    if not add_hostname_filter:
        return query

    try:
        from robot.libraries.BuiltIn import BuiltIn
        # Running within RF
        stdout, _, _ = BuiltIn().run_keyword("Command Execution", "hostname")
        if stdout:
            query = f"{query}{{host_name=\"{stdout}\"}}"
            _log(f"Added hostname to query: {query}")
    except Exception:
        None
    return query


def check_prometheus_query(host: str, port: int, query: str, add_hostname_filter: bool = True) -> None:
    query = _add_hostname_filter(query, add_hostname_filter)
    response = _run_prometheus_query(host, port, query)
    data_result_list: list = response.json().get("data", {}).get("result")
    if not data_result_list or len(data_result_list) == 0:
        raise Exception("Prometheus query returned no results")


def check_prometheus_exporter(host: str, port: int, query: str) -> None:
    """Check the metric is available int Prometheus Exporter
    Fails if the response is empty."""
    address = libipv6.add_brackets_if_ipv6(host)
    base_url = f"http://{address}:{port}/metrics"
    _log(f"Querying Prometheus Exporter: {base_url}")
    response = requests.get(base_url)
    _log(f"Response: {response.status_code} {response.text}")
    if response.status_code != 200:
        raise Exception(f"Prometheus query failed with status code {response.status_code}")
    if not response.text or len(response.text) == 0:
        raise Exception("Prometheus query returned no results")
    for line in response.text.splitlines():
        if line.startswith(query):
            _log(f"Found metric: {line}")
            return
    raise Exception(f"Prometheus query returned no results for metric: {query}")


def parse_certificate_metrics(text: str) -> dict:
    """Parse the microshift_certificate_* series of a Prometheus text exposition.

    Returns {"pending", "success", "last_success", "certs"}, where certs maps
    "<service>/<name>" to its not_before, not_after, status (the state whose
    gauge is 1) and rotation_policy. Unknown series are ignored. A failed
    collection drops the certificate and renewal_pending series, leaving certs
    empty and pending None.
    """
    result = {"pending": None, "success": 0.0, "last_success": 0.0, "certs": {}}
    for line in text.splitlines():
        if not line.startswith("microshift_certificate_"):
            continue
        name_part, _, value = line.rpartition(" ")
        metric, _, label_block = name_part.partition("{")
        labels = dict(_CERT_LABEL_RE.findall(label_block))
        value = float(value)
        if metric == "microshift_certificate_renewal_pending":
            result["pending"] = value
        elif metric == "microshift_certificate_inventory_collection_success":
            result["success"] = value
        elif metric == "microshift_certificate_inventory_last_success_timestamp_seconds":
            result["last_success"] = value
        elif metric in _CERT_SERIES:
            cert = result["certs"].setdefault(f"{labels['service']}/{labels['name']}", {
                "rotation_policy": labels["rotation_policy"],
                "not_before": None,
                "not_after": None,
                "status": None,
            })
            field = _CERT_SERIES[metric]
            if field != "status":
                cert[field] = value
            elif value == 1.0:
                cert["status"] = labels["status"]
    return result


if __name__ == '__main__':
    _log = print
    if len(sys.argv) != 5:
        _log(f"Error: Expected exactly 4 arguments, but got {len(sys.argv)-1}")
        _log("Usage: python script.py (query|exporter) host port query")
        sys.exit(1)

    action: str = sys.argv[1]
    host: str = sys.argv[2]
    port: int = int(sys.argv[3])
    query: str = sys.argv[4]

    if action == "query":
        check_prometheus_query(host, port, query)
    elif action == "exporter":
        check_prometheus_exporter(host, port, query)
    else:
        _log(f"ERROR: Unknown action '{action}'. Expected 'query' or 'exporter'.")
        sys.exit(1)
else:
    from robot.libraries.BuiltIn import BuiltIn
    _log = BuiltIn().log
