#!/usr/bin/env python3

"""
This Python script automates the process of rebasing a Git branch
of MicroShift repository on a given release.
"""

import json
import logging
import os
import re
import subprocess
import sys
import textwrap
from collections import namedtuple
from pathlib import Path
from timeit import default_timer as timer

from github import GithubException  # pygithub

sys.path.append(str(Path(__file__).resolve().parent / '../pyutils'))
import gitutils  # noqa: E402
import ghutils   # noqa: E402

AMD64_RELEASE_ENV = "AMD64_RELEASE"
ARM64_RELEASE_ENV = "ARM64_RELEASE"
RHOAI_RELEASE_ENV = "RHOAI_RELEASE"
SRIOV_RELEASE_ENV = "SRIOV_RELEASE"
OPM_VERSION_ENV = "OPM_RELEASE"
JOB_NAME_ENV = "JOB_NAME"
BUILD_ID_ENV = "BUILD_ID"
DRY_RUN_ENV = "DRY_RUN"
BASE_BRANCH_ENV = "BASE_BRANCH"

BOT_REMOTE_NAME = "bot-creds"
REMOTE_ORIGIN = "origin"

# If True, then just log action such as branch push and PR or comment creation
REMOTE_DRY_RUN = False

_extra_msgs = []

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')


RebaseScriptResult = namedtuple("RebaseScriptResult", ["success", "output"])


def try_get_env(var_name, die=True):
    """
    Attempts to retrieve the value of an environment variable with the given name, and
    exits the script if the variable is not defined.
    """
    val = os.getenv(var_name)
    if val is None or val == "":
        if die:
            logging.error(f"Could not get environment variable '{var_name}'")
            sys.exit(f"Could not get environment variable '{var_name}'")
        else:
            logging.info(f"Could not get environment variable '{var_name}' - ignoring")
            return ""
    return val


def run_rebase_sh(release_amd64, release_arm64):
    """Run the 'rebase.sh' script with the given release versions and return the script's output."""
    script_dir = os.path.abspath(os.path.dirname(__file__))
    args = [f"{script_dir}/rebase.sh", "to", release_amd64, release_arm64]
    logging.info(f"Running: '{' '.join(args)}'")
    start = timer()
    result = subprocess.run(
        args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True, check=False)
    logging.info(f"Return code: {result.returncode}. Output:\n" +
                 "==================================================\n" +
                 f"{result.stdout}" +
                 "==================================================\n")
    end = timer() - start
    logging.info(f"Script returned code: {result.returncode}. It ran for {end/60:.0f}m{end%60:.0f}s.")
    return RebaseScriptResult(success=result.returncode == 0, output=result.stdout)


def run_rebase_ai_model_serving_sh(release):
    """Run the 'rebase_ai_model_serving.sh' script with the given release version and return the script's output."""
    script_dir = os.path.abspath(os.path.dirname(__file__))
    args = [f"{script_dir}/rebase_ai_model_serving.sh", "to", release]
    env = os.environ.copy()
    env["NO_BRANCH"] = "true"
    logging.info(f"Running: '{' '.join(args)}'")
    start = timer()
    result = subprocess.run(
        args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True, check=False,
        env=env)
    logging.info(f"Return code: {result.returncode}. Output:\n" +
                 "==================================================\n" +
                 f"{result.stdout}" +
                 "==================================================\n")
    end = timer() - start
    logging.info(f"Script returned code: {result.returncode}. It ran for {end/60:.0f}m{end%60:.0f}s.")
    return RebaseScriptResult(success=result.returncode == 0, output=result.stdout)


def run_rebase_sriov_sh(release):
    """Run the 'rebase_sriov.sh' script with the given release version and return the script's output."""
    script_dir = os.path.abspath(os.path.dirname(__file__))
    args = [f"{script_dir}/rebase_sriov.sh", "to", release]
    env = os.environ.copy()
    env["NO_BRANCH"] = "true"
    logging.info(f"Running: '{' '.join(args)}'")
    start = timer()
    result = subprocess.run(
        args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True, check=False,
        env=env)
    logging.info(f"Return code: {result.returncode}. Output:\n" +
                 "==================================================\n" +
                 f"{result.stdout}" +
                 "==================================================\n")
    end = timer() - start
    logging.info(f"Script returned code: {result.returncode}. It ran for {end/60:.0f}m{end%60:.0f}s.")
    return RebaseScriptResult(success=result.returncode == 0, output=result.stdout)


