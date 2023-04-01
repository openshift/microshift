#!/usr/bin/env python3

import os
import sys
import time
import logging
import tempfile
import argparse
import subprocess
from pathlib import Path
import xml.etree.ElementTree as ET

logging.basicConfig(level=logging.DEBUG, format='%(asctime)s   %(levelname)s   %(message)s')

_SCRIPT_DIR = os.path.realpath(os.path.dirname(__file__))
_TESTS_DIR = Path(f"{_SCRIPT_DIR}/tests")


def run_cmd(args, extra_envs=None, strict=False):
    env = os.environ.copy()
    if extra_envs:
        env = {**env, **extra_envs}

    result = subprocess.run(args,
                            stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT,
                            env=env,
                            universal_newlines=True)
    if strict and result.returncode != 0:
        raise RuntimeError(f"Command {' '.join(args)} failed: {result.stdout}")
    return result.returncode == 0, result.stdout


def ssh(host, user, cmd, strict=False):
    c = ["ssh", f"{user}@{host}",  cmd]
    return run_cmd(c, strict=strict)


def cleanup_microshift(host, user):
    logging.debug("Cleaning MicroShift")
    _, output = ssh(host, user, "echo 1 | sudo -E microshift-cleanup-data --all", strict=True)
    return output


def setup_microshift(host, user):
    logging.debug("Setting up MicroShift")
    config = f"""cat << EOF | sudo tee /etc/microshift/config.yaml
---
apiServer:
  subjectAltNames:
  - {host}
EOF"""
    _, output = ssh(host, user, config, strict=True)
    logging.debug("Starting MicroShift systemd service")
    _, output2 = ssh(host, user, "sudo systemctl enable --now microshift", strict=True)
    logging.debug("MicroShift systemd service started")
    return output + "\n" + output2


def wait_until_microshift_is_fully_functional(host, user):
    logging.debug("Waiting until MicroShift is ready")
    cmd = ("sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh | " +
           "while IFS= read -r line; do printf '%s %s\\n' \"$(date +'%H:%M:%S.%N')\" \"$line\"; done")
    _, output = ssh(host, user, cmd, strict=True)
    return output


def get_kubeconfig(host, user):
    logging.debug("Fetching kubeconfig")
    cmd = f"sudo cat /var/lib/microshift/resources/kubeadmin/{host}/kubeconfig"
    success, output = ssh(host, user, cmd)
    if not success:
        raise RuntimeError("Failed to obtain KUBECONFIG from the remote")
    with tempfile.NamedTemporaryFile(mode="w+", delete=False) as f:
        f.write(output)
        return f.name


def get_cluster_debug_info(host, user):
    copy = ["scp", f"{_SCRIPT_DIR}/cluster-debug-info.sh", f"{user}@{host}:/tmp/cluster-debug-info.sh"]
    run = "sudo KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig /tmp/cluster-debug-info.sh"

    run_cmd(copy, strict=True)
    _, out = ssh(host, user, run, strict=True)
    return out


def wrapper42(xml_parent, func, *args):
    elem = ET.SubElement(xml_parent, func.__name__)
    start = time.time()
    elem.text = func(*args)
    elem.set("elapsed", f"{time.time() - start:.2f}")


def run_test(test, host, user):
    logging.info(f"Running test case {test.name}")
    testcase = ET.Element("testcase", attrib={"name": test.name})
    success = False
    konfig = ""
    try:
        wrapper42(testcase, cleanup_microshift, host, user)
        wrapper42(testcase, setup_microshift, host, user)
        wrapper42(testcase, wait_until_microshift_is_fully_functional, host, user)

        konfig = get_kubeconfig(host, user)
        xml_out = ET.SubElement(testcase, "output")
        start = time.time()
        logging.info(f"Starting {str(test)}")
        success, xml_out.text = run_cmd([str(test)], {"KUBECONFIG": konfig, "USHIFT_IP": host, "USHIFT_USER": user})
        elapsed = time.time() - start
        testcase.set("time", f"{elapsed:.2f}")
        if not success:
            ET.SubElement(testcase, "failure", attrib={"msg": ""})
        logging.info(f"Test case {test.name} - success={success} elapsed={elapsed:.2f}s")

    except Exception as e:
        logging.error(f"Exception happened: '{e}'")
        ET.SubElement(testcase, "failure", attrib={"msg": f"{e}"})
        wrapper42(testcase, get_cluster_debug_info, host, user)
    finally:
        if konfig:
            os.unlink(konfig)
    return testcase, success


def run_tests(tests, host, user):
    testsuite = ET.Element("testsuite", attrib={"tests": f"{len(tests)}"})
    logging.info("Running test suite")
    testsuite_start = time.time()
    failures = 0
    for t in tests:
        test_start = time.time()
        testcase, success = run_test(t, host, user)
        testsuite.append(testcase)
        test_elapsed = time.time() - test_start
        if not success:
            failures = failures + 1
        logging.info(f"Test case and other activities took {test_elapsed:.2f}s")

    testsuite_elapsed = time.time() - testsuite_start
    testsuite.set("time", f"{testsuite_elapsed:.2f}")
    testsuite.set("failures", f"{failures}")

    outdir = os.path.join(os.getenv("ARTIFACT_DIR", "_output"), "junit")
    Path(outdir).mkdir(parents=True, exist_ok=True)
    tree = ET.ElementTree(testsuite)
    tree.write(os.path.join(outdir, "result.xml"))
    logging.info(f"Test suite finished in {testsuite_elapsed:.2f}s: {len(tests)} tests, {failures} failures. \
                 Details stored in {outdir}/result.xml")
    return failures == 0


def list_tests(tests):
    [print(t.name) for t in tests]


def get_test_files(filter):
    filter = f"*{filter}*" if filter else "*.*"
    return list(_TESTS_DIR.glob(filter))


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('--filter', type=str)

    subparsers = parser.add_subparsers(dest='command')

    run = subparsers.add_parser("run", help="Run MicroShift end to end test suite")
    run.add_argument('--host', required=True, type=str, help="IP address of remote MicroShift instance")
    run.add_argument('--user', required=True, type=str, help="User of remote MicroShift instance. Must be configured for passwordless SSH")

    subparsers.add_parser("list", help="List tests from MicroShift end to end test suite")

    return parser.parse_args()


def main():
    args = get_args()
    tests = get_test_files(args.filter)

    if args.command == "list":
        list_tests(tests)
    elif args.command == "run":
        all_passed = run_tests(tests, args.host, args.user)
        sys.exit(0 if all_passed else 1)


if __name__ == "__main__":
    main()
