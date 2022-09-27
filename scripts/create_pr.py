#!/usr/bin/env python

"""Pull Request Creator

This script pushes current branch and creates GitHub Pull Request.
It's intended to be used as a GitHub App and requires following environment variables:
- APP_ID - application id, get from app's about page (https://github.com/settings/apps/$APP_NAME)
- KEY - path to application's private key - generate on app's about page
- ORG - organization
- REPO - repository

App requires following permissions:
- Contents Read+Write - to push branches
- Pull requests Read+Write - to create PRs
"""

import os
import sys
import subprocess
from git import Repo # GitPython
from github import GithubIntegration, Github # pygithub
from pathlib import Path

APP_ID_ENV = "APP_ID"
KEY_ENV = "KEY"
ORG_ENV = "ORG"
REPO_ENV = "REPO"

REMOTE_NAME = "bot-creds"

def try_get_env(var_name):
    val = os.getenv(var_name)
    if val is None or val == "":
        sys.exit(f"Env var {var_name} is empty")
    return val

app_id = try_get_env(APP_ID_ENV)
key_path = try_get_env(KEY_ENV)
org = try_get_env(ORG_ENV)
repo = try_get_env(REPO_ENV)

integration = GithubIntegration(app_id, Path(key_path).read_text())
app_installation = integration.get_installation(org, repo)
if app_installation == None:
    sys.exit(f"Failed to get app_installation for {org}/{repo}. Response: {app_installation.raw_data}")

installation_access_token = integration.get_access_token(app_installation.id).token

remote_url = f"https://x-access-token:{installation_access_token}@github.com/{org}/{repo}"
r = Repo('.')
try:
    remote = r.remote(REMOTE_NAME)
    remote.set_url(remote_url)
except ValueError:
    r.create_remote(REMOTE_NAME, remote_url)

remote = r.remote(REMOTE_NAME)
remote.push(r.active_branch.name)

gh = Github(installation_access_token)
repo = gh.get_repo(f"{org}/{repo}")
pr = repo.create_pull(title=r.active_branch.name, body='', base='main', head=r.active_branch.name, maintainer_can_modify=True)
print(f"Created pull request: {pr.html_url}")
