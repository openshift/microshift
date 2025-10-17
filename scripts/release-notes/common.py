import os
import logging
import github
import pathlib
import datetime
import subprocess
from urllib import request
import json
import urllib.error

GITHUB_ORG = 'openshift'
GITHUB_REPO = 'microshift'
REMOTE = "token-remote"
MAX_RELEASE_NOTE_BODY_SIZE = 125000
TRUNCATED_MESSAGE = '\n\n(release notes were truncated)\n\n'

GITHUB_TOKEN = ""

def load_github_token():
    global GITHUB_TOKEN
    GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN')

    if not GITHUB_TOKEN:
        logging.warning('GITHUB_TOKEN is not set, trying to authenticate using application credentials')
        GITHUB_TOKEN = get_access_token_for_app()

    if not GITHUB_TOKEN:
        raise RuntimeError('GITHUB_TOKEN does not appear to be set')


def get_access_token_for_app():
    """Return an access token for a GitHub App."""
    app_id = os.environ.get('APP_ID')
    if not app_id:
        raise RuntimeError('APP_ID is not set')
    key_path = os.environ.get('CLIENT_KEY')
    if not key_path:
        raise RuntimeError('CLIENT_KEY is not set')
    integration = github.GithubIntegration(
        app_id,
        pathlib.Path(key_path).read_text(encoding='utf-8'),
    )
    app_installation = integration.get_installation(GITHUB_ORG, GITHUB_REPO)
    if app_installation is None:
        raise RuntimeError(
            f"Failed to get app_installation for {GITHUB_ORG}/{GITHUB_REPO}. " +
            f"Response: {app_installation.raw_data}"
        )
    return integration.get_access_token(app_installation.id).token


def get_version_from_makefile():
    # The script runs as
    # .../microshift/scripts/release-notes/common.py
    # and we want the path to
    # .../microshift
    root_dir = pathlib.Path(__file__).parent.parent.parent
    version_makefile = root_dir / 'Makefile.version.aarch64.var'
    # Makefile contains something like
    #   OCP_VERSION := 4.16.0-0.nightly-arm64-2024-03-13-041907
    # and we want this ^^^^
    #
    # We get it as ['4', '16'] to make the next part of the process of
    # building the list of versions to scan easier.
    _full_version = version_makefile.read_text('utf8').split('=')[-1].strip()
    major, minor = _full_version.split('.')[:2]
    return major, minor


def redact(input):
    return str.replace(input, GITHUB_TOKEN, '~~REDACTED~~')


def run_process(cmd: list[str], env={}):
    """
    Helper function to run external commands and log (redacted) output.
    Stdout is returned as a str.
    If command fails, exception is raised.
    """
    cmd_to_log = redact(' '.join(cmd))
    logging.debug(f"Running command: {cmd_to_log}")

    # Include our existing environment settings to ensure values like
    # HOME and other git settings are propagated.
    env.update(os.environ)

    completed = subprocess.run(
        cmd,
        env=env,
        capture_output=True
    )
    sout = completed.stdout.decode('utf-8') if completed.stdout else ''
    serr = str.strip(redact(completed.stderr.decode('utf-8'))) if completed.stderr else ''
    logging.debug(f"Command '{cmd_to_log}' finished: rc='{completed.returncode}' stdout='{str.strip(redact(sout))}' stderr='{serr}'")

    if completed.returncode != 0:
        raise subprocess.CalledProcessError(completed.returncode, cmd_to_log, redact(sout), serr)

    return sout


def tag_exists(release_name):
    "Checks if a given tag exists in the local repository."
    try:
        run_process(["git", "show", "--quiet", release_name])
        return True
    except subprocess.CalledProcessError:
        return False


def add_token_remote():
    """
    Adds the Git remote to the given repository using
    the provided installation (or personal) access token.
    """
    try:
        run_process(["git", "remote", "remove", REMOTE])
    except subprocess.CalledProcessError:
        pass

    remote_url = f"https://x-access-token:{GITHUB_TOKEN}@github.com/{GITHUB_ORG}/{GITHUB_REPO}"
    run_process(["git", "remote", "add", REMOTE, remote_url])


def get_previous_tag(release_name):
    "Returns the name of the tag _before_ release_name on the branch."
    output = run_process(["git", "describe", f'{release_name}~1', '--abbrev=0'])
    return output.strip()


