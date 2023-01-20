#!/usr/bin/env python

import os
import sys
import re
import json
import fnmatch
from enum import Enum
from pathlib import Path

REBASE_SCRIPT = "./scripts/auto-rebase/rebase.sh"
RX_CP = "cp (-r )?(?P<src>\S+) (?P<dst>\S+)"
RX_ENVSUBST = "envsubst ?< ?(?P<src>\S+) ?> ?(?P<dst>\S+)"

UPDATE_MANIFESTS_START = "\nupdate_manifests() {\n"
UPDATE_MANIFESTS_END = "\n}\n"

IGNORES = [
    "assets/embed.go",  # embeds assets into the binary
    "assets/release/*",  # part of download_release() function
    "assets/components/ovn/**",  # TODO: manifests are present in cluster-network-operator repository, we should start including them (what about configure-ovs.sh?)
    "assets/components/lvms/**",  # TODO: remove when rebase automation is complete for topolvm
    "assets/core/priority-class-openshift-user-critical.yaml",  # PriorityClass required by `oc debug` (file manually introduced from cluster-config-operator)
    "assets/core/securityv1-local-apiservice.yaml",
    "assets/version/microshift-version.yaml"  # empty template processed at runtime
]

DEBUG = False


class MatchResult(Enum):
    OK = 1
    NOK = 2
    IGNORED = 3
    FAILURE = 4


def get_asset_processing_commands():
    """
    Extracts contents of update_manifests() function in rebase.sh.
    Only lines containing following commands are returned: cp, git restore, envsubst.
    """
    def pred(line):
        return (line.find("cp ") != -1 or
                line.find("git restore") != -1 or
                line.find("envsubst <") != -1)

    with open(REBASE_SCRIPT, "r") as f:
        contents = f.read()
        start = contents.find(UPDATE_MANIFESTS_START)
        end = contents.find(UPDATE_MANIFESTS_END, start) + len(UPDATE_MANIFESTS_END)
        return [line for line in contents[start:end].splitlines() if pred(line)]


def is_path_ignored(file):
    """
    Tries to glob-match given filepath with list of ignored paths
    """
    return len([ignore for ignore in IGNORES if len(fnmatch.filter([str(file)], ignore)) == 1]) == 1


def try_match_file_to_command(file, cmds):
    d = {}  # debug info

    if is_path_ignored(file):
        d.update({"why": "is_path_ignored", "ignores": IGNORES})
        return MatchResult.IGNORED, d

    filename = os.path.basename(file)
    dir = os.path.dirname(file)
    d.update({"filename": filename, "dir": dir})

    # Get only command lines that reference file's dir path
    relevant_cmds = [c for c in cmds if c.find(dir) != -1]

    # Rationale for reversing found commands:
    # assets/components/openshift-dns/node-resolver/daemonset.yaml is references in 3 lines:
    #  - cp "${STAGING_DIR}/"cluster-dns-operator/assets/node-resolver/* "${REPOROOT}"/assets/components/openshift-dns/node-resolver || true
    #  - git restore "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml.tmpl
    #  - envsubst < "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml.tmpl > "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml
    #
    # assets/components/openshift-dns/dns/configmap.yaml is referenced in 2 lines:
    # - cp "${STAGING_DIR}"/cluster-dns-operator/assets/dns/* "${REPOROOT}"/assets/components/openshift-dns/dns || true
    # - git restore "${REPOROOT}"/assets/components/openshift-dns/dns/configmap.yaml
    #
    # In both cases first command (cp) is used with glob, so it's very permissive,
    # therefore reverse is done to match the asset against more specific commands first
    #
    # Another (more robust) solution would be to sort them explicitly with custom sort function,
    # but since commands must be logically ordered, reversing them should be good enough
    relevant_cmds.reverse()
    d["relevant_cmds_in_rebase.sh"] = relevant_cmds

    for cmd in relevant_cmds:
        cmd = str(cmd).replace("\"", "")

        if "envsubst" in cmd:
            matches = re.search(RX_ENVSUBST, cmd)
            if matches == None:
                d.update({"why": "failed to match line against RX_ENVSUBST", "cmd": cmd, "RX_ENVSUBST": RX_ENVSUBST})
                return MatchResult.FAILURE, d

            if "rx_envsubst_results" not in d:
                d["rx_envsubst_results"] = []

            success = file in matches["dst"]
            d["rx_envsubst_results"].append({"cmd": cmd, "src": matches["src"], "dst": matches["dst"], "success": success})
            if success:
                d.update({"matching_method": "envsubst"})
                return MatchResult.OK, d

        if "git restore" in cmd and file in cmd:
            d.update({"matching_method": "git restore"})
            return MatchResult.OK, d

        m = re.search(RX_CP, cmd)
        if m != None:
            src_file = os.path.basename(m["src"])
            dst_file = os.path.basename(m["dst"])
            dst_dir = m["dst"][m["dst"].find('assets/'):]

            if "cp_regex_results" not in d:
                d["cp_regex_results"] = []

            dst_file_success = dst_file == filename
            src_file_success = src_file == filename and ".yaml" not in m["dst"]
            dst_dir_matches = dst_dir == dir
            glob_success = len(fnmatch.filter([filename], src_file)) == 1
            success = dst_file_success or src_file_success or (dst_dir_matches and glob_success)

            d["cp_regex_results"].append({
                "cmd": cmd,
                "src_file": src_file, "dst_file": dst_file,
                "src_file_success": src_file_success,
                "dst_file_success": dst_file_success,
                "dst_dir_matches": dst_dir_matches,
                "glob_success": glob_success,
                "dst_dir_matches and glob_success": dst_dir_matches and glob_success,
                "success": success})

            if success:
                d.update({"matching_method": "cp"})
                return MatchResult.OK, d

    return MatchResult.NOK, d


def main():
    cmds = get_asset_processing_commands()
    files = [str(y) for y in Path("./assets").glob("**/*.*")]
    noks = {}

    for file in files:
        match_result, debug_info = try_match_file_to_command(file, cmds)
        msg = f"{match_result.name:<10}{file}"
        msg += f": {json.dumps(debug_info)}" if DEBUG else ""
        print(msg)

        if match_result == MatchResult.FAILURE:
            print(f"Error occurred while matching. Debug info: { json.dumps(debug_info) }")
            sys.exit(1)

        if match_result not in [MatchResult.OK, MatchResult.IGNORED]:
            noks[file] = debug_info

    if len(noks) != 0:
        print("\nScript failed to find relevant commands in rebase.sh's update_manifests() function for following files in assets/")
        for file, debug_info in noks.items():
            print(f"- {file}: {json.dumps(debug_info)}")
        sys.exit(1)


if __name__ == "__main__":
    main()
