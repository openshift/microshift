"""Unit tests for pure logic functions in precheck scripts."""

import sys
import os
import unittest

# Add parent directory to path so we can import the modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from precheck_xyz import compute_recommendation, interpret_cves, format_text_short, _build_reason  # noqa: E402
from precheck_ecrc import parse_ecrc_version, format_text as ecrc_format_text  # noqa: E402
from precheck_nightly import classify_gap, format_gap, format_text as nightly_format_text  # noqa: E402


class TestClassifyGap(unittest.TestCase):
    def test_ok_zero(self):
        self.assertEqual(classify_gap(0), "OK")

    def test_ok_boundary(self):
        self.assertEqual(classify_gap(24), "OK")

    def test_ask_art_just_over(self):
        self.assertEqual(classify_gap(24.1), "ASK ART")

    def test_ask_art_large(self):
        self.assertEqual(classify_gap(200), "ASK ART")

    def test_negative(self):
        self.assertEqual(classify_gap(-1), "OK")


class TestFormatGap(unittest.TestCase):
    def test_zero(self):
        self.assertEqual(format_gap(0), "0h")

    def test_hours_only(self):
        self.assertEqual(format_gap(13), "13h")

    def test_days_and_hours(self):
        self.assertEqual(format_gap(75), "3d 3h")

    def test_negative(self):
        self.assertEqual(format_gap(-5), "0h")

    def test_exactly_one_day(self):
        self.assertEqual(format_gap(24), "1d 0h")


class TestParseEcrcVersion(unittest.TestCase):
    def test_ec(self):
        result = parse_ecrc_version("4.22.0-ec.5")
        self.assertEqual(result["type"], "EC")
        self.assertEqual(result["base"], "4.22.0")
        self.assertEqual(result["num"], 5)
        self.assertEqual(result["minor"], "4.22")

    def test_rc(self):
        result = parse_ecrc_version("4.22.0-rc.1")
        self.assertEqual(result["type"], "RC")
        self.assertEqual(result["num"], 1)

    def test_invalid(self):
        self.assertIsNone(parse_ecrc_version("4.22.0"))

    def test_invalid_format(self):
        self.assertIsNone(parse_ecrc_version("not-a-version"))


class TestInterpretCves(unittest.TestCase):
    def test_no_report(self):
        result = interpret_cves(None)
        self.assertEqual(result["impact"], "unknown")

    def test_skipped_report(self):
        result = interpret_cves({"skipped": True, "error": "no VPN"})
        self.assertEqual(result["impact"], "unknown")

    def test_no_cves(self):
        report = {
            "RHBA-2026:12345": {"type": "extras", "cves": {}},
            "RHBA-2026:12346": {"type": "image", "cves": {}},
        }
        result = interpret_cves(report)
        self.assertEqual(result["impact"], "none")

    def test_cve_not_affected(self):
        report = {
            "RHSA-2026:12345": {
                "type": "image",
                "cves": {"CVE-2026-1234": {}},
            }
        }
        result = interpret_cves(report)
        self.assertEqual(result["impact"], "none")

    def test_cve_must_release(self):
        report = {
            "RHSA-2026:12345": {
                "type": "image",
                "cves": {
                    "CVE-2026-1234": {
                        "jira_ticket": {
                            "id": "USHIFT-999",
                            "resolution": "Done-Errata",
                            "status": "Closed",
                        }
                    }
                },
            }
        }
        result = interpret_cves(report)
        self.assertEqual(result["impact"], "must_release")

    def test_cve_needs_review(self):
        report = {
            "RHSA-2026:12345": {
                "type": "image",
                "cves": {
                    "CVE-2026-5678": {
                        "jira_ticket": {
                            "id": "USHIFT-100",
                            "resolution": "",
                            "status": "In Progress",
                        }
                    }
                },
            }
        }
        result = interpret_cves(report)
        self.assertEqual(result["impact"], "needs_review")

    def test_metadata_skipped(self):
        report = {
            "RHBA-2026:99999": {
                "type": "metadata",
                "cves": {
                    "CVE-2026-0000": {
                        "jira_ticket": {
                            "id": "USHIFT-1",
                            "resolution": "Done-Errata",
                            "status": "Closed",
                        }
                    }
                },
            }
        }
        result = interpret_cves(report)
        self.assertEqual(result["impact"], "none")


