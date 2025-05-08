import sys
import requests
import libipv6


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


def check_prometheus_query(host: str, port: int, query: str) -> None:
    response = _run_prometheus_query(host, port, query)
    data_result_list: list = response.json().get("data", {}).get("result")
    if not data_result_list or len(data_result_list) == 0:
        raise Exception("Prometheus query returned no results")


def check_prometheus_query_is_missing(host: str, port: int, query: str) -> None:
    response = _run_prometheus_query(host, port, query)
    data_result_list: list = response.json().get("data", {}).get("result")
    if data_result_list and len(data_result_list) > 0:
        raise Exception("Prometheus query returned results")


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
