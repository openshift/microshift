import sys
import json
import requests
from datetime import datetime, timedelta


def query_loki(loki_url: str, query: str, limit: int = 10, start_time: datetime = None, end_time: datetime = None) -> dict:
    """
    Query Loki server with LogQL

    Args:
        loki_url (str): Base URL of Loki server (e.g., "http://localhost:3100")
        query (str): LogQL query expression
        limit (int, optional): Maximum number of entries to return
        start_time (datetime, optional): Start time for query window
        end_time (datetime, optional): End time for query window

    Returns:
        dict: JSON response from Loki
    """
    # Set start and end times
    if not end_time:
        end_time = datetime.now()
    if not start_time:
        start_time = end_time - timedelta(hours=1)  # Default to last hour

    # Create request
    url = f"{loki_url.rstrip('/')}/loki/api/v1/query_range"
    params = {
        "query": query,
        "start": int(start_time.timestamp() * 1e9),
        "end": int(end_time.timestamp() * 1e9),
        "limit": limit,
    }

    _log(f"Querying Loki at {url}")
    response = requests.get(url, params=params)
    _log(f"Response: {response.status_code} {response.text}")
    response.raise_for_status()
    return response.json()


def _print_results(results: dict) -> None:
    """Pretty print the Loki query results"""
    if not results or "data" not in results:
        _log("ERROR: No results or invalid response")
        sys.exit(1)

    streams = results["data"]["result"]
    for i, stream in enumerate(streams):
        _log(f"Labels: {stream['stream']}")
        for entry in stream["values"]:
            log_line = entry[1]
            try:
                log_line = json.loads(log_line)
                if "MESSAGE" in log_line["body"]:
                    _log(f"{log_line['body']['MESSAGE']}")
                else:
                    _log(f"{log_line['body']}")
            except json.JSONDecodeError as e:
                raise Exception(f"ERROR: Error decoding log line as JSON: {e}")


def check_loki_query(host: str, port: int, query: str, limit: int = 10) -> None:
    results = query_loki(f"http://{host}:{port}", query, limit)
    _print_results(results)


if __name__ == "__main__":
    _log = print
    if len(sys.argv) < 3:
        _log("Usage: python loki_query.py <loki_url> <query_expression> [limit]")
        _log("Example: python test/resources/loki_agullon.py localhost 3100 '{job=\\\"journald\\\",exporter=\\\"OTLP\\\"}' 10")
        sys.exit(1)

    host: str = sys.argv[1]
    port: int = sys.argv[2]
    query: str = sys.argv[3]
    limit: int = int(sys.argv[4]) if len(sys.argv) > 4 else 10

    check_loki_query(host, port, query, limit)
else:
    from robot.libraries.BuiltIn import BuiltIn
    _log = BuiltIn().log
