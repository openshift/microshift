#!/usr/bin/env python3

import logging

from git import PushInfo, Repo  # GitPython

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

    def add_file_to_staging_area(self, file_path):
        if self.dry_run:
            logging.info(f"[DRY RUN] git add {file_path}")
            return
        self.git_repo.index.add([file_path])

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

    def push(self, branch_name):
        if self.dry_run:
            logging.info(f"[DRY RUN] git push --force {branch_name}")
            return

        push_result = self.remote.push(branch_name, force=True)

        if len(push_result) != 1:
            raise Exception(f"Unexpected amount ({len(push_result)}) of items in push_result: {push_result}")
        if push_result[0].flags & PushInfo.ERROR:
            raise Exception(f"Pushing branch failed: {push_result[0].summary}")
        if push_result[0].flags & PushInfo.FORCED_UPDATE:
            logging.info(f"Branch '{branch_name}' existed and was updated (force push)")
