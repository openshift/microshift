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
from git import Repo, PushInfo # GitPython
from github import GithubIntegration, Github, GithubException # pygithub
from pathlib import Path

APP_ID_ENV = "APP_ID"
KEY_ENV = "KEY"
ORG_ENV = "ORG"
REPO_ENV = "REPO"

BOT_REMOTE_NAME = "bot-creds"

def try_get_env(var_name):
    val = os.getenv(var_name)
    if val is None or val == "":
        sys.exit(f"Env var {var_name} is empty")
    return val

app_id = try_get_env(APP_ID_ENV)
key_path = try_get_env(KEY_ENV)
org = try_get_env(ORG_ENV)
repo = try_get_env(REPO_ENV)

def commit_str(commit):
    return f"{commit.hexsha[:8]} - {commit.summary}"

def create_or_get_pr_url(ghrepo):
    prs = ghrepo.get_pulls(base='main', head=f"{org}:{r.active_branch.name}", state="all")
    if prs.totalCount == 1:
        print(f"{prs[0].state.capitalize()} pull request exists already: {prs[0].html_url}")
    elif prs.totalCount > 1:
        print(f"Found several existing PRs for '{r.active_branch.name}': {[(x.state, x.html_url) for x in prs]}")
    else:
        pr = ghrepo.create_pull(title=r.active_branch.name, body='', base='main', head=r.active_branch.name, maintainer_can_modify=True)
        print(f"Created pull request: {pr.html_url}")


integration = GithubIntegration(app_id, Path(key_path).read_text())
app_installation = integration.get_installation(org, repo)
if app_installation == None:
    sys.exit(f"Failed to get app_installation for {org}/{repo}. Response: {app_installation.raw_data}")
installation_access_token = integration.get_access_token(app_installation.id).token
gh = Github(installation_access_token)
ghrepo = gh.get_repo(f"{org}/{repo}")

r = Repo('.')
if r.active_branch.commit == r.branches["main"].commit:
    print(f"There's no new commit on branch {r.active_branch} compared to 'main'.\nLast commit: {r.active_branch.commit.hexsha[:8]} - \n\n{r.active_branch.commit.summary}'")
    sys.exit(0)

remote_url = f"https://x-access-token:{installation_access_token}@github.com/{org}/{repo}"
try:
    remote = r.remote(BOT_REMOTE_NAME)
    remote.set_url(remote_url)
except ValueError:
    r.create_remote(BOT_REMOTE_NAME, remote_url)

remote = r.remote(BOT_REMOTE_NAME)
remote.fetch()

# Check if branch with the same name exists in remote
matching_remote_branches = [ ref for ref in remote.refs if BOT_REMOTE_NAME + "/" + r.active_branch.name == ref.name ]
if len(matching_remote_branches) == 1:
    # Compare local and remote rebase branches by looking at their start on main branch (commit from which they branched off)
    merge_base_prev_rebase = r.merge_base("main", matching_remote_branches[0].name)
    merge_base_cur_rebase = r.merge_base("main", r.active_branch.name)
    if merge_base_prev_rebase[0] == merge_base_cur_rebase[0]:
        print(f"Branch {r.active_branch} already exists on remote and it's up to date.\n\
Branch-off commit: {commit_str(merge_base_cur_rebase[0])}\n")
        create_or_get_pr_url(ghrepo)
        sys.exit(0)
    else:
        print(f"Branch {r.active_branch} already exists on remote but it's out of date.\n\
Old branch-off commit: {commit_str(merge_base_prev_rebase[0])}\n\
New branch-off commit: {commit_str(merge_base_cur_rebase[0])}\n")

push_result = remote.push(r.active_branch.name, force=True)
if len(push_result) != 1:
    sys.exit(f"Unexpected amount ({len(push_result)}) of items in push_result: {push_result}")
if push_result[0].flags & PushInfo.ERROR:
    sys.exit(f"Pushing branch failed: {push_result[0].summary}")
if push_result[0].flags & PushInfo.FORCED_UPDATE:
    print("Branch was updated (force push)")

create_or_get_pr_url(ghrepo)