class TestComputeRecommendation(unittest.TestCase):
    def test_must_release_cve(self):
        evaluation = {
            "cve_impact": {"impact": "must_release", "details": [{"cve": "CVE-2026-1"}]},
            "commits": 5,
            "days_since": 10,
        }
        rec, reason = compute_recommendation(evaluation)
        self.assertEqual(rec, "ASK ART TO CREATE ARTIFACTS")
        self.assertIn("CVE fix", reason)

    def test_90_day_rule(self):
        evaluation = {
            "cve_impact": {"impact": "none"},
            "commits": 3,
            "days_since": 95,
        }
        rec, reason = compute_recommendation(evaluation)
        self.assertEqual(rec, "ASK ART TO CREATE ARTIFACTS")
        self.assertIn("90-day", reason)

    def test_skip_no_commits(self):
        evaluation = {
            "cve_impact": {"impact": "none"},
            "commits": 0,
            "days_since": 10,
        }
        rec, _ = compute_recommendation(evaluation)
        self.assertEqual(rec, "SKIP")

    def test_skip_commits_no_cves(self):
        evaluation = {
            "cve_impact": {"impact": "none"},
            "commits": 10,
            "days_since": 30,
        }
        rec, _ = compute_recommendation(evaluation)
        self.assertEqual(rec, "SKIP")

    def test_needs_review_cve_in_progress(self):
        evaluation = {
            "cve_impact": {"impact": "needs_review"},
            "commits": 5,
            "days_since": 10,
        }
        rec, _ = compute_recommendation(evaluation)
        self.assertEqual(rec, "NEEDS REVIEW")

    def test_needs_review_unknown_advisory(self):
        evaluation = {
            "cve_impact": {"impact": "unknown"},
            "commits": 5,
            "days_since": 10,
        }
        rec, _ = compute_recommendation(evaluation)
        self.assertEqual(rec, "NEEDS REVIEW")

    def test_skip_unknown_no_commits(self):
        evaluation = {
            "cve_impact": {"impact": "unknown"},
            "commits": 0,
            "days_since": 10,
        }
        rec, _ = compute_recommendation(evaluation)
        self.assertEqual(rec, "SKIP")


class TestNightlyFormatText(unittest.TestCase):
    def test_empty_branches(self):
        self.assertEqual(nightly_format_text([]), "No branches to check.")

    def test_ok_branch(self):
        branches = [{
            "stream": "4.21",
            "branch": "release-4.21",
            "status": "OK",
            "ocp_timestamp": "2026-04-10T14:30:00",
            "brew_timestamp": "2026-04-10T12:00:00",
            "gap_display": "2h",
        }]
        result = nightly_format_text(branches)
        self.assertIn("OK", result)
        self.assertIn("release-4.21", result)
        self.assertIn("2h", result)

    def test_eol_branch(self):
        branches = [{
            "stream": "4.14",
            "branch": "release-4.14",
            "status": "EOL",
            "lifecycle_phase": "End of life",
            "end_date": "2025-10-31",
        }]
        result = nightly_format_text(branches)
        self.assertIn("EOL", result)
        self.assertIn("End of life", result)

    def test_error_branch(self):
        branches = [{
            "stream": "4.21",
            "branch": "release-4.21",
            "status": "ERROR",
            "ocp_error": "timeout",
        }]
        result = nightly_format_text(branches)
        self.assertIn("ERROR", result)
        self.assertIn("timeout", result)

    def test_verbose_shows_nvr(self):
        branches = [{
            "stream": "4.21",
            "branch": "release-4.21",
            "status": "OK",
            "ocp_timestamp": "2026-04-10T14:30:00",
            "brew_timestamp": "2026-04-10T12:00:00",
            "ocp_nightly": "4.21.0-0.nightly-2026-04-10-143000",
            "brew_build": "microshift-4.21.0~0.nightly_2026_04_10_120000",
            "gap_display": "2h",
        }]
        result = nightly_format_text(branches, verbose=True)
        self.assertIn("OCP: 4.21.0-0.nightly", result)
        self.assertIn("Brew: microshift-4.21", result)


