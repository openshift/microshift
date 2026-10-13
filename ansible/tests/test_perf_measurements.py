"""Local-only tests; no setup-node playbook or remote inventory is executed.

Run: python3 -m unittest discover -s ansible/tests -v
"""

import contextlib
import importlib.util
import io
import http.server
import threading
import json
import os
import secrets
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

FILES = Path(__file__).resolve().parents[1] / "roles/common/tasks/files"


def load(name):
    spec = importlib.util.spec_from_file_location(name, FILES / f"{name}.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


class Producers(unittest.TestCase):
    def test_startup_sidecar_preserves_text(self):
        for nrestarts, expected in (("2", 2), ("", None)):
            with self.subTest(nrestarts=nrestarts), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                commands = {
                    "sudo": "exit 0", "mkdir": "exit 0", "date": "echo 1780077369",
                    "systemctl": 'case "$*" in *ActiveEnterTimestampMonotonic*) echo 36306000 ;; *InactiveExitTimestampMonotonic*) echo 1000000 ;; *NRestarts*) echo "' + nrestarts + '" ;; esac',
                    "systemd-analyze": "echo '35.306s microshift.service'",
                    "oc": 'case "$*" in *jsonpath*) echo "True True" ;; *"-o json"*) echo \'{"items":[]}\' ;; *) echo "pod 1/1 Running" ;; esac',
                }
                for name, body in commands.items():
                    path = root / name
                    path.write_text("#!/bin/sh\n" + body + "\n")
                    path.chmod(0o700)
                env = {**os.environ, "PATH": str(root) + ":" + os.environ["PATH"]}
                env.pop("PERF_JSON_OUTPUT", None)
                command = ["bash", str(FILES / "ready.sh"), "1", "1"]
                before = subprocess.run(command, env=env, capture_output=True, text=True, check=True)
                output = root / "boot.json"
                after = subprocess.run(command, env={**env, "PERF_JSON_OUTPUT": str(output)}, capture_output=True, text=True, check=True)
                self.assertEqual(before.stdout, after.stdout)
                data = json.loads(output.read_text())
                self.assertEqual(data["service_seconds"], 35.306)
                self.assertEqual(data["kubeconfig_seconds"], 0)
                self.assertEqual(data["start_epoch"], 1780077369)
                self.assertEqual(data["end_epoch"], 1780077369)
                self.assertEqual(data["nrestarts"], expected)

    def test_startup_jq_avoids_reserved_names(self):
        # jq 1.6 (RHEL 9) rejects keywords such as `end` as variable names.
        keywords = ("and", "as", "catch", "def", "elif", "else", "end", "foreach",
                    "if", "import", "include", "label", "not", "or", "reduce", "then", "try")
        script = (FILES / "ready.sh").read_text()
        for keyword in keywords:
            self.assertNotRegex(script, rf"--arg(json)? {keyword}\b")

    def test_disk_bytes_and_legacy_text(self):
        disk = load("disk-usage")
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "large").write_bytes(b"x" * 1200001)
            (root / "small").write_bytes(b"x")
            for request in (directory, str(root / "*")):
                result = disk.measure(request)
                old = subprocess.run(
                    f"du -BM -t 1M --max-depth=1 {request}",
                    shell=True, capture_output=True, text=True, check=True,
                )
                self.assertEqual(sorted(result["stdout"].splitlines()), sorted(old.stdout.splitlines()))
                self.assertTrue(all(isinstance(row["allocated_bytes"], int) for row in result["paths"]))
                large = next(row for row in disk.measure(str(root / "large"))["paths"])
                self.assertEqual(large["allocated_bytes"], (root / "large").stat().st_blocks * 512)
            missing = disk.measure(str(root / "missing"))
            self.assertEqual(missing["rc"], 1)
            self.assertEqual(missing["paths"], [])

    def test_images_use_unrounded_deduplicated_bytes(self):
        images = load("crio-df")
        fixtures = {
            images.IMAGES_JSON: [
                {"id": "a", "layer": "one", "names": ["registry/image:a"]},
                {"id": "b", "layer": "two", "names": ["registry/image:b"]},
            ],
            images.LAYERS_JSON: [
                {"id": "base", "diff-size": 1200001},
                {"id": "one", "parent": "base", "diff-size": 2000003},
                {"id": "two", "parent": "base", "diff-size": 3000007},
            ],
            images.VOLATILE_LAYERS_JSON: [{"parent": "one"}],
        }
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "images.json"
            with patch.object(images.os, "geteuid", return_value=0), \
                 patch.object(images, "load_json_file", side_effect=lambda path: fixtures.get(path, [])), \
                 patch.object(images, "get_compressed_sizes", return_value=({"digest": 1234567}, {}, set(), [{"id": "a", "compressed": 1234567, "name": "image", "layers": []}])):
                old_text, new_text = io.StringIO(), io.StringIO()
                with contextlib.redirect_stdout(old_text):
                    images.main(True)
                with contextlib.redirect_stdout(new_text):
                    images.main(True, output)
            self.assertEqual(old_text.getvalue(), new_text.getvalue())
            data = json.loads(output.read_text())
            self.assertEqual(data["storage_bytes"], 6200011)
            self.assertEqual(data["compressed_download_bytes"], 1234567)
            self.assertEqual(data["reclaimable_bytes"], 3000007)
            self.assertEqual(data["shared_layers"], 1)

    def test_network_sidecar_preserves_text_and_missing_values(self):
        network = load("prom-network-query")
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "network.json"
            for values, increases in [([[100, "100"], [110, "10"]], [1040562641.9, 5070460.2]), ([], [None, None])]:
                old_text, new_text = io.StringIO(), io.StringIO()
                for destination, extra in [(old_text, []), (new_text, ["--json-output", str(output)])]:
                    with patch.object(sys, "argv", ["query", "--prometheus", "http://127.0.0.1:9091", "100", "120", *extra]), \
                         patch.object(network, "get_increase", side_effect=increases), \
                         patch.object(network, "get_time_series", return_value=values), \
                         contextlib.redirect_stdout(destination):
                        self.assertEqual(network.main(), 0)
                self.assertEqual(old_text.getvalue(), new_text.getvalue())
                data = json.loads(output.read_text())
                self.assertEqual(data["rx_bytes"], int(increases[0]) if increases[0] is not None else None)
                self.assertEqual(data["samples"], len(values))
                self.assertEqual(data["counter_resets"], 1 if values else 0)


