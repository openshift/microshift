#!/usr/bin/env python

import os
import sys
import json
import logging
import subprocess
from collections import namedtuple
from timeit import default_timer as timer

from git import Repo, PushInfo  # GitPython
from github import GithubIntegration, Github, GithubException  # pygithub
from pathlib import Path

APP_ID_ENV = "APP_ID"  # GitHub App's ID
KEY_ENV = "KEY"  # Path to GitHub App's key
PAT_ENV = "TOKEN"  # Personal Access Token
ORG_ENV = "ORG"
REPO_ENV = "REPO"
AMD64_RELEASE_ENV = "AMD64_RELEASE"
ARM64_RELEASE_ENV = "ARM64_RELEASE"
LVMS_RELEASE_ENV = "LVMS_RELEASE"
JOB_NAME_ENV = "JOB_NAME"
BUILD_ID_ENV = "BUILD_ID"
DRY_RUN_ENV = "DRY_RUN"
BASE_BRANCH_ENV = "BASE_BRANCH"

BOT_REMOTE_NAME = "bot-creds"
REMOTE_ORIGIN = "origin"

# List of reviewers to always requestes review from
REVIEWERS = ["pmtk", "ggiguash"]

# If True, then just log action such as branch push and PR or comment creation
REMOTE_DRY_RUN = False

_extra_msgs = []

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')


RebaseScriptResult = namedtuple("RebaseScriptResult", ["success", "output"])


def try_get_env(var_name, die=True):
    val = os.getenv(var_name)
    if val is None or val == "":
        if die:
            logging.error(f"Could not get environment variable '{var_name}'")
            sys.exit(f"Could not get environment variable '{var_name}'")
        else:
            logging.info(f"Could not get environment variable '{var_name}' - ignoring")
            return ""
    return val


def run_rebase_sh(release_amd64, release_arm64, release_lvms):
    script_dir = os.path.abspath(os.path.dirname(__file__))
    args = [f"{script_dir}/rebase.sh", "to", release_amd64, release_arm64, release_lvms]
    logging.info(f"Running: '{' '.join(args)}'")
    start = timer()
    result = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, universal_newlines=True)
    logging.info(f"Return code: {result.returncode}. Output:\n" +
                 "==================================================\n" +
                 f"{result.stdout}" +
                 "==================================================\n")
    end = timer() - start
    logging.info(f"Script returned code: {result.returncode}. It ran for {end/60:.0f}m{end%60:.0f}s.")
    return RebaseScriptResult(success=result.returncode == 0, output=result.stdout)


def commit_str(commit):
    return f"{commit.hexsha[:8]} - {commit.summary}"


def get_installation_access_token(app_id, key_path, org, repo):
    integration = GithubIntegration(app_id, Path(key_path).read_text())
    app_installation = integration.get_installation(org, repo)
    if app_installation == None:
        sys.exit(f"Failed to get app_installation for {org}/{repo}. Response: {app_installation.raw_data}")
    return integration.get_access_token(app_installation.id).token


def make_sure_rebase_script_created_new_commits_or_exit(git_repo, base_branch):
    if git_repo.active_branch.commit == git_repo.branches[base_branch].commit:
        logging.info(f"There's no new commit on branch {git_repo.active_branch} compared to '{base_branch}' "
                     "meaning that the rebase.sh script didn't create any commits and "
                     "MicroShift is already rebased on top of given release.\n"
                     f"Last commit: {commit_str(git_repo.active_branch.commit)}")
        sys.exit(0)


def rebase_script_made_changes_considered_functional(git_repo, base_branch):
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
    remote_url = f"https://x-access-token:{token}@github.com/{org}/{repo}"
    try:
        remote = git_repo.remote(BOT_REMOTE_NAME)
        remote.set_url(remote_url)
    except ValueError:
        git_repo.create_remote(BOT_REMOTE_NAME, remote_url)

    return git_repo.remote(BOT_REMOTE_NAME)


def try_get_rebase_branch_ref_from_remote(remote, branch_name):
    remote.fetch()
    matching_remote_refs = [ref for ref in remote.refs if BOT_REMOTE_NAME + "/" + branch_name == ref.name]

    if len(matching_remote_refs) == 0:
        logging.info(f"Branch '{branch_name}' does not exist on remote")
        return None

    if len(matching_remote_refs) > 1:
        matching_branches = ", ".join([r.name for r in matching_remote_refs])
        logging.warning(f"Found more than one branch matching '{branch_name}' on remote: {matching_branches}. Taking first one")
        _extra_msgs.append(f"Found more than one branch matching '{branch_name}' on remote: {matching_branches}.")
        return matching_remote_refs[0]

    if len(matching_remote_refs) == 1:
        logging.info(f"Branch '{branch_name}' already exists on remote")
        return matching_remote_refs[0]


