"""Pre-run modifier that skips tests by their name.

Tests to skip are specified by Test Name
"""

import os
from robot.api import SuiteVisitor


class SkipTests(SuiteVisitor):

    def __init__(self, list_skip_test):
        self.list_skip_test = list_skip_test

        if self.list_skip_test:
            print("List of tests to be skipped:")
            print(f" - {self.list_skip_test.replace(',', f'{os.linesep} - ')}")
        else:
            print("No tests to be skipped")

    def visit_test(self, test):
        """Set tag to skip tests"""

        # set `robot:skip` tag
        if test.name in self.list_skip_test.split(","):
            test.tags.add("robot:skip")
