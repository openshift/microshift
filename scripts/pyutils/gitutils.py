#!/usr/bin/env python3

import base64
import logging

from git import Repo  # GitPython
from github import GithubException, InputGitTreeElement

BOT_REMOTE_NAME = "bot-creds"
REMOTE_ORIGIN = "origin"


class GitUtils():
    def __init__(self, dry_run=False):
        self.dry_run = dry_run
        self.git_repo = Repo(".")
        self.remote = None

    def file_changed(self, file_path) -> bool:
        changedFiles = [item.a_path for item in self.git_repo.index.diff(None)]
        return file_path in changedFiles

    def add_files_to_staging_area(self, file_paths=[], all=False):
        if self.dry_run:
            if all:
                logging.info("[DRY RUN] git add -A")
            else:
                logging.info(f"[DRY RUN] git add {file_paths}")
            return

        if all:
            self.git_repo.git.add("-A")
        else:
            self.git_repo.index.add(file_paths)

    def commit(self, message):
        if self.dry_run:
            logging.info(f"[DRY RUN] git commit -m {message}")
            return
        self.git_repo.index.commit(message)

    def checkout_branch(self, branch_name):
        if self.dry_run:
            logging.info(f"[DRY RUN] git checkout -b {branch_name}")
            return
        new_branch = self.git_repo.create_head(branch_name)
        new_branch.checkout()

    def setup_remote_with_token(self, token, org, repo):
        """
        Sets up the Git remote for the given repository using
        the provided installation or personal access token.
        """
        if self.dry_run:
            logging.info(f"[DRY RUN] git remote add {BOT_REMOTE_NAME} https://x-access-token:TOKEN@github.com/{org}/{repo}")
            return

        remote_url = f"https://x-access-token:{token}@github.com/{org}/{repo}"
        try:
            remote = self.git_repo.remote(BOT_REMOTE_NAME)
            remote.set_url(remote_url)
        except ValueError:
            self.git_repo.create_remote(BOT_REMOTE_NAME, remote_url)

        self.remote = self.git_repo.remote(BOT_REMOTE_NAME)
        return self.remote

    def remove_remote_with_token(self):
        if self.dry_run:
            logging.info(f"[DRY RUN] git remote remove {BOT_REMOTE_NAME}")
            return
        self.remote.remove(self.git_repo, BOT_REMOTE_NAME)

    def push(self, branch_name, base_branch, gh_repo):
        """
        Replays local commits onto GitHub via the API so they are marked Verified.
        Commits created through the GitHub API are automatically signed by GitHub
        when using a GitHub App token, unlike commits created with git push.
        """
        if self.dry_run:
            logging.info(f"[DRY RUN] Creating verified commits via GitHub API for branch {branch_name}")
            return

        # Collect commits between base_branch and branch_name, oldest first
        commits = list(reversed(list(
            self.git_repo.iter_commits(f"{base_branch}..{branch_name}")
        )))

        if not commits:
            logging.info(f"No commits to push for branch {branch_name}")
            return

        parent_sha = gh_repo.get_branch(base_branch).commit.sha

        for local_commit in commits:
            # diff from parent → this commit: a=parent state, b=commit state
            diffs = (local_commit.parents[0].diff(local_commit)
                     if local_commit.parents else local_commit.diff(None))

            tree_elements = []
            for diff in diffs:
                if diff.deleted_file:
                    # sha=None signals a deletion to the GitHub tree API
                    tree_elements.append(InputGitTreeElement(
                        path=diff.a_path, mode="100644", type="blob", sha=None))
                else:
                    content = diff.b_blob.data_stream.read()
                    try:
                        blob = gh_repo.create_git_blob(content.decode("utf-8"), "utf-8")
                    except UnicodeDecodeError:
                        blob = gh_repo.create_git_blob(
                            base64.b64encode(content).decode("ascii"), "base64")
                    mode = "100755" if diff.b_blob.mode == 0o100755 else "100644"
                    tree_elements.append(InputGitTreeElement(
                        path=diff.b_path, mode=mode, type="blob", sha=blob.sha))

            parent_gh_commit = gh_repo.get_git_commit(parent_sha)
            new_tree = gh_repo.create_git_tree(tree_elements, parent_gh_commit.tree)
            new_commit = gh_repo.create_git_commit(local_commit.message, new_tree, [parent_gh_commit])
            logging.info(f"Created verified commit {new_commit.sha[:8]}: {local_commit.summary}")
            parent_sha = new_commit.sha

        try:
            ref = gh_repo.get_git_ref(f"heads/{branch_name}")
            ref.edit(parent_sha, force=True)
            logging.info(f"Updated branch '{branch_name}' to {parent_sha[:8]}")
        except GithubException:
            gh_repo.create_git_ref(f"refs/heads/{branch_name}", parent_sha)
            logging.info(f"Created branch '{branch_name}' at {parent_sha[:8]}")

    def get_remote_branch(self, branch_name):
        """
        Get the reference for the given branch on the specified Git remote,
        otherwise return None if the branch does not exist.
        """
        if self.dry_run:
            return None

        self.remote.fetch()
        matching_remote_refs = [ref for ref in self.remote.refs if BOT_REMOTE_NAME + "/" + branch_name == ref.name]

        if len(matching_remote_refs) == 0:
            logging.info(f"Branch '{branch_name}' does not exist on remote")
            return None

        if len(matching_remote_refs) > 1:
            matching_branches = ", ".join([r.name for r in matching_remote_refs])
            logging.warning(f"Found more than one branch matching '{branch_name}' " +
                            f"on remote: {matching_branches}. Taking first one")
            return matching_remote_refs[0]

        if len(matching_remote_refs) == 1:
            logging.info(f"Branch '{branch_name}' already exists on remote")
            return matching_remote_refs[0]

        return None

    def is_local_branch_based_on_newer_base_branch_commit(self, base_branch_name, remote_branch_name, local_branch_name):
        """
        Compares local and remote rebase branches by looking at their start on base branch.
        Returns True if local branch starts on newer commit and should to be pushed to remote,
        otherwise False.
        """
        if self.dry_run:
            return True

        remote_merge_base = self.git_repo.merge_base(base_branch_name, remote_branch_name)
        local_merge_base = self.git_repo.merge_base(base_branch_name, local_branch_name)

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


def commit_str(commit):
    """Returns the first 8 characters of the commit's SHA hash and the commit summary."""
    return f"{commit.hexsha[:8]} - {commit.summary}"