def is_local_branch_based_on_newer_base_branch_commit(git_repo, base_branch_name, remote_branch_name, local_branch_name):
    """
    Compares local and remote rebase branches by looking at their start on base branch.
    Returns True if local branch is starts on newer commit and needs to be pushed to remote, otherwise False.
    """
    remote_merge_base = git_repo.merge_base(base_branch_name, remote_branch_name)
    local_merge_base = git_repo.merge_base(base_branch_name, local_branch_name)

    if remote_merge_base[0] == local_merge_base[0]:
        logging.info(f"Remote branch is up to date. Branch-off commit: {commit_str(remote_merge_base[0])}")
        return False
    else:
        logging.info(f"Remote branch is older - it needs updating. "
                     f"Remote branch is on top of {base_branch_name}'s commit: '{commit_str(remote_merge_base[0])}'. "
                     f"Local branch is on top of {base_branch_name}'s commit '{commit_str(local_merge_base[0])}'")
        return True


def try_get_pr(gh_repo, org, base_branch, branch_name):
    prs = gh_repo.get_pulls(base=base_branch, head=f"{org}:{branch_name}", state="all")

    if prs.totalCount == 0:
        logging.info(f"PR for branch {branch_name} does not exist yet on {gh_repo.full_name}")
        return None

    pr = None
    if prs.totalCount > 1:
        pr = prs[0]
        logging.warning(f"Found more than one PR for branch {branch_name} on {gh_repo.full_name} - this is unexpected, continuing with first one of: {[(x.state, x.html_url) for x in prs]}")

    if prs.totalCount == 1:
        pr = prs[0]
        logging.info(f"Found PR #{pr.number} for branch {branch_name} on {gh_repo.full_name}: {pr.html_url}")

    if pr.state == 'closed':
        logging.warning(f"PR #{pr.number} is not open - new PR will be created")
        if pr.is_merged():
            logging.warning(f"PR #{pr.number} for '{branch_name}' branch is already merged but rebase.sh produced results")
            _extra_msgs.append(f"PR #{pr.number} for '{branch_name}' was already merged but rebase.sh produced results")
        else:
            _extra_msgs.append(f"PR #{pr.number} for '{branch_name}' exists already but was closed")
        return None
    return pr


def generate_pr_description(branch_name, amd_tag, arm_tag, prow_job_url, rebase_script_succeded):
    base = (f"amd64: {amd_tag}\n"
            f"arm64: {arm_tag}\n"
            f"prow job: {prow_job_url}\n"
            "\n"
            "/label tide/merge-method-squash\n")
    return (base if rebase_script_succeded
            else "# rebase.sh failed - check committed rebase_sh.log\n\n" + base)


def create_pr(gh_repo, base_branch, branch_name, title, desc):
    if REMOTE_DRY_RUN:
        logging.info(f"[DRY RUN] Create PR: branch='{branch_name}', title='{title}', desc='{desc}'")
        logging.info(f"[DRY RUN] Requesting review from {REVIEWERS}")
        return

    pr = gh_repo.create_pull(title=title, body=desc, base=base_branch, head=branch_name, maintainer_can_modify=True)
    logging.info(f"Created pull request: {pr.html_url}")
    try:
        pr.create_review_request(reviewers=REVIEWERS)
        logging.info(f"Requested review from {REVIEWERS}")
    except GithubException as e:
        logging.info(f"Failed to request review from {REVIEWERS} because: {e}")
    return pr


def update_pr(pr, title, desc):
    if REMOTE_DRY_RUN:
        logging.info(f"[DRY RUN] Update PR #{pr.number}: {title}\n{desc}")
        return

    pr.edit(title=title, body=desc)
    pr.update()  # arm64 release or prow job url might've changed
    logging.info(f"Updated PR #{pr.number}: {title}\n{desc}")


def post_comment(pr, comment=""):
    if len(_extra_msgs) != 0:
        if comment != "":
            comment += "\n\n"
        comment += "Extra messages:\n - " + "\n - ".join(_extra_msgs)

    if comment.strip() != "":
        logging.info(f"Comment to post: {comment}")
        if REMOTE_DRY_RUN:
            logging.info(f"[DRY RUN] Posted a comment")
            return
        issue = pr.as_issue()
        issue.create_comment(comment)
    else:
        logging.info(f"No content for comment")


def push_branch_or_die(remote, branch_name):
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
    parts = release.split(":")
    if len(parts) == 2:
        return parts[1]
    else:
        logging.error(f"Couldn't find tag in '{release}' - using it as is as branch name")
        _extra_msgs.append(f"Couldn't find tag in '{release}' - using it as is as branch name")
        return release