def run_rebase_cert_manager_sh(release):
    """Run the 'rebase_cert_manager.sh' script with the given release version and return the script's output."""
    script_dir = os.path.abspath(os.path.dirname(__file__))
    args = [f"{script_dir}/rebase_cert_manager.sh", "to", release]
    env = os.environ.copy()
    env["NO_BRANCH"] = "true"
    logging.info(f"Running: '{' '.join(args)}'")
    start = timer()
    result = subprocess.run(
        args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True, check=False,
        env=env)
    logging.info(f"Return code: {result.returncode}. Output:\n" +
                 "==================================================\n" +
                 f"{result.stdout}" +
                 "==================================================\n")
    end = timer() - start
    logging.info(f"Script returned code: {result.returncode}. It ran for {end/60:.0f}m{end%60:.0f}s.")
    return RebaseScriptResult(success=result.returncode == 0, output=result.stdout)


def make_sure_rebase_script_created_new_commits_or_exit(git_repo, base_branch):
    """Exit the script if the 'rebase.sh' script did not create any new commits."""
    if git_repo.active_branch.commit == git_repo.branches[base_branch].commit:
        logging.info(f"There's no new commit on branch {git_repo.active_branch} compared to '{base_branch}' "
                     "meaning that the rebase.sh script didn't create any commits and "
                     "MicroShift is already rebased on top of given release.\n"
                     f"Last commit: {gitutils.commit_str(git_repo.active_branch.commit)}")
        sys.exit(0)


def rebase_script_made_changes_considered_functional(git_repo, base_branch):
    """
    Returns True if the changes made by the 'rebase.sh' script are
    considered functional, False otherwise.
    """
    logging.info(f"Deciding if PR should be created by diffing against {base_branch} branch")
    diffs = git_repo.active_branch.commit.diff(base_branch)
    logging.info(f"Following files changed: {[ d.a_path for d in diffs ]}")

    for d in diffs:
        if 'scripts/auto-rebase/' in d.a_path:
            logging.info(f" - {d.a_path} - ignoring")
            continue

        if "assets/release/release-" in d.a_path:
            old_images = set(json.loads(d.a_blob.data_stream.read())['images'].items())
            new_images = set(json.loads(d.b_blob.data_stream.read())['images'].items())
            diff = old_images ^ new_images
            if not diff:
                logging.info(f" - {d.a_path} - images did not change - ignoring")
                continue
            logging.info(f" - {d.a_path} - images changed")
            return True

        logging.info(f" - File {d.a_path} is considered functional")
        return True

    return False


def generate_pr_description(amd_tag, arm_tag, prow_job_url, rebase_script_succeded):  # pylint: disable=unused-argument
    """
    Returns a string that represents the body of a pull request (PR) description.
    Note: This function expects that there is a "scripts/auto-rebase/changelog.txt" file present.
    """
    try:
        with open("scripts/auto-rebase/changelog.txt", mode="r", encoding='utf-8') as file:
            changelog = file.read()
    except Exception as err:
        logging.warning(f"Unable to read changelog file: {err}")
        changelog = ""

    # The GitHub API has a length limit for commit messages. It is
    # longer than the limit imposed here, but this limit seems like
    # the maximum that it would be reasonable to expect someone to
    # actually try to read.
    if len(changelog) > 5000:
        logging.warning("Truncating changelog from %d characters to 5000.", len(changelog))
        logging.info(f"Old changelog:\n{changelog}")
        changelog = f'{changelog[:5000]}\n\nThe change list was truncated. See scripts/auto-rebase/changelog.txt in the PR for the full details.'
        logging.info(f"New changelog:\n{changelog}")

    template = textwrap.dedent("""
    amd64: {amd_tag}
    arm64: {arm_tag}
    prow job: {prow_job_url}

    {changelog}

    /label tide/merge-method-squash
    /label backport-risk-assessed
    /label jira/valid-bug
    """)
    base = template.format(**locals())
    return (base if rebase_script_succeded
            else "# rebase.sh failed - check committed rebase_sh.log\n\n" + base)


