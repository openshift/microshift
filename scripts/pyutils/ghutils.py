#!/usr/bin/env python3

import os
import logging
import sys
from pathlib import Path
from github import GithubIntegration, Github

APP_ID_ENV = "APP_ID"   # GitHub App's ID
KEY_ENV = "KEY"         # Path to GitHub App's key
PAT_ENV = "GH_TOKEN"    # Personal Access Token

ORG_ENV = "ORG"
_DEFAULT_ORG = "openshift"
REPO_ENV = "REPO"
_DEFAULT_REPO = "microshift"


class GithubUtils:
    def __init__(self, dry_run=False):
        self.dry_run = dry_run
        self.org, self.repo = self._get_org_repo_from_env()
        self.token = self._get_gh_token_from_env()
        self.gh_repo = Github(self.token).get_repo(f"{self.org}/{self.repo}")

    def _get_org_repo_from_env(self) -> tuple[str, str]:
        if self.dry_run:
            logging.info(f"[DRY RUN] Using default org and repo: {_DEFAULT_ORG}/{_DEFAULT_REPO}")
            return _DEFAULT_ORG, _DEFAULT_REPO
        return try_get_env(ORG_ENV, default=_DEFAULT_ORG), try_get_env(REPO_ENV, default=_DEFAULT_REPO)

    def _get_gh_token_from_env(self) -> str:
        """
        Returns a token to be used with GitHub API.
        It's either Personal Access Token if TOKEN env is set,
        or Installation Access Token which is intended to be used with GitHub Apps.
        """
        personal_access_token = try_get_env(PAT_ENV)
        if personal_access_token != "":
            logging.info("Using Personal Access Token to access GitHub API")
            return personal_access_token

        app_id = try_get_env(APP_ID_ENV, die=True)
        key_path = try_get_env(KEY_ENV, die=True)
        integration = GithubIntegration(app_id, Path(key_path).read_text(encoding='utf-8'))
        app_installation = integration.get_repo_installation(self.org, self.repo)
        if app_installation is None:
            sys.exit(f"Failed to get app_installation for {self.org}/{self.repo}. " +
                     f"Response: {app_installation.raw_data}")
        return integration.get_access_token(app_installation.id).token

    def is_branch_under_active_development(self, branch):
        """
        Checks title of the issue #1239 in the openshift/microshift repository to check if
        given branch is frozen and thus under active development is happening on main branch.

        It returns True if given branch is the first on the list of frozen branches.
        In such case the target (base) branch of newly created PR should be switch to main.
        """
        if self.dry_run:
            logging.info(f"[DRY RUN] Assuming branch {branch} is under active development")
            return True

        issue = self.gh_repo.get_issue(number=1239)
        title = issue.title
        try:
            branches_part = title.split('|', 1)[1].strip()
            frozen_branches = [x.replace('branch:', '') for x in branches_part.split()]
            if len(frozen_branches) == 0:
                raise Exception(f"Unexpected amount of branch in the Issue 1239 title: {title}")
            # Assuming the first branch name is the release under development right now.
            # No job creating PRs should run against the next release branch.
            return branch == frozen_branches[0]
        except Exception as e:
            raise RuntimeError(f"Failed to parse freeze issue title: {title} ({e})")

    def create_pr(self, base_branch, branch_name, title, desc):
        """Create a Pull Request"""
        if self.dry_run:
            logging.info(f"[DRY RUN] Create PR: {base_branch=} <- {branch_name=}: {title=} {desc=}")
            return None

        if (self.org == _DEFAULT_ORG and self.repo == _DEFAULT_REPO and
                self.is_branch_under_active_development(base_branch)):
            base_branch = "main"

        pull_req = self.gh_repo.create_pull(
            title=title, body=desc, base=base_branch, head=branch_name, maintainer_can_modify=True)
        logging.info(f"Created pull request: {pull_req.html_url}")
        return pull_req

    def update_pr(self, pull_req, title, desc):
        """Updates the title and description of a pull request on a GitHub repository."""
        if self.dry_run:
            logging.info(f"[DRY RUN] Update PR #{pull_req.number}: {title}\n{desc}")
            return

        pull_req.edit(title=title, body=desc)
        pull_req.update()  # arm64 release or prow job url might've changed
        logging.info(f"Updated PR #{pull_req.number}: {title}\n{desc}")

    def post_comment(self, pull_req, comment="", extra_msgs=[]):
        """Post a comment on a Pull Request with the contents of the extra_msgs list."""
        if self.dry_run:
            logging.info(f"[DRY RUN] Post comment: {comment}")
            return

        if len(extra_msgs) != 0:
            if comment != "":
                comment += "\n\n"
            comment += "Extra messages:\n - " + "\n - ".join(extra_msgs)

        if comment.strip() != "":
            logging.info(f"Comment to post: {comment}")
            if self.dry_run:
                logging.info("[DRY RUN] Posted a comment")
                return
            issue = pull_req.as_issue()
            issue.create_comment(comment)
        else:
            logging.info("No content for comment")

    def get_existing_pr_for_a_branch(self, base_branch, branch_name):
        """
        Try to get a pull request for a branch on a GitHub repository.
        Returns
        - The pull request if it exists and is open, otherwise None.
        - If more than one pull request is found, then the first one will be used.
        """
        prs = self.gh_repo.get_pulls(base=base_branch, head=f"{self.org}:{branch_name}", state="all")

        if prs.totalCount == 0:
            logging.info(f"PR for branch {branch_name} does not exist yet on {self.gh_repo.full_name}")
            return None

        pull_req = None
        if prs.totalCount > 1:
            pull_req = prs[0]
            logging.warning(
                f"Found more than one PR for branch {branch_name} on {self.gh_repo.full_name} -" +
                f"this is unexpected, continuing with first one of: {[(x.state, x.html_url) for x in prs]}"
            )

        if prs.totalCount == 1:
            pull_req = prs[0]
            logging.info(f"Found PR #{pull_req.number} for branch {branch_name} on {self.gh_repo.full_name}: {pull_req.html_url}")

        if pull_req.state == 'closed':
            logging.warning(f"PR #{pull_req.number} is not open - new PR will be created")
            if pull_req.is_merged():
                logging.warning(f"PR #{pull_req.number} for '{branch_name}' branch is already merged but rebase.sh produced results")
            else:
                logging.warning(f"PR #{pull_req.number} for '{branch_name}' exists already but was closed")
            return None
        return pull_req


def try_get_env(var_name, default=None, die=False) -> str:
    val = os.getenv(var_name)
    if val is None or val == "":
        if default is not None:
            logging.info(f"'{var_name}' env var is unset, using '{default}'")
            return default
        if die:
            raise Exception(f"Could not get environment variable '{var_name}'")
        else:
            logging.info(f"Could not get environment variable '{var_name}' - ignoring")
            return ""
    return val
