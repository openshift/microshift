"""Local-only tests; no images are pulled or built and no podman commands are run.

Run: python3 -m unittest discover -s test/bin/pyutils/tests -v
"""

import importlib.util
import os
import sys
import types
import unittest
from pathlib import Path
from unittest.mock import patch

PYUTILS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(PYUTILS))

# common.py imports psutil at module level, but it is only used by the
# process helpers that are mocked in these tests
try:
    import psutil  # noqa: F401
except ImportError:
    sys.modules["psutil"] = types.ModuleType("psutil")

import common  # noqa: E402

CENTOS_BIB_DEFAULT = "quay.io/centos-bootc/bootc-image-builder:latest"
RHEL9_BIB = "registry.redhat.io/rhel9/bootc-image-builder:latest"
RHEL10_BIB = "registry.redhat.io/rhel10/bootc-image-builder:latest"
IBC = "ghcr.io/osbuild/image-builder-cli:latest"
# Placeholder references used only to verify the override plumbing
CENTOS9_OVERRIDE = "registry.example.com/test/centos9-bib:override"
CENTOS10_OVERRIDE = "registry.example.com/test/centos10-bib:override"

# Representative image references, mostly taken from the bootc blueprints in
# the repository, covering every alias recognized by get_bib_image
CENTOS9_REFS = (
    "quay.io/centos-bootc/centos-bootc:stream9",
    "localhost/cos9-bootc-source:latest",
    "localhost/cos9-bootc-source-isolated:latest",
    "localhost/cos9-bootc-source-optionals:latest",
)
CENTOS10_REFS = (
    "quay.io/centos-bootc/centos-bootc:stream10",
    "localhost/cos10-bootc-source:latest",
    "localhost/cos10-bootc-source-isolated:latest",
    "localhost/cos10-bootc-source-optionals:latest",
)
RHEL9_REFS = (
    "registry.redhat.io/rhel9/rhel-bootc:9.8",
    "localhost/rhel98-bootc-source-isolated:latest",
    "localhost/rhel98-bootc-brew-lrel-optional:latest",
    "localhost/rhel98-bootc-source-ai-model-serving:latest",
    "localhost/rhel-9-bootc:latest",
)
RHEL10_REFS = (
    "registry.redhat.io/rhel10/rhel-bootc:10.2",
    "localhost/rhel102-bootc-source-isolated:latest",
    "localhost/rhel102-bootc-brew-lrel-optional:latest",
    "localhost/rhel102-bootc-source-ai-model-serving:latest",
)


def load_module(**overrides):
    """Import a fresh copy of build_bootc_images with the given builder overrides.

    The builder images are resolved at import time, so each load reads the
    environment prepared here.
    """
    # Paths are only stored by the module at import time, never accessed
    env = {name: "/nonexistent/build-bootc-images-test" for name in (
        "SCRIPTDIR", "BOOTC_IMAGE_DIR", "BOOTC_ISO_DIR", "IMAGEDIR", "VM_DISK_BASEDIR",
        "CONTAINER_LIST", "LOCAL_REPO", "BASE_REPO", "NEXT_REPO", "BREW_REPO", "HOME", "GOMPLATE")}
    env.update({"UNAME_M": "x86_64", "MIRROR_REGISTRY_URL": "localhost:5000"})
    env.update(overrides)
    with patch.dict(os.environ, env):
        for name in ("BIB_IMAGE_CENTOS9", "BIB_IMAGE_CENTOS10"):
            if name not in overrides:
                os.environ.pop(name, None)
        spec = importlib.util.spec_from_file_location(
            "build_bootc_images_under_test", PYUTILS / "build_bootc_images.py")
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
    return module