def get_release_tag(release):
    """
    Given a release string in the format "abc:xyz", returns the "xyz" portion of the string.
    If the string is not in that format, logs an error and returns the original string.
    """
    parts = release.split(":")
    if len(parts) == 2:
        return parts[1]

    logging.error(f"Couldn't find tag in '{release}' - using it as is as branch name")
    _extra_msgs.append(f"Couldn't find tag in '{release}' - using it as is as branch name")
    return release


def try_create_prow_job_url():
    """
    Attempts to infer the URL for a Prow job based on JOB_NAME_ENV and BUILD_ID_ENV environment variables.
    If successful, logs the URL and returns it. Otherwise, logs a warning and returns "-".
    """
    job_name = try_get_env(JOB_NAME_ENV, False)
    build_id = try_get_env(BUILD_ID_ENV, False)
    if job_name != "" and build_id != "":
        url = f"https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/{job_name}/{build_id}"
        logging.info(f"Inferred probable prow job url: {url}")
        return url

    logging.warning("Couldn't infer prow job url. " +
                    f"Env vars: '{JOB_NAME_ENV}'='{job_name}', '{BUILD_ID_ENV}'='{build_id}'")
    _extra_msgs.append("Couldn't infer prow job url. " +
                       f"Env vars: '{JOB_NAME_ENV}'='{job_name}', '{BUILD_ID_ENV}'='{build_id}'")
    return "-"


def create_pr_title(branch_name, successful_rebase):
    """
    Given a branch name, creates a pull request title. If the rebase was successful,
    the title is just the branch name, else "**FAILURE**" followed by the branch name.
    """
    return f"NO-ISSUE: {branch_name}" if successful_rebase else f"**FAILURE** {branch_name}"


def get_expected_branch_name(amd, arm):
    """
    Given amd and arm release strings, constructs a branch name in the format
    "rebase-{version_stream}_amd64-{date}_arm64-{date}".
    """
    amd_tag, arm_tag = get_release_tag(amd), get_release_tag(arm)
    rxp = "(?P<version_stream>.+)-(?P<date>[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{6})"
    match_amd, match_arm = re.match(rxp, amd_tag), re.match(rxp, arm_tag)
    return f"rebase-{match_amd['version_stream']}_amd64-{match_amd['date']}_arm64-{match_arm['date']}"


def cleanup_branches(gh_repo):
    """
    Deletes branches with names in the format "rebase-4*" that are
    associated with closed pull requests for a given github repo.
    """
    logging.info("Cleaning up branches for closed PRs")
    rebase_branches = [b for b in gh_repo.get_branches() if b.name.startswith("rebase-")]
    deleted_branches = []
    for branch in rebase_branches:
        prs = gh_repo.get_pulls(head=f"{gh_repo.owner.login}:{branch.name}", state="all")
        all_prs_are_closed = all(pr.state == "closed" for pr in prs)
        logging.info(f"'{branch.name}' is referenced in following PRs: " + ", ".join([f"#{pr.number} ({pr.state})" for pr in prs]))
        if all_prs_are_closed:
            ref = gh_repo.get_git_ref(f"heads/{branch.name}")
            if REMOTE_DRY_RUN:
                logging.info(f"[DRY RUN] Delete '{ref.ref}'")
                deleted_branches.append(branch.name)
            else:
                try:
                    ref.delete()
                    logging.info(f"Deleted '{ref.ref}'")
                    deleted_branches.append(branch.name)
                except GithubException as err:
                    logging.warning(f"Failed to delete '{ref.ref}' because: {err}")
                    _extra_msgs.append(f"Failed to delete '{ref.ref}' because: {err}")

    if len(deleted_branches) != 0:
        _extra_msgs.append("Deleted following branches: " + ", ".join(deleted_branches))