class ArchiveAndIndex(unittest.TestCase):
    def exercise(self, response_status=200, item_error=False, enabled=True, metadata=None,
                 body=None, credentials="generated", url=None, allow_http=True,
                 requests=None, failure=None, message=None):
        """Archive four measurements, upload twice, then archive one more.

        credentials: "generated" (random password), "password-only", a dict, "missing", "malformed" or None
        (no file and no environment credentials). url: format string with {port}.
        requests/failure default to what an ordinary upload with this response would do.
        """
        received = []
        secret = secrets.token_urlsafe(16)

        class Handler(http.server.BaseHTTPRequestHandler):
            def do_POST(self):
                payload = self.rfile.read(int(self.headers["Content-Length"]))
                assert payload.endswith(b"\n")
                rows = [json.loads(line) for line in payload.splitlines()]
                assert len(rows) == 8
                received.append(rows)
                self.send_response(response_status)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(body if body is not None else json.dumps({
                    "errors": item_error,
                    "items": [{"index": {"status": 400 if item_error else 201}}
                              for _ in range(4)],
                }).encode())

            def log_message(self, *args):
                pass

        if failure is None:
            failure = enabled and (response_status != 200 or item_error or body is not None)
        if requests is None:
            requests = (1 if failure else 2) if enabled else 0

        server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            with tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                (root / "roles").symlink_to(FILES.parents[2])
                output = root / "results"
                output.mkdir()
                credentials_file = root / "credentials.json"
                if credentials == "password-only":
                    credentials = {"username": "", "password": secret}
                if credentials == "generated":
                    credentials = {"username": "fixture", "password": secret}
                if isinstance(credentials, dict):
                    credentials_file.write_text(json.dumps(credentials))
                    credentials_file.chmod(0o600)
                elif credentials == "malformed":
                    credentials_file.write_text("not json")
                    credentials_file.chmod(0o600)
                endpoint = (url or "http://127.0.0.1:{port}").format(port=server.server_port, secret=secret)
                variables = {
                    "ansible_python_interpreter": sys.executable,
                    "results_dir": str(output), "perf_run_id": "fixture-run",
                    "perf_es_index": "fixture",
                    "perf_es_credentials_file": str(credentials_file) if credentials is not None else "",
                    "perf_es_url": endpoint if enabled else "",
                    "perf_es_allow_http": allow_http,
                    "perf_microshift_version": "4.22.0~rc.5",
                    "perf_microshift_nevra": "microshift-0:4.22.0~rc.5-1.el9.x86_64",
                    "ansible_facts": {"hostname": "fixture", "kernel": "test-kernel",
                                      "distribution_version": "9.8", "processor": ["0", "AuthenticAMD", "Test CPU"]},
                    "expect_failure": failure,
                }
                if metadata is not None:
                    variables["perf_extra_metadata"] = metadata
                (root / "variables.json").write_text(json.dumps(variables))
                (root / "test.yml").write_text("""---
- name: Local archive and upload test
  hosts: localhost
  connection: local
  gather_facts: false
  tasks:
    - name: Archive each measurement family
      ansible.builtin.include_tasks: roles/common/tasks/perf-record.yml
      loop: [startup, disk, images, startup_network]
      loop_control:
        loop_var: family
      vars:
        perf_measurement_key: "{{ family }}"
        perf_measurement_type: "{{ family }}"
        perf_measurement_timestamp: '2026-10-12T00:00:00Z'
        perf_measurement_data: {value_bytes: 1234567}
    - name: Upload
      ansible.builtin.include_tasks: roles/common/tasks/perf-index.yml
    - name: Replay or skip after failure
      ansible.builtin.include_tasks: roles/common/tasks/perf-index.yml
    - name: Check result
      ansible.builtin.assert:
        that: perf_es_unavailable | default(false) | bool == expect_failure | bool
    - name: Continue capture after upload
      ansible.builtin.include_tasks: roles/common/tasks/perf-record.yml
      vars:
        perf_measurement_key: after-upload
        perf_measurement_type: disk
        perf_measurement_timestamp: '2026-10-12T00:00:01Z'
        perf_measurement_data: {value_bytes: 7654321}
""")
                env = {key: value for key, value in os.environ.items()
                       if key not in ("PERF_ES_USERNAME", "PERF_ES_PASSWORD")}
                result = subprocess.run(
                    ["ansible-playbook", "-i", "localhost,", str(root / "test.yml"),
                     "-e", "@" + str(root / "variables.json")],
                    env={**env, "ANSIBLE_HOME": str(root / "ansible-home"),
                         "ANSIBLE_LOCAL_TEMP": str(root / "ansible-tmp")},
                    capture_output=True, text=True, timeout=90,
                )
                logs = result.stdout + result.stderr
                self.assertEqual(result.returncode, 0, logs)
                self.assertNotIn(secret, logs)
                self.assertEqual("WARNING: Performance indexing failed" in result.stdout, failure, logs)
                if message is not None:
                    self.assertIn(message, result.stdout)
                docs = [json.loads(line) for line in (output / "measurements.ndjson").read_text().splitlines()]
                self.assertEqual(len(docs), 5)
                self.assertEqual(len({doc["measurement_id"] for doc in docs}), 5)
                self.assertTrue(all(doc["cpu_model"] == "Test CPU" for doc in docs))
                self.assertTrue(all(doc["labels"] == (metadata or {}) for doc in docs))
                self.assertTrue(all(doc["run_id"] == "fixture-run" for doc in docs))
                for batch in received:
                    self.assertTrue(all(doc["labels"] == (metadata or {}) for doc in batch[1::2]))
                self.assertEqual(len(received), requests)
                if requests == 2:
                    self.assertEqual(received[0], received[1])
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

    def test_offline_archive(self):
        self.exercise(enabled=False)

    def test_disabled_indexing_ignores_missing_credentials(self):
        self.exercise(enabled=False, credentials="missing")

    def test_extra_metadata_in_archive_and_bulk(self):
        self.exercise(metadata={
            "cycle_id": "cycle-5.0", "iteration": "3", "rc": "5.0.0-rc.1",
            "kube_burner_uuid": "density-uuid", "netperf_uuid": "network-uuid",
            "run_id": "label-does-not-override-envelope",
        })

    def test_bulk_replay(self):
        self.exercise()

    def test_bulk_item_failure_continues_capture(self):
        self.exercise(item_error=True, message="rejected documents: 4")

    def test_auth_failure_continues_capture(self):
        self.exercise(response_status=401, message="HTTP status: 401")

    def test_malformed_bulk_responses_continue_capture(self):
        for body in (b"[]", b"null", b'"text"', b"not json", b"{}", b'{"errors": true, "items": 7}',
                     b'{"errors": true, "items": [7]}', b'{"errors": true, "items": [{"index": null}]}',
                     b'{"errors": true, "items": [{"index": {}}]}', b'{"errors": false, "items": null}',
                     b'{"errors": false, "items": {}}', b'{"errors": false, "items": "text"}',
                     b'{"errors": true, "items": [{"index": {"status": "400"}}]}'):
            with self.subTest(body=body):
                self.exercise(body=body, message="rejected documents: unavailable")

    def test_http_credentials_refused_by_default(self):
        self.exercise(allow_http=False, requests=0, failure=True, message="refusing to send credentials")

    def test_http_username_only_refused_by_default(self):
        self.exercise(allow_http=False, credentials={"username": "fixture", "password": ""},
                      requests=0, failure=True, message="refusing to send credentials")

    def test_http_password_only_refused_by_default(self):
        self.exercise(allow_http=False, credentials="password-only",
                      requests=0, failure=True, message="refusing to send credentials")

    def test_http_without_credentials_allowed(self):
        self.exercise(allow_http=False, credentials=None)

    def test_url_credentials_refused(self):
        self.exercise(url="http://fixture:{secret}@127.0.0.1:{port}", credentials=None,
                      requests=0, failure=True, message="must not contain credentials")

    def test_unsupported_scheme_refused(self):
        self.exercise(url="ftp://127.0.0.1:{port}", requests=0, failure=True,
                      message="must use http or https")

    def test_missing_credentials_file_continues_capture(self):
        self.exercise(credentials="missing", requests=0, failure=True)

    def test_malformed_credentials_file_continues_capture(self):
        self.exercise(credentials="malformed", requests=0, failure=True)


if __name__ == "__main__":
    unittest.main()