class TestEcrcFormatText(unittest.TestCase):
    def test_ready(self):
        data = {
            "version": "4.22.0-ec.5",
            "status": "READY",
            "ocp_phase": "Accepted",
            "brew": {"found": True, "build_date": "2026-04-09"},
        }
        result = ecrc_format_text(data)
        self.assertIn("OK", result)
        self.assertIn("4.22.0-ec.5", result)
        self.assertIn("2026-04-09", result)

    def test_ocp_pending(self):
        data = {
            "version": "4.22.0-ec.6",
            "status": "OCP_PENDING",
            "ocp_phase": "Pending",
        }
        result = ecrc_format_text(data)
        self.assertIn("ASK ART", result)
        self.assertIn("Pending", result)

    def test_not_found(self):
        data = {
            "version": "4.22.0-ec.99",
            "status": "NOT_FOUND",
        }
        result = ecrc_format_text(data)
        self.assertIn("not on release controller", result)

    def test_type_mismatch(self):
        data = {
            "type": "EC",
            "actual_type": "RC",
            "type_mismatch": True,
            "version": "4.22.0-rc.1",
        }
        result = ecrc_format_text(data)
        self.assertIn("OK", result)
        self.assertIn("No active EC", result)
        self.assertIn("latest is RC", result)

    def test_verbose_next_versions(self):
        data = {
            "version": "4.22.0-ec.5",
            "status": "READY",
            "ocp_phase": "Accepted",
            "brew": {"found": True, "build_date": "2026-04-09"},
            "next_versions": [
                {"version": "4.22.0-ec.6", "exists": False},
                {"version": "4.22.0-rc.1", "exists": True, "phase": "Accepted"},
            ],
        }
        result = ecrc_format_text(data, verbose=True)
        self.assertIn("Next:", result)
        self.assertIn("4.22.0-rc.1 (Accepted)", result)
        self.assertIn("4.22.0-ec.6 (not found)", result)

    def test_brew_error(self):
        data = {
            "version": "4.22.0-ec.5",
            "status": "RPMS_NOT_BUILT",
            "ocp_phase": "Accepted",
            "brew": {"found": False, "error": "VPN not connected"},
        }
        result = ecrc_format_text(data)
        self.assertIn("VPN not connected", result)

    def test_brew_not_found(self):
        data = {
            "version": "4.22.0-ec.5",
            "status": "RPMS_NOT_BUILT",
            "ocp_phase": "Accepted",
            "brew": {"found": False},
        }
        result = ecrc_format_text(data)
        self.assertIn("not found", result)


class TestXyzFormatTextShort(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(format_text_short([]), "No versions to evaluate.")

    def test_already_released(self):
        evals = [{"version": "4.21.7", "recommendation": "ALREADY RELEASED"}]
        result = format_text_short(evals)
        self.assertIn("ALREADY RELEASED", result)
        self.assertIn("4.21.7", result)
        self.assertNotIn("OCP:", result)

    def test_skip_with_ocp(self):
        evals = [{
            "version": "4.18.37",
            "recommendation": "SKIP",
            "ocp_status": "available",
            "cve_impact": {"impact": "none"},
            "last_released": "4.18.36",
            "days_since": 24,
        }]
        result = format_text_short(evals)
        self.assertIn("SKIP", result)
        self.assertIn("[OCP: available]", result)
        self.assertIn("no CVEs", result)

    def test_eol_lifecycle(self):
        evals = [{
            "version": "4.14.50",
            "recommendation": "SKIP",
            "ocp_status": "available",
            "lifecycle_status": "End of life",
            "lifecycle_end_date": "2025-10-31",
        }]
        result = format_text_short(evals)
        self.assertIn("End of life", result)

    def test_ask_art(self):
        evals = [{
            "version": "4.21.9",
            "recommendation": "ASK ART TO CREATE ARTIFACTS",
            "ocp_status": "not_available",
            "cve_impact": {"impact": "must_release", "details": [{"cve": "CVE-2026-1"}]},
        }]
        result = format_text_short(evals)
        self.assertIn("ASK ART TO CREATE ARTIFACTS", result)
        self.assertIn("[OCP: NOT available]", result)


class TestBuildReason(unittest.TestCase):
    def test_no_data(self):
        self.assertEqual(_build_reason({}), "advisory unknown")

    def test_no_cves_with_last(self):
        result = _build_reason({
            "cve_impact": {"impact": "none"},
            "last_released": "4.21.6",
            "days_since": 30,
        })
        self.assertIn("no CVEs", result)
        self.assertIn("last: 4.21.6 (30d ago)", result)

    def test_advisory_skipped(self):
        result = _build_reason({
            "cve_impact": {"impact": "unknown"},
            "advisory_report": {"skipped": True, "error": "no VPN"},
        })
        self.assertIn("advisory report unavailable", result)


class TestJqlSanitization(unittest.TestCase):
    def test_normal_version_unchanged(self):
        from lib.art_jira import _sanitize_jql_value
        self.assertEqual(_sanitize_jql_value("4.21.8"), "4.21.8")

    def test_special_chars_escaped(self):
        from lib.art_jira import _sanitize_jql_value
        result = _sanitize_jql_value('4.21"')
        # The double quote must be escaped
        self.assertTrue(result.endswith('\\"'))
        self.assertFalse(result.endswith('1"'))

    def test_ecrc_version(self):
        from lib.art_jira import _sanitize_jql_value
        result = _sanitize_jql_value("4.22.0-ec.5")
        self.assertIn("\\-", result)


if __name__ == "__main__":
    unittest.main()
