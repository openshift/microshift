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

from git import PushInfo, Repo  # GitPython
from github import Github, GithubException, GithubIntegration  # pygithub

APP_ID_ENV = "APP_ID"  # GitHub App's ID
KEY_ENV = "KEY"  # Path to GitHub App's key
PAT_ENV = "TOKEN"  # Personal Access Token
ORG_ENV = "ORG"
REPO_ENV = "REPO"
AMD64_RELEASE_ENV = "AMD64_RELEASE"
ARM64_RELEASE_ENV = "ARM64_RELEASE"
RHOAI_RELEASE_ENV = "RHOAI_RELEASE"
JOB_NAME_ENV = "JOB_NAME"
BUILD_ID_ENV = "BUILD_ID"
DRY_RUN_ENV = "DRY_RUN"
BASE_BRANCH_ENV = "BASE_BRANCH"

BOT_REMOTE_NAME = "bot-creds"
REMOTE_ORIGIN = "origin"

# List of reviewers to always request review from
REVIEWERS = []

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


def commit_str(commit):
    """Returns the first 8 characters of the commit's SHA hash and the commit summary."""
    return f"{commit.hexsha[:8]} - {commit.summary}"


def get_installation_access_token(app_id, key_path, org, repo):
    """Get a installation access token for a GitHub App installation."""
    integration = GithubIntegration(app_id, Path(key_path).read_text(encoding='utf-8'))
    app_installation = integration.get_installation(org, repo)
    if app_installation is None:
        sys.exit(f"Failed to get app_installation for {org}/{repo}. " +
                 f"Response: {app_installation.raw_data}")
    return integration.get_access_token(app_installation.id).token


def make_sure_rebase_script_created_new_commits_or_exit(git_repo, base_branch):
    """Exit the script if the 'rebase.sh' script did not create any new commits."""
    if git_repo.active_branch.commit == git_repo.branches[base_branch].commit:
        logging.info(f"There's no new commit on branch {git_repo.active_branch} compared to '{base_branch}' "
                     "meaning that the rebase.sh script didn't create any commits and "
                     "MicroShift is already rebased on top of given release.\n"
                     f"Last commit: {commit_str(git_repo.active_branch.commit)}")
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


def get_remote_with_token(git_repo, token, org, repo):
    """
    Returns the Git remote for the given repository using
    the provided installation (or personal) access token.
    """
    remote_url = f"https://x-access-token:{token}@github.com/{org}/{repo}"
    try:
        remote = git_repo.remote(BOT_REMOTE_NAME)
        remote.set_url(remote_url)
    except ValueError:
        git_repo.create_remote(BOT_REMOTE_NAME, remote_url)

    return git_repo.remote(BOT_REMOTE_NAME)


def try_get_rebase_branch_ref_from_remote(remote, branch_name):
    """
    Get the reference for the given branch on the specified Git remote,
    otherwise return None if the branch does not exist.
    """
    remote.fetch()
    matching_remote_refs = [ref for ref in remote.refs if BOT_REMOTE_NAME + "/" + branch_name == ref.name]

    if len(matching_remote_refs) == 0:
        logging.info(f"Branch '{branch_name}' does not exist on remote")
        return None

    if len(matching_remote_refs) > 1:
        matching_branches = ", ".join([r.name for r in matching_remote_refs])
        logging.warning(f"Found more than one branch matching '{branch_name}' " +
                        f"on remote: {matching_branches}. Taking first one")
        _extra_msgs.append(f"Found more than one branch matching '{branch_name}' " +
                           f"on remote: {matching_branches}.")
        return matching_remote_refs[0]

    if len(matching_remote_refs) == 1:
        logging.info(f"Branch '{branch_name}' already exists on remote")
        return matching_remote_refs[0]

    return None


def is_local_branch_based_on_newer_base_branch_commit(git_repo, base_branch_name, remote_branch_name, local_branch_name):
    """
    Compares local and remote rebase branches by looking at their start on base branch.
    Returns True if local branch starts on newer commit and needs to be pushed to remote,
    otherwise False.
    """
    remote_merge_base = git_repo.merge_base(base_branch_name, remote_branch_name)
    local_merge_base = git_repo.merge_base(base_branch_name, local_branch_name)

    if remote_merge_base[0] == local_merge_base[0]:
        logging.info("Remote branch is up to date. " +
                     f"Branch-off commit: {commit_str(remote_merge_base[0])}")
        return False

    logging.info(
        f"Remote branch is older - it needs updating. "
        f"Remote branch is on top of {base_branch_name}'s commit: '{commit_str(remote_merge_base[0])}'. "
        f"Local branch is on top of {base_branch_name}'s commit '{commit_str(local_merge_base[0])}'"
    )
    return True


