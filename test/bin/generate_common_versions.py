import requests
import subprocess
import os

ARCH = os.uname().machine

MINOR = 20


def get_candidate_repo(minor, dev_preview=False):
    """
    :param minor: the minor version, e.g. 19 for 4.19
    :param dev_preview: if True, returns the engineering candidate repo, otherwise returns the release candidate repo

    ``get_candidate_repo`` returns the URL of the EC/RC repository for the specified minor version.
    """
    if dev_preview:
        return f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp-dev-preview/latest-4.{minor}/el9/os"
    else:
        return f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/microshift/ocp/latest-4.{minor}/el9/os"


def get_beta_repo(minor):
    """
    :param minor: the minor version, e.g. 19. for 4.19

    ``get_beta_repo`` returns the URL of the beta repository for the specified minor version.
    """
    for i in range(minor, minor-3, -1):
        url = f"https://mirror.openshift.com/pub/openshift-v4/{ARCH}/dependencies/rpms/4.{i}-el9-beta"
        if repo_exists(url) and provides_deps(url):
            return f'"{url}"'

    return None


def provides_deps(repo_url):
    """
    :param repo_url: the URL of the repository to check

    ``provides_deps`` checks if the repository provides the cri-o package.
    """
    try:
        temp = f"this,{repo_url}"
        _ = subprocess.run(
                ['dnf', 'repoquery', 'cri-o', '--quiet',
                 '--queryformat', '%{version}-%{release}', '--disablerepo', '\'*\'', '--repofrompath', temp],
                capture_output=True,
                text=True,
                check=True
            )

        return True

    except subprocess.CalledProcessError:
        return False


def dnf_repo_is_enabled(repo_name):
    """
    :param repo_name: the name of the repository to check

    ``dnf_repo_is_enabled`` checks if the repository is enabled on this system.
    """
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
    """
    :param repo_url: the URL of a repository

    ``repo_exists`` checks if a URL points to a valid repository.
    """
    url = repo_url + "/repodata/repomd.xml"
    r = requests.get(url)
    if r.status_code == 404:
        return False
    else:
        return True


def get_sub_repo(minor):
    """
    :param minor: the minor version, e.g. 19 for 4.19

    ``get_sub_repo`` returns the name of the subscription repository for the specified
    minor version, if the repository provides the microshift package, otherwise returns None.
    """
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


def get_release_repo(minor):
    """
    :param minor: the minor version, e.g. 19 for 4.19

    ``get_release_repo`` gets the repository for the specified minor version. It first
    tries the subscription repository, if it does not provide the microshift package,
    goes on to try the candidate repositories (EC, then RC). returns the name of the
    subscription repo, the candidate repo URL, or None.
    """
    repo = get_sub_repo(minor)
    if repo is not None:
        return f'"{repo}"'

    rc = get_candidate_repo(minor, dev_preview=False)
    if repo_exists(rc):
        return f'"{rc}"'

    ec = get_candidate_repo(minor, dev_preview=True)
    if repo_exists(ec):
        return f'"{ec}"'

    return '""'


def get_release_version(repo):
    """
    :param repo: the name or the URL of the repo, enclosed in double quotes

    ``get_release_version`` returns a string to be used as a bash variable,
    with a call to the right function depending on whether the `repo` param
    is a URL or not. if neither, returns empty double quotes.
    """
    if repo.startswith("\"rhocp"):
        return f'"$(get_vrel_from_rhsm {repo})"'
    elif repo.startswith("\"https"):
        return f'"$(get_vrel_from_beta {repo})"'
    else:
        return '""'


minor_version = MINOR
previous_minor_version = minor_version - 1
yminus2_minor_version = minor_version - 2
fake_next_minor_version = minor_version + 1

current_release_repo = get_release_repo(minor_version)
current_release_version = get_release_version(current_release_repo)

previous_release_repo = get_release_repo(previous_minor_version)
previous_release_version = get_release_version(previous_release_repo)

yminus2_release_repo = get_release_repo(yminus2_minor_version)
yminus2_release_version = get_release_version(yminus2_release_repo)

rhocp_minor_y = minor_version if get_sub_repo(minor_version) is not None else '""'
rhocp_minor_y_beta = get_beta_repo(minor_version)

rhocp_minor_y1 = previous_minor_version
rhocp_minor_y1_beta = get_beta_repo(previous_minor_version)

rhocp_minor_y2 = yminus2_minor_version

cncf_sonobuoy_version = "v0.57.3"

print(f"export MINOR_VERSION={minor_version}")
print(f"export PREVIOUS_MINOR_VERSION={previous_minor_version}")
print(f"export YMINUS2_MINOR_VERSION={yminus2_minor_version}")
print(f"export FAKE_NEXT_MINOR_VERSION={fake_next_minor_version}")
print("")

print(f"CURRENT_RELEASE_REPO={current_release_repo}")
print(f"CURRENT_RELEASE_VERSION={current_release_version}")
print("export CURRENT_RELEASE_REPO")
print("export CURRENT_RELEASE_VERSION")
print("")

print(f"PREVIOUS_RELEASE_REPO={previous_release_repo}")
print(f"PREVIOUS_RELEASE_VERSION={previous_release_version}")
print("export PREVIOUS_RELEASE_REPO")
print("export PREVIOUS_RELEASE_VERSION")
print("")

print(f"YMINUS2_RELEASE_REPO={yminus2_release_repo}")
print(f"YMINUS2_RELEASE_VERSION={yminus2_release_version}")
print("export YMINUS2_RELEASE_REPO")
print("export YMINUS2_RELEASE_VERSION")
print("")

print(f"RHOCP_MINOR_Y={rhocp_minor_y}")
print(f"RHOCP_MINOR_Y_BETA={rhocp_minor_y_beta}")
print("export RHOCP_MINOR_Y")
print("export RHOCP_MINOR_Y_BETA")
print("")

print(f"RHOCP_MINOR_Y1={rhocp_minor_y1}")
print(f"RHOCP_MINOR_Y1_BETA={rhocp_minor_y1_beta}")
print("export RHOCP_MINOR_Y1")
print("export RHOCP_MINOR_Y1_BETA")
print("")

print(f"export RHOCP_MINOR_Y2={rhocp_minor_y2}")
print("")

print(f"export CNCF_SONOBUOY_VERSION={cncf_sonobuoy_version}")
