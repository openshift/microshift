import requests
import subprocess
import os

ARCH = os.uname().machine

MINOR_VERSION = 19
PREVIOUS_MINOR_VERSION = MINOR_VERSION - 1
YMINUS2_MINOR_VERSION = MINOR_VERSION - 2
FAKE_NEXT_MINOR_VERSION = MINOR_VERSION + 1

def get_candidate_repo(dev_preview=False):
    if dev_preview:
        return f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp-dev-preview/latest-4.{MINOR_VERSION}/el9/os"
    else:
        return f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp/latest-4.{MINOR_VERSION}/el9/os"

def dnf_repo_is_enabled(repo_name):
    try:
        result = subprocess.run(
            ["dnf", "repolist"],
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            check=True
        )
        return repo_name in result.stdout
    except subprocess.CalledProcessError:
        return False


def repo_exists(repo_url):
    url = repo_url + "/repodata/repomd.xml"
    r = requests.get(url)
    if r.status_code == 404:
        return False
    else:
        return True

""" def get_current_release_from_sub_repos(minor=MINOR_VERSION):
    try:
        repo = f"rhocp-4.{minor}-for-rhel-9-{ARCH}-rpms"
        result = subprocess.run(
                ['sudo', 'dnf', 'repoquery', 'microshift', '--quiet',
                '--queryformat', '%{version}-%{release}', '--repo', repo],
                capture_output=True,
                text=True,
                check=True
            )

        versions = result.stdout.strip().splitlines()
        if not versions:
            return None

        sort_proc = subprocess.run(
            ['sort', '--version-sort'],
            input='\n'.join(versions),
            capture_output=True,
            text=True,
            check=True
        )

        sorted_versions = sort_proc.stdout.strip().splitlines()
        if sorted_versions:
            return repo, sorted_versions[-1] 
        else:
            return None, None
    except subprocess.CalledProcessError:
        return None, None """

def get_sub_repo(minor=MINOR_VERSION):
    try:
        repo = f"rhocp-4.{minor}-for-rhel-9-{ARCH}-rpms"
        _ = subprocess.run(
                ['sudo', 'dnf', 'repoquery', 'microshift', '--quiet',
                '--queryformat', '%{version}-%{release}', '--repo', repo],
                capture_output=True,
                text=True,
                check=True
            )

        return repo
        
    except subprocess.CalledProcessError:
        return None



def get_current_release_repo():
    repo = get_sub_repo()
    if repo is not None:
        return repo
    rc = get_candidate_repo(dev_preview=False)
    if repo_exists(rc):
        return rc
    
    ec = get_candidate_repo(dev_preview=True)
    if repo_exists(ec):
        return ec

    return None
    
def get_current_release_version(repo):
    if 




CURRENT_RELEASE_REPO = get_current_release_repo()
CURRENT_RELEASE_VERSION = get_current_release_version(CURRENT_RELEASE_REPO)

print(CURRENT_RELEASE_REPO)