def try_create_prow_job_url():
    job_name = try_get_env(JOB_NAME_ENV, False)
    build_id = try_get_env(BUILD_ID_ENV, False)
    if job_name != "" and build_id != "":
        url = f"https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/{job_name}/{build_id}"
        logging.info(f"Inferred probable prow job url: {url}")
        return url
    else:
        logging.warning(f"Couldn't infer prow job url. Env vars: '{JOB_NAME_ENV}'='{job_name}', '{BUILD_ID_ENV}'='{build_id}'")
        _extra_msgs.append(f"Couldn't infer prow job url. Env vars: '{JOB_NAME_ENV}'='{job_name}', '{BUILD_ID_ENV}'='{build_id}'")
        return "-"


def create_pr_title(branch_name, successful_rebase):
    return branch_name if successful_rebase else f"**FAILURE** {branch_name}"


def get_expected_branch_name(amd, arm):
    amd_tag, arm_tag = get_release_tag(amd), get_release_tag(arm)
    import re
    rx = "(?P<version_stream>.+)-(?P<date>[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{6})"
    match_amd, match_arm = re.match(rx, amd_tag), re.match(rx, arm_tag)
    return f"rebase-{match_amd['version_stream']}_amd64-{match_amd['date']}_arm64-{match_arm['date']}"


def cleanup_branches(gh_repo):
    logging.info("Cleaning up branches for closed PRs")
    rebase_branches = [b for b in gh_repo.get_branches() if b.name.startswith("rebase-4")]
    deleted_branches = []
    for branch in rebase_branches:
        prs = gh_repo.get_pulls(head=f"{gh_repo.owner.login}:{branch.name}", state="all")
        all_prs_are_closed = all([pr.state == "closed" for pr in prs])
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
                except GithubException as e:
                    logging.warning(f"Failed to delete '{ref.ref}' because: {e}")
                    _extra_msgs.append(f"Failed to delete '{ref.ref}' because: {e}")

    if len(deleted_branches) != 0:
        _extra_msgs.append(f"Deleted following branches: " + ", ".join(deleted_branches))


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
    org = try_get_env(ORG_ENV)
    repo = try_get_env(REPO_ENV)
    release_amd = try_get_env(AMD64_RELEASE_ENV)
    release_arm = try_get_env(ARM64_RELEASE_ENV)
    release_lvms = try_get_env(LVMS_RELEASE_ENV)
    base_branch_override = try_get_env(BASE_BRANCH_ENV, die=False)

    global REMOTE_DRY_RUN
    REMOTE_DRY_RUN = False if try_get_env(DRY_RUN_ENV, die=False) == "" else True
    if REMOTE_DRY_RUN:
        logging.info("Dry run mode")

    token = get_token(org, repo)
    gh_repo = Github(token).get_repo(f"{org}/{repo}")
    git_repo = Repo('.')
    base_branch = (git_repo.active_branch.name
        if base_branch_override == ""
        else base_branch_override)

    rebase_result = run_rebase_sh(release_amd, release_arm, release_lvms)
    if rebase_result.success:
        # TODO How can we inform team that rebase job ran successfully just there was nothing new?
        make_sure_rebase_script_created_new_commits_or_exit(git_repo, base_branch)
        if rebase_script_made_changes_considered_functional(git_repo, base_branch):
            logging.info("Detected functional changes made by rebase script - proceeding with creating PR")
        else:
            logging.info("Rebase did not produce any change considered to be functional - quiting")
            sys.exit(0)
    else:
        logging.warning("Rebase script failed - everything will be committed")
        with open('rebase_sh.log', 'w') as writer:
            writer.write(rebase_result.output)
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

    rbranch_does_not_exists = remote_branch == None
    rbranch_exists_and_needs_update = (
        remote_branch != None and
        is_local_branch_based_on_newer_base_branch_commit(git_repo, base_branch, remote_branch.name, rebase_branch_name)
    )
    if rbranch_does_not_exists or rbranch_exists_and_needs_update:
        push_branch_or_die(git_remote, rebase_branch_name)

    prow_job_url = try_create_prow_job_url()
    pr_title = create_pr_title(rebase_branch_name, rebase_result.success)
    desc = generate_pr_description(rebase_branch_name, get_release_tag(release_amd), get_release_tag(release_arm), prow_job_url, rebase_result.success)

    comment = ""
    pr = try_get_pr(gh_repo, org, base_branch, rebase_branch_name)
    if pr == None:
        pr = create_pr(gh_repo, base_branch, rebase_branch_name, pr_title, desc)
    else:
        update_pr(pr, pr_title, desc)
        comment = f"Rebase job updated the branch\n{desc}"

    if base_branch == "main":
        cleanup_branches(gh_repo)
    post_comment(pr, comment)

    git_remote.remove(git_repo, BOT_REMOTE_NAME)
    sys.exit(0 if rebase_result.success else 1)


if __name__ == "__main__":
    main()
