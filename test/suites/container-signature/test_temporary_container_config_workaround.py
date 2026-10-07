#!/usr/bin/env python3
"""Focused tests for the temporary COS 10 container configuration helper."""

from __future__ import annotations

import copy
import datetime
import importlib.util
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import tempfile
try:
    import tomllib
except ImportError:  # pragma: no cover - exercised by the RHEL 9 controller
    import tomli as tomllib
import unittest
from unittest import mock


HELPER_PATH = (
    Path(__file__).parent
    / "assets"
    / "temporary_container_config_workaround.py"
)
SPEC = importlib.util.spec_from_file_location("container_workaround", HELPER_PATH)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError(f"cannot load helper from {HELPER_PATH}")
helper = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(helper)


class TemporaryContainerConfigWorkaroundTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.write(
            "/usr/share/containers/storage.conf",
            """
[storage]
driver = "overlay"
runroot = "/run/containers/storage"
graphroot = "/var/lib/containers/storage"
""",
        )
        self.write(
            "/usr/share/containers/registries.d/registry.example.yaml",
            "docker:\n  registry.example:\n    sigstore: https://example.test\n",
        )

    def write(self, name: str, content: str) -> Path:
        path = self.root / name.lstrip("/")
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content)
        return path

    def state_dir(self) -> Path:
        return helper.rooted(self.root, helper.DEFAULT_STATE_DIR)

    def state_file(self) -> Path:
        return self.state_dir() / helper.STATE_FILE

    def read_state(self) -> dict:
        return json.loads(self.state_file().read_text())

    def write_state(self, state: dict) -> None:
        self.state_file().write_text(json.dumps(state))

    def signature_path(self, name: str = "registry.example.yaml") -> Path:
        return self.root / "etc/containers/registries.d" / name

    def assert_apply_and_restore_succeeds(self) -> None:
        helper.apply(self.root, self.state_dir())
        helper.restore(self.root, self.state_dir())
        self.assertFalse(self.state_dir().exists())

    def test_runtime_uses_native_or_declared_toml_parser(self) -> None:
        expected = "tomllib" if sys.version_info >= (3, 11) else "tomli"
        self.assertEqual(helper.tomllib.__name__, expected)

    def test_recursive_merge_and_complete_type_round_trip(self) -> None:
        document = {
            "types": {
                "boolean": True,
                "integer": 42,
                "float": 1.25,
                "infinity": float("inf"),
                "not_a_number": float("nan"),
                "negative_not_a_number": float("-nan"),
                "string": (
                    "line\nquote\"snowman \N{SNOWMAN} emoji "
                    "\N{GRINNING FACE} delete \x7f"
                ),
                "date": datetime.date(2026, 10, 7),
                "time": datetime.time(12, 30, 1, 500000),
                "local_datetime": datetime.datetime(2026, 10, 7, 12, 30),
                "offset_datetime": datetime.datetime(
                    2026, 10, 7, 12, 30, tzinfo=datetime.timezone.utc
                ),
                "array": ["first", "second"],
                "inline": {"quoted.\N{GRINNING FACE}.key": "value"},
                "array_of_tables": [{"name": "one"}, {"name": "two"}],
            }
        }
        encoded = helper.serialize_toml(document)
        self.assertTrue(
            helper.toml_semantically_equal(tomllib.loads(encoded.decode()), document)
        )

        merged = {"storage": {"driver": "overlay", "options": {"x": 1}}}
        helper.recursive_merge(
            merged,
            {"storage": {"options": {"x": 2, "array": ["replacement"]}}},
        )
        self.assertEqual(
            merged,
            {
                "storage": {
                    "driver": "overlay",
                    "options": {"x": 2, "array": ["replacement"]},
                }
            },
        )

    def test_rootful_global_order_and_higher_precedence_duplicate(self) -> None:
        self.write(
            "/usr/share/containers/storage.conf.d/20-last.conf",
            '[storage.options]\norder = "share-20"\n',
        )
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/15-rootful.conf",
            '[storage.options]\nrootful = "share-15"\n',
        )
        self.write(
            "/usr/share/containers/storage.conf.d/10-duplicate.conf",
            '[storage.options]\nduplicate = "share"\n',
        )
        self.write(
            "/etc/containers/storage.conf.d/10-duplicate.conf",
            '[storage.options]\nduplicate = "etc"\norder = "etc-10"\n',
        )
        self.write(
            "/etc/containers/storage.rootful.conf.d/10-duplicate.conf",
            '[storage.options]\nduplicate = "etc-rootful"\norder = "etc-rootful-10"\n',
        )
        self.write(
            "/etc/containers/storage.conf.d/05-first.conf",
            '[storage.options]\norder = "etc-05"\n',
        )
        self.write(
            "/home/tester/.config/containers/storage.conf.d/99-rootless.conf",
            '[storage.options]\norder = "rootless-must-be-ignored"\n',
        )

        inputs = helper.select_storage_inputs(self.root)
        self.assertEqual(
            [path.name for path in inputs],
            [
                "storage.conf",
                "05-first.conf",
                "10-duplicate.conf",
                "15-rootful.conf",
                "20-last.conf",
            ],
        )
        merged, _ = helper.merge_storage(inputs)
        self.assertEqual(
            merged["storage"]["options"]["duplicate"], "etc-rootful"
        )
        self.assertEqual(merged["storage"]["options"]["rootful"], "share-15")
        self.assertEqual(merged["storage"]["options"]["order"], "share-20")

    def test_admin_main_and_full_vendor_settings_are_preserved(self) -> None:
        admin_signature = self.write(
            "/etc/containers/registries.d/registry.example.yaml",
            "admin-owned: true\n",
        )
        self.write(
            "/etc/containers/storage.conf",
            """
[storage]
driver = "vfs"
runroot = "/admin/runroot"
graphroot = "/admin/graphroot"
[storage.options]
additionalimagestores = ["/admin/images"]
[storage.options.overlay]
mountopt = "nodev"
""",
        )
        self.write(
            "/usr/share/containers/storage.conf.d/00-storage.conf",
            """
[storage]
driver = "overlay"
""",
        )
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/00-storage-rootful.conf",
            """
[storage.options]
additionalimagestores = ["/usr/lib/containers/storage"]
[storage.options.overlay]
mountopt = "nodev,metacopy=on"
""",
        )
        self.write(
            "/etc/containers/storage.rootful.conf.d/99-admin.conf",
            """
[storage]
driver = "vfs"
""",
        )

        helper.apply(self.root, self.state_dir())
        with (self.root / "etc/containers/storage.conf").open("rb") as stream:
            merged = tomllib.load(stream)
        self.assertEqual(merged["storage"]["driver"], "vfs")
        self.assertEqual(merged["storage"]["runroot"], "/admin/runroot")
        self.assertEqual(merged["storage"]["graphroot"], "/admin/graphroot")
        self.assertEqual(
            merged["storage"]["options"]["additionalimagestores"],
            ["/usr/lib/containers/storage"],
        )
        self.assertEqual(
            merged["storage"]["options"]["overlay"]["mountopt"],
            "nodev,metacopy=on",
        )
        helper.restore(self.root, self.state_dir())
        with (self.root / "etc/containers/storage.conf").open("rb") as stream:
            restored = tomllib.load(stream)
        self.assertEqual(restored["storage"]["driver"], "vfs")
        self.assertEqual(restored["storage"]["graphroot"], "/admin/graphroot")
        self.assertEqual(admin_signature.read_text(), "admin-owned: true\n")

    def test_malformed_input_is_rejected_before_configuration_write(self) -> None:
        self.write(
            "/etc/containers/storage.conf.d/50-broken.conf",
            "[storage\ninvalid = true\n",
        )
        with self.assertRaisesRegex(helper.WorkaroundError, "cannot parse TOML"):
            helper.apply(self.root, self.state_dir())
        self.assertFalse((self.root / "etc/containers/storage.conf").exists())

    def test_apply_is_idempotent_and_restore_tracks_original_absence(self) -> None:
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/00-storage-rootful.conf",
            """
[storage.options]
additionalimagestores = ["/usr/lib/containers/storage"]
[storage.options.overlay]
mountopt = "nodev,metacopy=on"
""",
        )
        helper.apply(self.root, self.state_dir())
        storage = self.root / "etc/containers/storage.conf"
        first = storage.read_bytes()
        helper.apply(self.root, self.state_dir())
        self.assertEqual(storage.read_bytes(), first)
        self.assertTrue(
            (self.root / "etc/containers/registries.d/registry.example.yaml").is_file()
        )

        helper.restore(self.root, self.state_dir())
        self.assertFalse(storage.exists())
        self.assertFalse(
            (self.root / "etc/containers/registries.d/registry.example.yaml").exists()
        )
        self.assertFalse(self.state_dir().exists())

    def test_restore_preserves_signature_directory_original_state(self) -> None:
        admin_dir = self.root / "etc/containers/registries.d"
        cases = ("absent", "preexisting-empty", "preexisting-populated")
        for case in cases:
            with self.subTest(case=case):
                if self.state_dir().exists():
                    helper.restore(self.root, self.state_dir())
                if admin_dir.exists():
                    for child in admin_dir.iterdir():
                        child.unlink()
                    admin_dir.rmdir()

                external = admin_dir / "external.yaml"
                if case != "absent":
                    admin_dir.mkdir(parents=True)
                if case == "preexisting-populated":
                    external.write_text("external: true\n")

                helper.apply(self.root, self.state_dir())
                helper.restore(self.root, self.state_dir())

                if case == "absent":
                    self.assertFalse(admin_dir.exists())
                else:
                    self.assertTrue(admin_dir.is_dir())
                if case == "preexisting-populated":
                    self.assertEqual(external.read_text(), "external: true\n")

    def test_restore_preserves_external_file_in_helper_created_directory(self) -> None:
        admin_dir = self.root / "etc/containers/registries.d"
        helper.apply(self.root, self.state_dir())
        external = admin_dir / "external.yaml"
        external.write_text("external: true\n")

        helper.restore(self.root, self.state_dir())

        self.assertTrue(admin_dir.is_dir())
        self.assertEqual(external.read_text(), "external: true\n")
        self.assertFalse(self.signature_path().exists())

    def test_restore_failure_retains_recovery_state(self) -> None:
        helper.apply(self.root, self.state_dir())
        signature = (
            self.root / "etc/containers/registries.d/registry.example.yaml"
        )
        signature.unlink()
        signature.symlink_to("/unexpected-admin-file")

        with self.assertRaisesRegex(helper.WorkaroundError, "Recovery data retained"):
            helper.restore(self.root, self.state_dir())
        self.assertTrue(self.state_dir().is_dir())
        self.assertTrue((self.state_dir() / helper.STATE_FILE).is_file())

    def test_apply_failure_rolls_back_and_retains_recovery_copy(self) -> None:
        with mock.patch.object(
            helper, "atomic_create", side_effect=OSError("injected copy failure")
        ):
            with self.assertRaisesRegex(helper.WorkaroundError, "was rolled back"):
                helper.apply(self.root, self.state_dir())

        self.assertFalse((self.root / "etc/containers/storage.conf").exists())
        self.assertTrue(self.state_dir().is_dir())
        self.assertTrue((self.state_dir() / helper.STATE_FILE).is_file())

    def test_apply_and_rollback_failure_retains_live_recovery_copy(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            """
[storage]
driver = "vfs"
runroot = "/admin/runroot"
graphroot = "/admin/graphroot"
""",
        )
        original_storage = storage.read_bytes()
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/50-test.conf",
            '[storage.options]\nadditionalimagestores = ["/vendor/images"]\n',
        )
        with mock.patch.object(
            helper, "atomic_create", side_effect=OSError("injected apply failure")
        ):
            with mock.patch.object(
                helper, "restore", side_effect=OSError("injected rollback failure")
            ):
                with self.assertRaisesRegex(
                    helper.WorkaroundError, "Rollback also failed"
                ):
                    helper.apply(self.root, self.state_dir())

        self.assertTrue(storage.is_file())
        self.assertNotEqual(storage.read_bytes(), original_storage)
        self.assertTrue(self.state_dir().is_dir())
        self.assertTrue((self.state_dir() / helper.STATE_FILE).is_file())

        helper.restore(self.root, self.state_dir(), remove_state=False)
        self.assertEqual(storage.read_bytes(), original_storage)
        self.assertTrue((self.state_dir() / helper.STATE_FILE).is_file())

    def test_baseline_classifier_accepts_only_pass_or_exact_signature(self) -> None:
        cause = (
            "SignatureValidationFailed: Source image rejected: "
            "A signature was required, but no signature exists"
        )
        self.assertEqual(
            helper.classify_baseline_result(0, "image-reference", "", cause),
            "pass",
        )
        self.assertEqual(
            helper.classify_baseline_result(
                1,
                "",
                cause,
                cause,
            ),
            "signature_failure",
        )
        wrapped_failures = (
            "FATA[0001] pulling image: rpc error: code = Unknown desc = "
            f"verifying signatures: {cause}",
            "ERRO[0001] pulling image: rpc error: code = Unknown desc = "
            f"{cause}",
            'time="2026-10-07T12:00:00Z" level=error msg="pulling image: '
            "rpc error: code = Unknown desc = verifying signatures: "
            f'{cause}"',
        )
        for output in wrapped_failures:
            with self.subTest(output=output):
                self.assertEqual(
                    helper.classify_baseline_result(1, "", output, cause),
                    "signature_failure",
                )
        rejected = {
            "empty success": (0, "", ""),
            "network": (1, "", "dial tcp: connection refused"),
            "authentication": (1, "", "unauthorized: authentication required"),
            "mixed multiline": (1, "", f"{cause}\ndial tcp: timeout"),
            "network same line": (1, "", f"dial tcp timeout before {cause}"),
            "authentication same line": (1, "", f"unauthorized: {cause}"),
            "fatal same line": (1, "", f"fatal unrelated before {cause}"),
            "mixed known wrapper": (
                1,
                "",
                "FATA[0001] pulling image: rpc error: code = Unknown desc = "
                f"dial tcp timeout before {cause}",
            ),
            "unrelated fatal": (2, "", "fatal: runtime unavailable"),
        }
        for name, (return_code, stdout, stderr) in rejected.items():
            with self.subTest(name=name):
                with self.assertRaises(helper.WorkaroundError):
                    helper.classify_baseline_result(
                        return_code, stdout, stderr, cause
                    )

    def test_actual_cos10_signature_failures_are_strictly_classified(self) -> None:
        cause = (
            "SignatureValidationFailed: Source image rejected: "
            "A signature was required, but no signature exists"
        )
        fixture_path = (
            Path(__file__).parent
            / "fixtures"
            / "cos10_crictl_signature_failures.json"
        )
        fixtures = json.loads(fixture_path.read_text())
        self.assertEqual(
            {fixture["architecture"] for fixture in fixtures},
            {"x86_64", "aarch64"},
        )

        for fixture in fixtures:
            with self.subTest(architecture=fixture["architecture"]):
                self.assertEqual(
                    helper.classify_baseline_result(
                        fixture["rc"],
                        fixture["stdout"],
                        fixture["stderr"],
                        cause,
                    ),
                    "signature_failure",
                )
                with self.assertRaises(helper.WorkaroundError):
                    helper.classify_baseline_result(
                        fixture["rc"],
                        fixture["stdout"],
                        fixture["stderr"] + "\ndial tcp: connection refused",
                        cause,
                    )
                mixed_same_line = fixture["stderr"].replace(
                    "not an OCI artifact\"",
                    "not an OCI artifact; unauthorized: authentication required\"",
                    1,
                )
                with self.assertRaises(helper.WorkaroundError):
                    helper.classify_baseline_result(
                        fixture["rc"],
                        fixture["stdout"],
                        mixed_same_line,
                        cause,
                    )

    def test_policy_backup_copy_failure_cannot_unlock_policy_mutation(self) -> None:
        policy = self.write("/etc/containers/policy.json", "original policy\n")
        backup_dir = self.root / "backup"
        backup_dir.mkdir()
        command = helper.build_policy_backup_command(policy, backup_dir)
        fake_bin = self.root / "fake-bin"
        fake_bin.mkdir()
        fake_cp = fake_bin / "cp"
        fake_cp.write_text(
            "#!/bin/sh\n"
            "for last; do :; done\n"
            "printf 'stale partial backup' >\"$last\"\n"
            "exit 73\n"
        )
        fake_cp.chmod(0o755)
        environment = os.environ.copy()
        existing_path = environment["PATH"]
        environment["PATH"] = str(fake_bin) + os.pathsep + existing_path

        result = subprocess.run(
            command,
            shell=True,
            text=True,
            capture_output=True,
            env=environment,
            check=False,
        )
        if result.returncode == 0:
            policy.write_text("test policy\n")

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(policy.read_text(), "original policy\n")
        self.assertFalse((backup_dir / "policy.json").exists())
        self.assertEqual(list(backup_dir.glob("*.partial")), [])

    def test_pre_manifest_creation_failure_is_clean_and_retryable(self) -> None:
        state_dir = self.state_dir()
        real_ensure_directory = helper.ensure_directory

        def fail_after_creation(path, mode=0o755):
            real_ensure_directory(path, mode)
            if path == state_dir:
                raise OSError("injected state directory creation failure")

        with mock.patch.object(
            helper, "ensure_directory", side_effect=fail_after_creation
        ):
            with self.assertRaisesRegex(OSError, "injected state directory"):
                helper.apply(self.root, state_dir)

        self.assertFalse(state_dir.exists())
        self.assertFalse((self.root / "etc/containers/storage.conf").exists())
        self.assertFalse(self.signature_path().exists())
        self.assert_apply_and_restore_succeeds()

    def test_pre_manifest_backup_failure_is_clean_and_retryable(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        original = storage.read_bytes()
        real_atomic_write = helper.atomic_write

        def fail_backup(path, content, **metadata):
            if path.name == helper.STORAGE_BACKUP:
                raise OSError("injected backup failure")
            return real_atomic_write(path, content, **metadata)

        with mock.patch.object(helper, "atomic_write", side_effect=fail_backup):
            with self.assertRaisesRegex(OSError, "injected backup failure"):
                helper.apply(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), original)
        self.assertFalse(self.state_dir().exists())
        self.assertFalse(self.signature_path().exists())
        self.assert_apply_and_restore_succeeds()

    def test_pre_manifest_write_and_fsync_failures_are_clean_and_retryable(self) -> None:
        real_atomic_write = helper.atomic_write
        real_fsync_directory = helper.fsync_directory
        failures = ("write", "fsync")
        for failure in failures:
            with self.subTest(failure=failure):
                if failure == "write":
                    def fail_manifest(path, content, **metadata):
                        if path == self.state_file():
                            real_atomic_write(path, content, **metadata)
                            raise OSError("injected initial manifest write failure")
                        return real_atomic_write(path, content, **metadata)

                    patcher = mock.patch.object(
                        helper, "atomic_write", side_effect=fail_manifest
                    )
                    message = "initial manifest write"
                else:
                    failed = False

                    def fail_manifest_fsync(path):
                        nonlocal failed
                        if path == self.state_dir() and not failed:
                            failed = True
                            raise OSError("injected initial manifest fsync failure")
                        return real_fsync_directory(path)

                    patcher = mock.patch.object(
                        helper, "fsync_directory", side_effect=fail_manifest_fsync
                    )
                    message = "initial manifest fsync"

                with patcher:
                    with self.assertRaisesRegex(OSError, message):
                        helper.apply(self.root, self.state_dir())

                self.assertFalse(self.state_dir().exists())
                self.assertFalse((self.root / "etc/containers/storage.conf").exists())
                self.assertFalse(self.signature_path().exists())
                self.assert_apply_and_restore_succeeds()

    def test_insecure_recovery_metadata_is_rejected_before_live_mutation(self) -> None:
        helper.apply(self.root, self.state_dir())
        storage = self.root / "etc/containers/storage.conf"
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()

        mode_cases = (
            (self.state_dir(), 0o777),
            (self.state_file(), 0o666),
        )
        for path, insecure_mode in mode_cases:
            with self.subTest(path=path.name):
                original_mode = stat.S_IMODE(path.stat().st_mode)
                path.chmod(insecure_mode)
                with self.assertRaisesRegex(helper.WorkaroundError, "permissions"):
                    helper.apply(self.root, self.state_dir())
                self.assertEqual(storage.read_bytes(), applied_storage)
                self.assertEqual(signature.read_bytes(), applied_signature)
                path.chmod(original_mode)

        owner_cases = (
            ("owner", "geteuid", os.geteuid() + 1),
            ("group", "getegid", os.getegid() + 1),
        )
        for name, accessor, unexpected in owner_cases:
            with self.subTest(name=name):
                with mock.patch.object(helper.os, accessor, return_value=unexpected):
                    with self.assertRaisesRegex(helper.WorkaroundError, "owner"):
                        helper.apply(self.root, self.state_dir())
                self.assertEqual(storage.read_bytes(), applied_storage)
                self.assertEqual(signature.read_bytes(), applied_signature)

        helper.restore(self.root, self.state_dir())

    def test_insecure_backup_is_rejected_before_restore_mutation(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        helper.apply(self.root, self.state_dir())
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()
        backup = self.state_dir() / helper.STORAGE_BACKUP
        backup.chmod(0o666)

        with self.assertRaisesRegex(helper.WorkaroundError, "permissions"):
            helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), applied_storage)
        self.assertEqual(signature.read_bytes(), applied_signature)
        backup.chmod(0o600)
        helper.restore(self.root, self.state_dir())

    def test_initial_apply_rejects_insecure_backup_before_live_mutation(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        original_storage = storage.read_bytes()
        real_write_state = helper.write_state

        def tamper_prepared_backup(state_dir, state):
            real_write_state(state_dir, state)
            if state["status"] == "prepared":
                (state_dir / helper.STORAGE_BACKUP).chmod(0o666)

        with mock.patch.object(
            helper, "write_state", side_effect=tamper_prepared_backup
        ):
            with self.assertRaisesRegex(helper.WorkaroundError, "permissions"):
                helper.apply(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), original_storage)
        self.assertFalse(self.signature_path().exists())
        self.assertTrue(self.state_dir().is_dir())

    def test_initial_apply_rejects_corrupt_backup_before_live_mutation(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        original_storage = storage.read_bytes()
        real_write_state = helper.write_state

        def tamper_prepared_backup(state_dir, state):
            real_write_state(state_dir, state)
            if state["status"] == "prepared":
                (state_dir / helper.STORAGE_BACKUP).write_bytes(b"tampered")

        with mock.patch.object(
            helper, "write_state", side_effect=tamper_prepared_backup
        ):
            with self.assertRaisesRegex(helper.WorkaroundError, "integrity check"):
                helper.apply(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), original_storage)
        self.assertFalse(self.signature_path().exists())
        self.assertTrue(self.state_dir().is_dir())

    def test_repeat_apply_revalidates_recovery_backup(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        helper.apply(self.root, self.state_dir())
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()
        (self.state_dir() / helper.STORAGE_BACKUP).write_bytes(b"tampered")

        with self.assertRaisesRegex(helper.WorkaroundError, "integrity check"):
            helper.apply(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), applied_storage)
        self.assertEqual(signature.read_bytes(), applied_signature)
        self.assertTrue(self.state_dir().is_dir())

    def test_symlinked_recovery_path_component_is_rejected(self) -> None:
        helper.apply(self.root, self.state_dir())
        storage = self.root / "etc/containers/storage.conf"
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()
        var_tmp = self.root / "var/tmp"
        actual_var_tmp = self.root / "actual-var-tmp"
        var_tmp.rename(actual_var_tmp)
        var_tmp.symlink_to(actual_var_tmp, target_is_directory=True)

        with self.assertRaisesRegex(helper.WorkaroundError, "symlink"):
            helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), applied_storage)
        self.assertEqual(signature.read_bytes(), applied_signature)

    def test_complete_manifest_schema_rejects_unsafe_nested_values(self) -> None:
        self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        helper.apply(self.root, self.state_dir())
        valid = self.read_state()
        invalid_states = []

        def changed(*path_and_value):
            state = copy.deepcopy(valid)
            *path, value = path_and_value
            target = state
            for key in path[:-1]:
                target = target[key]
            target[path[-1]] = value
            return state

        invalid_states.extend(
            [
                ("version type", changed("version", True)),
                ("status", changed("status", "complete")),
                (
                    "signature directory state",
                    changed("signature_directory_originally_exists", "no"),
                ),
                ("storage path", changed("storage", "path", "/etc/passwd")),
                ("backup name", changed("storage", "backup", "../passwd")),
                ("backup hash", changed("storage", "backup_sha256", "xyz")),
                ("generated hash", changed("storage", "generated_sha256", "0")),
                ("mode", changed("storage", "mode", 0o10000)),
                ("uid", changed("storage", "uid", -1)),
                ("gid type", changed("storage", "gid", True)),
                ("inputs", changed("storage", "inputs", ["/etc/passwd"])),
                (
                    "signature path",
                    changed(
                        "created_signature_files",
                        0,
                        {
                            "path": "/etc/containers/registries.d/../policy.json",
                            "sha256": valid["created_signature_files"][0]["sha256"],
                        },
                    ),
                ),
                (
                    "signature hash",
                    changed(
                        "created_signature_files",
                        0,
                        {
                            "path": valid["created_signature_files"][0]["path"],
                            "sha256": "not-a-hash",
                        },
                    ),
                ),
            ]
        )
        extra_key = copy.deepcopy(valid)
        extra_key["unexpected"] = True
        invalid_states.append(("extra key", extra_key))

        for name, state in invalid_states:
            with self.subTest(name=name):
                with self.assertRaises(helper.WorkaroundError):
                    helper.validate_recovery_state(state, self.root)

        helper.restore(self.root, self.state_dir())

    def test_truncated_state_preflight_leaves_all_live_files_unchanged(self) -> None:
        helper.apply(self.root, self.state_dir())
        storage = self.root / "etc/containers/storage.conf"
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()
        self.state_file().write_text('{"version": 1, "storage":')

        with self.assertRaisesRegex(helper.WorkaroundError, "cannot read"):
            helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), applied_storage)
        self.assertEqual(signature.read_bytes(), applied_signature)
        self.assertTrue(self.state_dir().is_dir())

    def test_malformed_signature_preflight_prevents_storage_restore(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        helper.apply(self.root, self.state_dir())
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()
        state = self.read_state()
        state["created_signature_files"][0]["path"] = "/etc/passwd"
        self.write_state(state)

        with self.assertRaisesRegex(helper.WorkaroundError, "signature target"):
            helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), applied_storage)
        self.assertEqual(signature.read_bytes(), applied_signature)
        self.assertTrue(self.state_dir().is_dir())

    def test_corrupt_backup_preflight_leaves_all_live_files_unchanged(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        helper.apply(self.root, self.state_dir())
        signature = self.signature_path()
        applied_storage = storage.read_bytes()
        applied_signature = signature.read_bytes()
        (self.state_dir() / helper.STORAGE_BACKUP).write_bytes(b"truncated")

        with self.assertRaisesRegex(helper.WorkaroundError, "integrity check"):
            helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), applied_storage)
        self.assertEqual(signature.read_bytes(), applied_signature)
        self.assertTrue(self.state_dir().is_dir())

    def test_restore_preserves_original_storage_mode_uid_and_gid(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        storage.chmod(0o640)
        original_content = storage.read_bytes()
        original_metadata = storage.stat()
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/50-test.conf",
            '[storage.options]\nadditionalimagestores = ["/vendor/images"]\n',
        )
        helper.apply(self.root, self.state_dir())
        storage.write_bytes(original_content)
        storage.chmod(0o666)

        real_atomic_write = helper.atomic_write
        with mock.patch.object(
            helper, "atomic_write", wraps=real_atomic_write
        ) as atomic_write:
            helper.restore(self.root, self.state_dir())

        restore_call = next(
            call
            for call in atomic_write.call_args_list
            if call.args and call.args[0] == storage
        )
        self.assertEqual(restore_call.kwargs["mode"], 0o640)
        self.assertEqual(restore_call.kwargs["uid"], original_metadata.st_uid)
        self.assertEqual(restore_call.kwargs["gid"], original_metadata.st_gid)
        restored_metadata = storage.stat()
        self.assertEqual(stat.S_IMODE(restored_metadata.st_mode), 0o640)
        self.assertEqual(restored_metadata.st_uid, original_metadata.st_uid)
        self.assertEqual(restored_metadata.st_gid, original_metadata.st_gid)
        self.assertFalse(self.state_dir().exists())

    def test_partial_restore_chmod_failure_retains_state_and_is_retryable(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        storage.chmod(0o640)
        original_content = storage.read_bytes()
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/50-test.conf",
            '[storage.options]\nadditionalimagestores = ["/vendor/images"]\n',
        )
        helper.apply(self.root, self.state_dir())
        storage.write_bytes(original_content)
        storage.chmod(0o666)

        with mock.patch.object(
            helper.os,
            "fchmod",
            side_effect=PermissionError("injected restore chmod failure"),
        ):
            with self.assertRaisesRegex(
                helper.WorkaroundError, "injected restore chmod failure"
            ):
                helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), original_content)
        self.assertEqual(stat.S_IMODE(storage.stat().st_mode), 0o666)
        self.assertTrue(self.state_dir().is_dir())
        helper.restore(self.root, self.state_dir())
        self.assertEqual(stat.S_IMODE(storage.stat().st_mode), 0o640)
        self.assertFalse(self.state_dir().exists())

    def test_partial_restore_chown_failure_retains_state_and_is_retryable(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        storage.chmod(0o640)
        original_content = storage.read_bytes()
        self.write(
            "/usr/share/containers/storage.rootful.conf.d/50-test.conf",
            '[storage.options]\nadditionalimagestores = ["/vendor/images"]\n',
        )
        helper.apply(self.root, self.state_dir())
        storage.write_bytes(original_content)
        storage.chmod(0o666)

        with mock.patch.object(
            helper.os,
            "fchown",
            side_effect=PermissionError("injected restore chown failure"),
        ):
            with self.assertRaisesRegex(
                helper.WorkaroundError, "injected restore chown failure"
            ):
                helper.restore(self.root, self.state_dir())

        self.assertEqual(storage.read_bytes(), original_content)
        self.assertEqual(stat.S_IMODE(storage.stat().st_mode), 0o666)
        self.assertTrue(self.state_dir().is_dir())
        helper.restore(self.root, self.state_dir())
        self.assertEqual(stat.S_IMODE(storage.stat().st_mode), 0o640)
        self.assertFalse(self.state_dir().exists())

    def test_storage_target_symlink_is_rejected_on_apply_and_restore(self) -> None:
        storage = self.root / "etc/containers/storage.conf"
        storage.parent.mkdir(parents=True, exist_ok=True)
        unrelated = self.write("/unrelated-storage.conf", "do not change\n")
        storage.symlink_to(unrelated)
        with self.assertRaisesRegex(helper.WorkaroundError, "symlink"):
            helper.apply(self.root, self.state_dir())
        self.assertEqual(unrelated.read_text(), "do not change\n")
        self.assertFalse(self.state_dir().exists())

        storage.unlink()
        helper.apply(self.root, self.state_dir())
        signature = self.signature_path()
        signature_content = signature.read_bytes()
        storage.unlink()
        storage.symlink_to(unrelated)
        with self.assertRaisesRegex(helper.WorkaroundError, "symlink"):
            helper.restore(self.root, self.state_dir())
        self.assertTrue(storage.is_symlink())
        self.assertEqual(unrelated.read_text(), "do not change\n")
        self.assertEqual(signature.read_bytes(), signature_content)
        self.assertTrue(self.state_dir().exists())

    def test_atomic_file_faults_do_not_replace_destination(self) -> None:
        directory = self.root / "atomic"
        directory.mkdir()
        existing = directory / "existing"
        existing.write_bytes(b"original")

        fault_cases = (
            ("permission", "fchmod", helper.atomic_write, existing),
            ("replace", "replace", helper.atomic_write, existing),
            ("link", "link", helper.atomic_create, directory / "link-target"),
            ("fsync", "fsync", helper.atomic_create, directory / "fsync-target"),
        )
        for name, operation, writer, destination in fault_cases:
            with self.subTest(name=name):
                destination.unlink(missing_ok=True)
                if writer is helper.atomic_write:
                    destination.write_bytes(b"original")
                with mock.patch.object(
                    helper.os, operation, side_effect=OSError(f"injected {name}")
                ):
                    with self.assertRaisesRegex(OSError, f"injected {name}"):
                        writer(
                            destination,
                            b"replacement",
                            mode=0o600,
                            uid=os.geteuid(),
                            gid=os.getegid(),
                        )
                if writer is helper.atomic_write:
                    self.assertEqual(destination.read_bytes(), b"original")
                else:
                    self.assertFalse(destination.exists())
                self.assertEqual(list(directory.glob(f".{destination.name}.*.tmp")), [])

    def test_unlink_failure_retains_recoverable_state(self) -> None:
        helper.apply(self.root, self.state_dir())
        signature = self.signature_path()
        storage = self.root / "etc/containers/storage.conf"
        real_unlink = Path.unlink

        def injected_unlink(path, *args, **kwargs):
            if path == signature:
                raise PermissionError("injected unlink failure")
            return real_unlink(path, *args, **kwargs)

        with mock.patch.object(Path, "unlink", new=injected_unlink):
            with self.assertRaisesRegex(
                helper.WorkaroundError, "injected unlink failure"
            ):
                helper.restore(self.root, self.state_dir())

        self.assertFalse(storage.exists())
        self.assertTrue(signature.exists())
        self.assertTrue(self.state_dir().exists())
        helper.restore(self.root, self.state_dir())
        self.assertFalse(signature.exists())
        self.assertFalse(self.state_dir().exists())

    def test_directory_fsync_failure_after_replace_is_rolled_back(self) -> None:
        storage = self.write(
            "/etc/containers/storage.conf",
            '[storage]\ndriver = "vfs"\n',
        )
        original_storage = storage.read_bytes()
        storage_parent = storage.parent
        real_fsync_directory = helper.fsync_directory
        failed = False

        def fail_live_storage_once(path):
            nonlocal failed
            if path == storage_parent and not failed:
                failed = True
                raise OSError("injected directory fsync failure")
            return real_fsync_directory(path)

        with mock.patch.object(
            helper, "fsync_directory", side_effect=fail_live_storage_once
        ):
            with self.assertRaisesRegex(helper.WorkaroundError, "was rolled back"):
                helper.apply(self.root, self.state_dir())

        self.assertTrue(failed)
        self.assertEqual(storage.read_bytes(), original_storage)
        self.assertFalse(self.signature_path().exists())
        self.assertTrue(self.state_dir().exists())

    def test_apply_failure_after_two_signature_copies_rolls_back_all(self) -> None:
        self.write(
            "/usr/share/containers/registries.d/registry.second.yaml",
            "docker: {}\n",
        )
        self.write(
            "/usr/share/containers/registries.d/registry.third.yaml",
            "docker: {}\n",
        )
        real_atomic_create = helper.atomic_create
        created_signatures = []

        def fail_third_signature(path, content, **metadata):
            if path.parent == self.signature_path().parent:
                if len(created_signatures) == 2:
                    raise PermissionError("injected third signature failure")
                created_signatures.append(path)
            return real_atomic_create(path, content, **metadata)

        with mock.patch.object(
            helper, "atomic_create", side_effect=fail_third_signature
        ):
            with self.assertRaisesRegex(helper.WorkaroundError, "was rolled back"):
                helper.apply(self.root, self.state_dir())

        self.assertEqual(len(created_signatures), 2)
        self.assertFalse((self.root / "etc/containers/storage.conf").exists())
        self.assertEqual(
            list((self.root / "etc/containers/registries.d").glob("*.yaml")),
            [],
        )
        self.assertTrue(self.state_dir().is_dir())


if __name__ == "__main__":
    unittest.main()
