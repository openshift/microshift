#!/usr/bin/env python3

"""Release Note Tool

This script partially automates the process of publishing release
notes for released builds of MicroShift.
"""

import argparse
import os
import dnf
import json
import textwrap
import logging
import platform

from github import Github  # pygithub

from gen_gh_releases_from_mirror import Release
import common

logging.basicConfig(level=logging.DEBUG, format='%(asctime)s %(levelname)-8s [%(filename)s:%(lineno)d] %(message)s')


def get_rpm_releases():
    """Gets the releases from the released MicroShift RPMs in the RHOCP repositories.
    """
    arch = platform.machine()
    _, minor = common.get_version_from_makefile()

    with dnf.Base() as base:
        base.read_all_repos()

        # Disable all repos (in the memory, not on the host) to just query RHOCP repos and make operation faster.
        base.repos.all().disable()
        for version in range(14, int(minor)+1):
            name = f'rhocp-4.{version}-for-rhel-9-{arch}-rpms'
            repo = base.repos.get(name)
            if repo:
                repo.enable()
                logging.info(f'Enabled repo: {name} - {repo.name}')

        try:
            base.fill_sack()
        except dnf.exceptions.RepoError as e:
            if 'All mirrors were tried' in str(e):
                # Not a real error, ignore
                logging.info(f'Repo {name} is not up yet, skipping')
            else:
                raise e
        except Exception as e:
            raise e

        releases = []

        query = base.sack.query().available().filter(name='microshift')
        for pkg in query:
            # Example pkg.version value: 4.16.47
            # Example pkg.release value: 202509010322.p0.g54f411b.assembly.4.16.47.el9

            pkg_release = pkg.release.split('.')
            release_date = pkg_release[0]
            commit_sha = pkg_release[2].lstrip('g')
            full_commit_sha = str.strip(common.run_process(["git", "rev-parse", "--verify", f"{commit_sha}^{{commit}}"]))
            if not full_commit_sha:
                raise RuntimeError(f"No full commit SHA found for {pkg}")

            rel = Release(
                f"{pkg.version}-{release_date}.{pkg_release[1]}",      # 4.16.47-202509010322.p0 - name of the GH release / tag
                full_commit_sha,                                       # 54f411ba7c8853ab28b84f018391e81993e04c99
                pkg.version,                                           # 4.16.47
                "", "", "",                                            # Not applicable as these fields are intended for differentiating ECs and RCs.
                release_date)                                          # 202509010322
            releases.append(rel)
            logging.info(f"Transformed RPM into Release struct: {pkg} => {rel}")

        return releases


def publish_release(new_release, take_action):
    """
    Creates a release preamble, then tags and publishes the release.
    """
    product_version = new_release.product_version
    xy = '.'.join(product_version.split('.')[0:2])

    # Set up the release notes preamble with download links
    preamble = textwrap.dedent(f"""
    To install MicroShift {product_version}, follow the documentation: https://docs.redhat.com/en/documentation/red_hat_build_of_microshift/{xy}
    """)

    common.publish_release(new_release, preamble, take_action)


def is_branch_synced_with_main(ci_job_branch):
    """
    Checks if the job is running against a branch that is synced with main.
    This script should only run on a single branch as it handles all versions of released RPMs.
    However, periodics against main are not allowed therefore we need to check if the current branch is synced with main.
    """
    gh_repo = Github().get_repo(f"{common.GITHUB_ORG}/{common.GITHUB_REPO}")
    issue = gh_repo.get_issue(number=1239)
    title = issue.title
    try:
        branches_part = title.split('|', 1)[1].strip()
        branch_tokens = [x.replace('branch:', '') for x in branches_part.split()]
        if ci_job_branch in branch_tokens:
            return True
    except Exception as e:
        raise RuntimeError(f"Failed to parse freeze issue title: {title} ({e})")
    return False


def get_args_parser():
    parser = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    subparsers = parser.add_subparsers(dest="command")

    parser_query = subparsers.add_parser(
        "query", help="Query the RHOCP repositories for all MicroShift RPMs. It's recommended to run this command as root to update the DNF cache.")
    parser_query.add_argument(
        '--ci-job-branch',
        action='store',
        default="",
        dest='ci_job_branch',
        help='The branch the CI job is running against - unnecessary if running the script locally',
    )
    parser_query.add_argument(
        '--keep-existing-tags',
        action='store_true',
        default=False,
        dest='keep_existing_tags',
        help='Keep releases that already exist based on the local repository tags. Default is to filter them out to avoid calling GitHub API unnecessarily.',
    )
    parser_query.add_argument(
        "--output", help="Path to save the results file", required=True)

    parser_publish = subparsers.add_parser(
        "publish", help="Publish the draft releases to the GitHub repository. Avoid running as root to avoid potential local git repo issues.")
    parser_publish.add_argument(
        "--input", help="Path to the results file generated by the 'query' command", required=True)
    parser_publish.add_argument(
        "-n", "--dry-run", action="store_true", default=False, help="Report but take no action")

    return parser


def main():
    common.load_github_token()  # Must be first thing in order for common.redact() to work properly.

    parser = get_args_parser()
    args = parser.parse_args()

    if args.command == "query":
        if args.ci_job_branch and not is_branch_synced_with_main(args.ci_job_branch):
            logging.warning(f"The CI job is running against a branch that is not synced with main: {args.ci_job_branch}")
            exit(0)

        if os.getuid() != 0:
            logging.warning("Script is not running as root - dnf cache will not be updated.")

        rpm_releases = get_rpm_releases()
        # Sort oldest first, so created GH releases have something to diff against (older patch versions within the same minor version).
        rpm_releases.sort(key=lambda r: r.release_date)
        if not args.keep_existing_tags:
            rpm_releases = [r for r in rpm_releases if not common.tag_exists(r.release_name)]
        with open(args.output, "w", encoding="utf-8") as f:
            # Cannot serialize Release objects directly to JSON, so convert to dicts first.
            release_dicts = [r._asdict() for r in rpm_releases]
            f.write(json.dumps(release_dicts).replace("},", "},\n"))
            logging.info(f"Saved {len(rpm_releases)} releases to {args.output}")

    elif args.command == "publish":
        if os.getuid() == 0:
            logging.warning("Running 'publish' as root is not recommended to avoid potential local git repo issues.")
            exit(0)

        with open(args.input, "r", encoding="utf-8") as f:
            j = json.load(f)
            rpm_releases = [Release(**r) for r in j]
        common.add_token_remote()
        logging.info(f"Attempting to publish {len(rpm_releases)} draft releases")
        i = 1
        for r in rpm_releases:
            if not common.github_release_exists(r.release_name):
                publish_release(r, not args.dry_run)
            else:
                logging.info(f"Release {r.release_name} already exists on remote GitHub repository, skipping")
            logging.info(f"Progress: {i}/{len(rpm_releases)}")
            i += 1

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