def main():
    """
    The main function of the script. Reads environment variables, retrieves the necessary
    information from GitHub, performs a rebase, creates a pull request, and cleans up old branches.
    """
    release_amd = try_get_env(AMD64_RELEASE_ENV)
    release_arm = try_get_env(ARM64_RELEASE_ENV)
    rhoai_release = try_get_env(RHOAI_RELEASE_ENV)
    sriov_release = try_get_env(SRIOV_RELEASE_ENV)
    opm_version = try_get_env(OPM_VERSION_ENV)
    base_branch_override = try_get_env(BASE_BRANCH_ENV, die=False)

    global REMOTE_DRY_RUN
    REMOTE_DRY_RUN = try_get_env(DRY_RUN_ENV, die=False) != ""
    if REMOTE_DRY_RUN:
        logging.info("Dry run mode")

    g = gitutils.GitUtils(dry_run=REMOTE_DRY_RUN)
    base_branch = (
        g.git_repo.active_branch.name
        if base_branch_override == ""
        else base_branch_override
    )

    rebase_result = run_rebase_sh(release_amd, release_arm)
    ai_rebase_result = run_rebase_ai_model_serving_sh(rhoai_release)
    sriov_rebase_result = run_rebase_sriov_sh(sriov_release)
    cert_manager_rebase_result = run_rebase_cert_manager_sh(opm_version)

    rebases_succeeded = all([rebase_result.success, ai_rebase_result.success, sriov_rebase_result.success, cert_manager_rebase_result.success])

    if rebases_succeeded:
        # TODO How can we inform team that rebase job ran successfully just there was nothing new?
        make_sure_rebase_script_created_new_commits_or_exit(g.git_repo, base_branch)
        if rebase_script_made_changes_considered_functional(g.git_repo, base_branch):
            logging.info("Detected functional changes made by rebase script - proceeding with creating PR")
        else:
            logging.info("Rebase did not produce any change considered to be functional - quitting")
            sys.exit(0)
    else:
        logging.warning("Rebase script failed - everything will be committed")
        with open('rebase.log', mode='w', encoding='utf-8') as writer:
            output = ("rebase.sh:\n" +
                      f"{rebase_result.output}" +
                      "==================================================\n" +
                      "rebase_ai_model_serving.sh:\n" +
                      f"{ai_rebase_result.output}")
            writer.write(output)
        if g.git_repo.active_branch.name == base_branch:
            # rebase.sh didn't reach the step that would create a branch
            # so script needs to create it
            g.checkout_branch(get_expected_branch_name(release_amd, release_arm))
        g.add_files_to_staging_area(all=True)
        g.commit("rebase.sh failure artifacts")

    gh = ghutils.GithubUtils(dry_run=REMOTE_DRY_RUN)

    rebase_branch_name = g.git_repo.active_branch.name
    adjusted_base_branch = "main" if gh.is_branch_under_active_development(base_branch) else base_branch
    logging.info(f"Adjusted base branch: {adjusted_base_branch}")

    g.setup_remote_with_token(gh.token, gh.org, gh.repo)
    remote_branch = g.get_remote_branch(rebase_branch_name)  # {BOT_REMOTE_NAME}/{rebase_branch_name}

    rbranch_does_not_exists = remote_branch is None
    rbranch_exists_and_needs_update = (
        remote_branch is not None and
        g.is_local_branch_based_on_newer_base_branch_commit(base_branch, remote_branch.name, rebase_branch_name)
    )
    if rbranch_does_not_exists or rbranch_exists_and_needs_update:
        g.push(rebase_branch_name)

    prow_job_url = try_create_prow_job_url()
    pr_title = create_pr_title(rebase_branch_name, rebase_result.success)
    desc = generate_pr_description(get_release_tag(release_amd), get_release_tag(release_arm), prow_job_url, rebase_result.success)

    comment = ""
    pull_req = gh.get_existing_pr_for_a_branch(adjusted_base_branch, rebase_branch_name)
    if pull_req is None:
        pull_req = gh.create_pr(adjusted_base_branch, rebase_branch_name, pr_title, desc)
    else:
        gh.update_pr(pull_req, pr_title, desc)
        comment = f"Rebase job updated the branch\n{desc}"

    if adjusted_base_branch == "main":
        cleanup_branches(gh.gh_repo)

    gh.post_comment(pull_req, comment, _extra_msgs)

    g.remove_remote_with_token()
    sys.exit(0 if rebases_succeeded else 1)


if __name__ == "__main__":
    main()