def tag_release(tag, sha, buildtime):
    env = {}
    timestamp = buildtime.strftime('%Y-%m-%d %H:%M')
    env['GIT_COMMITTER_DATE'] = timestamp

    logging.info(f"Using 'GIT_COMMITTER_DATE={timestamp}' for 'git tag {tag} {sha}'")
    run_process(['git', 'tag', '-m', tag, tag, sha], env)


def push_tag(tag):
    run_process(['git', 'push', REMOTE, tag])


def publish_release(new_release, preamble, take_action):
    """Does the work to tag and publish a release.
    """
    release_name = new_release.release_name
    commit_sha = new_release.commit_sha
    release_date = new_release.release_date

    if not tag_exists(release_name):
        # release_date looks like 202402022103
        buildtime = datetime.datetime.strptime(release_date, '%Y%m%d%H%M')
        tag_release(release_name, commit_sha, buildtime)

    # Get the previous tag on the branch as the starting point for the
    # release notes.
    previous_tag = get_previous_tag(release_name)

    # Auto-generate the release notes ourselves, add the preamble,
    # then make sure the results fit within the size limits imposed by
    # the API.
    generated_notes = github_release_notes(previous_tag, release_name, commit_sha)
    notes = f'{preamble}\n{generated_notes["body"]}'
    if len(notes) > MAX_RELEASE_NOTE_BODY_SIZE:
        lines = notes.splitlines()
        last_line = lines[-1]
        notes_content_we_can_truncate = notes[:-len(last_line)]
        amount_we_can_keep = MAX_RELEASE_NOTE_BODY_SIZE - len(last_line) - len(TRUNCATED_MESSAGE)
        truncated = notes_content_we_can_truncate[:amount_we_can_keep]
        if truncated[-1] == '\n':
            notes_to_keep = truncated
        else:
            # don't leave a partial line
            notes_to_keep = truncated.rpartition('\n')[0].rstrip()
        notes = f'{notes_to_keep}{TRUNCATED_MESSAGE}{last_line}'

    if not take_action:
        logging.info(f'Dry run for new release {new_release} on commit {commit_sha} from {release_date}')
        logging.info(notes)
        return

    push_tag(release_name)

    # Create draft release with message that includes download URLs and history
    github_release_create(release_name, notes)


def github_release_create(tag, notes):
    results = github_api(
        f'/repos/{GITHUB_ORG}/{GITHUB_REPO}/releases',
        tag_name=tag,
        name=tag,
        body=notes,
        draft=False,
        prerelease=True,
    )
    logging.info(f'Created new release {tag}:{ {"url":results["html_url"], "body": results["body"]} }')


def github_release_notes(previous_tag, tag_name, target_commitish):
    results = github_api(
        f'/repos/{GITHUB_ORG}/{GITHUB_REPO}/releases/generate-notes',
        tag_name=tag_name,
        target_commitish=target_commitish,
        previous_tag_name=previous_tag,
    )
    return results


def github_release_exists(tag):
    try:
        github_api(f'/repos/{GITHUB_ORG}/{GITHUB_REPO}/releases/tags/{tag}')
        return True
    except Exception:
        return False


def github_api(path, **data):
    url = f'https://api.github.com/{path.lstrip("/")}'
    if data:
        r = request.Request(
            url=url,
            data=json.dumps(data).encode('utf-8'),
        )
    else:
        r = request.Request(url=url)

    logging.info(f"GitHub API Request: { {'method':r.get_method(), 'url': url, 'data': data} }")
    r.add_header('Accept', 'application/vnd.github+json')
    r.add_header('User-agent', 'microshift-release-notes')
    r.add_header('Authorization', f'Bearer {GITHUB_TOKEN}')
    r.add_header('X-GitHub-Api-Version', '2022-11-28')

    try:
        response = request.urlopen(r)
    except urllib.error.URLError as e:
        logging.error(f"GitHub API Request Failed: '{str(e.fp.readlines())}'")
        # e.fp.readlines() sinks the response body but it's not read in any other place,
        # so just re-raise for the exception type.
        raise
    except Exception as err:
        logging.error(f"GitHub API Request Failed: '{err}'")
        raise

    return json.loads(response.read().decode('utf-8'))