def try_get_pr(gh_repo, org, base_branch, branch_name):
    """
    Try to get a pull request for a branch on a GitHub repository.
    Returns
    - The pull request if it exists and is open, otherwise None.
    - If more than one pull request is found, then the first one will be used.
    """
    prs = gh_repo.get_pulls(base=base_branch, head=f"{org}:{branch_name}", state="all")

    if prs.totalCount == 0:
        logging.info(f"PR for branch {branch_name} does not exist yet on {gh_repo.full_name}")
        return None

    pull_req = None
    if prs.totalCount > 1:
        pull_req = prs[0]
        logging.warning(
            f"Found more than one PR for branch {branch_name} on {gh_repo.full_name} -" +
            f"this is unexpected, continuing with first one of: {[(x.state, x.html_url) for x in prs]}"
        )

    if prs.totalCount == 1:
        pull_req = prs[0]
        logging.info(f"Found PR #{pull_req.number} for branch {branch_name} on {gh_repo.full_name}: {pull_req.html_url}")

    if pull_req.state == 'closed':
        logging.warning(f"PR #{pull_req.number} is not open - new PR will be created")
        if pull_req.is_merged():
            logging.warning(f"PR #{pull_req.number} for '{branch_name}' branch is already merged but rebase.sh produced results")
            _extra_msgs.append(f"PR #{pull_req.number} for '{branch_name}' was already merged but rebase.sh produced results")
        else:
            _extra_msgs.append(f"PR #{pull_req.number} for '{branch_name}' exists already but was closed")
        return None
    return pull_req


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
    /label cherry-pick-approved
    /label backport-risk-assessed
    /label jira/valid-bug
    """)
    base = template.format(**locals())
    return (base if rebase_script_succeded
            else "# rebase.sh failed - check committed rebase_sh.log\n\n" + base)


def create_pr(gh_repo, base_branch, branch_name, title, desc):
    """
    Creates a pull request (and requests reviews) for a given branch on a GitHub repository.
    If the `REMOTE_DRY_RUN` variable is True, it logs the PR creation request without actually creating it.
    """
    if REMOTE_DRY_RUN:
        logging.info(f"[DRY RUN] Create PR: branch='{branch_name}', title='{title}', desc='{desc}'")
        logging.info(f"[DRY RUN] Requesting review from {REVIEWERS}")
        return None

    pull_req = gh_repo.create_pull(
        title=title, body=desc, base=base_branch, head=branch_name, maintainer_can_modify=True)
    logging.info(f"Created pull request: {pull_req.html_url}")
    try:
        pull_req.create_review_request(reviewers=REVIEWERS)
        logging.info(f"Requested review from {REVIEWERS}")
    except GithubException as err:
        logging.info(f"Failed to request review from {REVIEWERS} because: {err}")
    return pull_req


def update_pr(pull_req, title, desc):
    """Updates the title and description of a pull request on a GitHub repository."""
    if REMOTE_DRY_RUN:
        logging.info(f"[DRY RUN] Update PR #{pull_req.number}: {title}\n{desc}")
        return

    pull_req.edit(title=title, body=desc)
    pull_req.update()  # arm64 release or prow job url might've changed
    logging.info(f"Updated PR #{pull_req.number}: {title}\n{desc}")


def post_comment(pull_req, comment=""):
    """
    Posts a comment on a GitHub pull request with
    the contents of the global `_extra_msgs` list.
    """
    if len(_extra_msgs) != 0:
        if comment != "":
            comment += "\n\n"
        comment += "Extra messages:\n - " + "\n - ".join(_extra_msgs)

    if comment.strip() != "":
        logging.info(f"Comment to post: {comment}")
        if REMOTE_DRY_RUN:
            logging.info("[DRY RUN] Posted a comment")
            return
        issue = pull_req.as_issue()
        issue.create_comment(comment)
    else:
        logging.info("No content for comment")


def push_branch_or_die(remote, branch_name):
    """
    Attempts to push a branch to a remote Git repository,
    and terminates the program if the push fails.
    """
    if REMOTE_DRY_RUN:
        logging.info(f"[DRY RUN] git push --force {branch_name}")
        return

    # TODO add retries
    push_result = remote.push(branch_name, force=True)

    if len(push_result) != 1:
        sys.exit(f"Unexpected amount ({len(push_result)}) of items in push_result: {push_result}")
    if push_result[0].flags & PushInfo.ERROR:
        sys.exit(f"Pushing branch failed: {push_result[0].summary}")
    if push_result[0].flags & PushInfo.FORCED_UPDATE:
        logging.info(f"Branch '{branch_name}' existed and was updated (force push)")


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


def get_token(org, repo):
    """
    Returns a token to be used with GitHub API.
    It's either Personal Access Token if TOKEN env is set,
    or Installation Access Token which is intended to be used with GitHub Apps.
    """
    personal_access_token = try_get_env(PAT_ENV, die=False)
    if personal_access_token != "":
        logging.info("Using Personal Access Token to access GitHub API")
        return personal_access_token

    app_id = try_get_env(APP_ID_ENV)
    key_path = try_get_env(KEY_ENV)
    return get_installation_access_token(app_id, key_path, org, repo)


def main():
    """
    The main function of the script. Reads environment variables, retrieves the necessary
    information from GitHub, performs a rebase, creates a pull request, and cleans up old branches.
    """
    org = try_get_env(ORG_ENV)
    repo = try_get_env(REPO_ENV)
    release_amd = try_get_env(AMD64_RELEASE_ENV)
    release_arm = try_get_env(ARM64_RELEASE_ENV)
    rhoai_release = try_get_env(RHOAI_RELEASE_ENV)
    base_branch_override = try_get_env(BASE_BRANCH_ENV, die=False)

    global REMOTE_DRY_RUN
    REMOTE_DRY_RUN = try_get_env(DRY_RUN_ENV, die=False) != ""
    if REMOTE_DRY_RUN:
        logging.info("Dry run mode")

    token = get_token(org, repo)
    gh_repo = Github(token).get_repo(f"{org}/{repo}")
    git_repo = Repo('.')
    base_branch = (
        git_repo.active_branch.name
        if base_branch_override == ""
        else base_branch_override
    )

    rebase_result = run_rebase_sh(release_amd, release_arm)
    ai_rebase_result = run_rebase_ai_model_serving_sh(rhoai_release)

    rebases_succeeded = rebase_result.success and ai_rebase_result.success

    if rebases_succeeded:
        # TODO How can we inform team that rebase job ran successfully just there was nothing new?
        make_sure_rebase_script_created_new_commits_or_exit(git_repo, base_branch)
        if rebase_script_made_changes_considered_functional(git_repo, base_branch):
            logging.info("Detected functional changes made by rebase script - proceeding with creating PR")
        else:
            logging.info("Rebase did not produce any change considered to be functional - quitting")
            sys.exit(0)
    else:
        logging.warning("Rebase script failed - everything will be committed")
        with open('rebase.log', mode='w', encoding='utf-8') as writer:
            output = ("rebase.sh:\n" +
                      f"{rebase_result.stdout}" +
                      "==================================================\n" +
                      "rebase_ai_model_serving.sh:\n" +
                      f"{ai_rebase_result.stdout}")
            writer.write(output)
        if git_repo.active_branch.name == base_branch:
            # rebase.sh didn't reach the step that would create a branch
            # so script needs to create it
            branch = git_repo.create_head(get_expected_branch_name(release_amd, release_arm))
            branch.checkout()
        git_repo.git.add(A=True)
        git_repo.index.commit("rebase.sh failure artifacts")

    rebase_branch_name = git_repo.active_branch.name
    git_remote = get_remote_with_token(git_repo, token, org, repo)
    remote_branch = try_get_rebase_branch_ref_from_remote(git_remote, rebase_branch_name)  # {BOT_REMOTE_NAME}/{rebase_branch_name}

    rbranch_does_not_exists = remote_branch is None
    rbranch_exists_and_needs_update = (
        remote_branch is not None and
        is_local_branch_based_on_newer_base_branch_commit(git_repo, base_branch, remote_branch.name, rebase_branch_name)
    )
    if rbranch_does_not_exists or rbranch_exists_and_needs_update:
        push_branch_or_die(git_remote, rebase_branch_name)

    prow_job_url = try_create_prow_job_url()
    pr_title = create_pr_title(rebase_branch_name, rebase_result.success)
    desc = generate_pr_description(get_release_tag(release_amd), get_release_tag(release_arm), prow_job_url, rebase_result.success)

    comment = ""
    pull_req = try_get_pr(gh_repo, org, base_branch, rebase_branch_name)
    if pull_req is None:
        pull_req = create_pr(gh_repo, base_branch, rebase_branch_name, pr_title, desc)
    else:
        update_pr(pull_req, pr_title, desc)
        comment = f"Rebase job updated the branch\n{desc}"

    if base_branch == "main":
        cleanup_branches(gh_repo)
    post_comment(pull_req, comment)

    git_remote.remove(git_repo, BOT_REMOTE_NAME)
    sys.exit(0 if rebase_result.success else 1)


if __name__ == "__main__":
    main()