class BibImageSelection(unittest.TestCase):
    def assert_selection(self, module, refs, expected):
        for ref in refs:
            with self.subTest(ref=ref):
                self.assertEqual(module.get_bib_image(ref), expected)

    def test_defaults_are_preserved(self):
        module = load_module()
        self.assertEqual(module.BIB_IMAGE_CENTOS9, CENTOS_BIB_DEFAULT)
        self.assertEqual(module.BIB_IMAGE_CENTOS10, CENTOS_BIB_DEFAULT)
        self.assertEqual(module.BIB_IMAGE_RHEL9, RHEL9_BIB)
        self.assertEqual(module.BIB_IMAGE, RHEL10_BIB)
        self.assertEqual(module.IBC_IMAGE, IBC)
        self.assert_selection(module, CENTOS9_REFS, CENTOS_BIB_DEFAULT)
        self.assert_selection(module, CENTOS10_REFS, CENTOS_BIB_DEFAULT)
        self.assert_selection(module, RHEL9_REFS, RHEL9_BIB)
        self.assert_selection(module, RHEL10_REFS, RHEL10_BIB)

    def test_both_centos_overrides(self):
        module = load_module(BIB_IMAGE_CENTOS9=CENTOS9_OVERRIDE, BIB_IMAGE_CENTOS10=CENTOS10_OVERRIDE)
        self.assert_selection(module, CENTOS9_REFS, CENTOS9_OVERRIDE)
        self.assert_selection(module, CENTOS10_REFS, CENTOS10_OVERRIDE)
        self.assert_selection(module, RHEL9_REFS, RHEL9_BIB)
        self.assert_selection(module, RHEL10_REFS, RHEL10_BIB)

    def test_centos9_override_only(self):
        module = load_module(BIB_IMAGE_CENTOS9=CENTOS9_OVERRIDE)
        self.assert_selection(module, CENTOS9_REFS, CENTOS9_OVERRIDE)
        self.assert_selection(module, CENTOS10_REFS, CENTOS_BIB_DEFAULT)
        self.assert_selection(module, RHEL9_REFS, RHEL9_BIB)
        self.assert_selection(module, RHEL10_REFS, RHEL10_BIB)

    def test_empty_overrides_use_default(self):
        module = load_module(BIB_IMAGE_CENTOS9="", BIB_IMAGE_CENTOS10="")
        self.assertEqual(module.BIB_IMAGE_CENTOS9, CENTOS_BIB_DEFAULT)
        self.assertEqual(module.BIB_IMAGE_CENTOS10, CENTOS_BIB_DEFAULT)
        self.assert_selection(module, CENTOS9_REFS, CENTOS_BIB_DEFAULT)
        self.assert_selection(module, CENTOS10_REFS, CENTOS_BIB_DEFAULT)
        self.assertNotIn("", module.get_builder_images())

    def test_centos10_override_only(self):
        module = load_module(BIB_IMAGE_CENTOS10=CENTOS10_OVERRIDE)
        self.assert_selection(module, CENTOS9_REFS, CENTOS_BIB_DEFAULT)
        self.assert_selection(module, CENTOS10_REFS, CENTOS10_OVERRIDE)
        self.assert_selection(module, RHEL9_REFS, RHEL9_BIB)
        self.assert_selection(module, RHEL10_REFS, RHEL10_BIB)


class CleanupAtexit(unittest.TestCase):
    def run_cleanup(self, module, running=None):
        """Run cleanup_atexit and return the queried ancestors and stop commands."""
        running = running or {}
        ancestors = []
        stops = []

        def fake_shell(args, dry_run=False, *_):
            if args[:3] == ["sudo", "podman", "ps"]:
                ancestor = args[args.index("--filter") + 1][len("ancestor="):]
                ancestors.append(ancestor)
                return running.get(ancestor, "")
            if args[:3] == ["sudo", "podman", "stop"]:
                stops.append(args)
                return ""
            self.fail(f"Unexpected command: {args}")

        with patch.object(common, "find_subprocesses", return_value=[]), \
                patch.object(common, "print_msg"), \
                patch.object(common, "run_command_in_shell", side_effect=fake_shell):
            module.cleanup_atexit(dry_run=False)
        return ancestors, stops

    def test_shared_default_queried_once(self):
        module = load_module()
        ancestors, stops = self.run_cleanup(module)
        self.assertEqual(ancestors, [RHEL9_BIB, RHEL10_BIB, CENTOS_BIB_DEFAULT, IBC])
        self.assertEqual(stops, [])

    def test_distinct_overrides_are_queried(self):
        module = load_module(BIB_IMAGE_CENTOS9=CENTOS9_OVERRIDE, BIB_IMAGE_CENTOS10=CENTOS10_OVERRIDE)
        ancestors, _ = self.run_cleanup(module)
        self.assertEqual(ancestors, [RHEL9_BIB, RHEL10_BIB, CENTOS9_OVERRIDE, CENTOS10_OVERRIDE, IBC])

    def test_single_override_keeps_default(self):
        module = load_module(BIB_IMAGE_CENTOS10=CENTOS10_OVERRIDE)
        ancestors, _ = self.run_cleanup(module)
        self.assertEqual(ancestors, [RHEL9_BIB, RHEL10_BIB, CENTOS_BIB_DEFAULT, CENTOS10_OVERRIDE, IBC])

    def test_running_builder_containers_are_stopped(self):
        module = load_module(BIB_IMAGE_CENTOS9=CENTOS9_OVERRIDE, BIB_IMAGE_CENTOS10=CENTOS10_OVERRIDE)
        _, stops = self.run_cleanup(module, running={
            CENTOS9_OVERRIDE: "c9a\nc9b",
            CENTOS10_OVERRIDE: "c10a",
        })
        self.assertEqual(stops, [
            ["sudo", "podman", "stop", "c9a c9b"],
            ["sudo", "podman", "stop", "c10a"],
        ])


if __name__ == "__main__":
    unittest.main()
