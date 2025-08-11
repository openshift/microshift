"""
Following file implements a listener for running SOS Report after failed test case.
See Robot Framework User Guide for more information about listener interfaces:
https://robotframework.org/robotframework/latest/RobotFrameworkUserGuide.html#listener-interface

This listener listens for two events:
- Variable creation - store the variable name if it contains 'namespace' or 'ns'
- Keyword failure - failure of "test-level" keyword triggers SOS report collection

The goal is to collect minimal sos report whenever test fails.
Sos reports collected at the end of the scenario might not contain relevant information for debugging failures
because cluster objects might be already deleted (like suite or test teardown removing namespaces)
or long running Pods (like MicroShift core Pods) might've been restarted couple times and the relevant logs are lost.
"""

import re
from robot import result, running
from robot.libraries.BuiltIn import BuiltIn

suite_namespaces = []
test_namespaces = []
test_started = False


def start_test(data: running.model.TestCase, res: result.model.TestCase):
    global test_started
    test_started = True


def end_test(data: running.model.TestCase, res: result.model.TestCase):
    global test_started
    test_started = False
    test_namespaces.clear()


def start_suite(data: running.model.TestSuite, res: result.model.TestSuite):
    suite_namespaces.clear()


def end_keyword(data: running.model.Keyword, res: result.model.Keyword):
    # If a test case level keyword failed, collect SOS report.
    # If the keyword failed on different level (like inside Wait Until Keyword Succeeds), it's ignored.
    if res.status == "FAIL" and isinstance(res.parent, result.model.TestCase):
        BuiltIn().log("sos-on-failure-listener.py: Failure was detected, collecting SOS report")
        BuiltIn().import_resource('microshift-host.resource')
        cmd = "microshift-sos-report --profiles microshift --plugins ''"
        if len(suite_namespaces) > 0 or len(test_namespaces) > 0:
            cmd += f" --namespaces {','.join(suite_namespaces + test_namespaces)}"
        stdout, _, _ = BuiltIn().run_keyword("Command Execution", cmd)
        m = re.search(r'(\/\S+\/sosreport\S+.tar.xz)', stdout)
        if m:
            BuiltIn().log(f"sos-on-failure-listener.py: SOS report collected to {m.group(1)}")


def start_var(data: running.model.Var, res: result.model.Var):
    # Store variable's name if:
    # - it has SUITE scope
    # - name contains 'namespace' or 'ns' (case insensitive)
    # - it's not already in the list of test namespaces
    # start_var requires RF >= 7.0 and gets information about variables set using VAR keyword.
    if (res.scope == "SUITE"
       and ('namespace' in res.name.lower() or 'ns' in res.name.lower())
       and res.name not in test_namespaces):
        if test_started:
            test_namespaces.append(res.name)
        else:
            suite_namespaces.append(res.name)
