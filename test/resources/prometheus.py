from robot.libraries.BuiltIn import BuiltIn

import requests

_log = BuiltIn().log


def check_prometheus_query(host: str, port: int, query: str) -> None:
    """Executes the given query against the prometheus instance.
    Fails if the response is empty."""
    base_url = f"http://{host}:{port}/api/v1/query"
    encoded_query = requests.utils.quote(query)
    url = f"{base_url}?query={encoded_query}"
    _log(f"Querying Prometheus at {url}")
    response = requests.get(url)
    _log(f"Response: {response.status_code} {response.text}")
    if response.status_code != 200:
        raise Exception(f"Prometheus query failed with status code {response.status_code}")
    if response.json().get("status") != "success":
        raise Exception(f"Prometheus query failed with status: {response.json().get('status')}")
    if not response.json().get("data", {}).get("result") or len(response.json().get("data", {}).get("result")) == 0:
        raise Exception("Prometheus query returned no results")
