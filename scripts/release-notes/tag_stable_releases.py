#!/usr/bin/env python3

import argparse
import datetime
import logging
import os
import platform
import subprocess
import sys

import dnf

logger = logging.getLogger()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--verbose', '-v', action='store_true', dest='verbose')
    args = parser.parse_args()

    log_level = logging.WARN
    if args.verbose:
        log_level = logging.DEBUG

    logging.basicConfig(
        stream=sys.stdout,
        level=log_level,
    )

    machine_arch = platform.machine()

    logger.debug('reading dnf database')
    base = dnf.Base()
    base.read_all_repos()
    base.fill_sack()
    logger.debug('finding matching repositories')
    repos = base.repos.get_matching(f'rhocp-4.*-{machine_arch}-rpms')

    for repo in sorted(repos, key=lambda r: r.name):
        reponame = repo.id
        logger.debug('starting %s', reponame)

        q = base.sack.query().available().filter(
            name='microshift',
            reponame=reponame,
        )

        for pkg in q:
            # a release string looks like:
            #   202305161335.p0.g17cae44.assembly.4.13.0.el9
            sha = pkg.release.split('.')[2].lstrip('g')
            buildtime = datetime.datetime.fromtimestamp(pkg.buildtime)
            tag = f'v{pkg.version}'
            logger.debug('package %s sha %s tag %s buildtime %s', pkg, sha, tag, buildtime)
            if tag_exists(tag):
                logger.debug('found tag %s', tag)
            else:
                tag_release(tag, sha, buildtime)


def tag_exists(tag):
    "Checks if a given tag exists in the local repository."
    try:
        subprocess.run(["git", "show", tag],
                       stdout=subprocess.DEVNULL,
                       stderr=subprocess.DEVNULL,
                       check=True)
        return True
    except subprocess.CalledProcessError:
        return False


def tag_release(tag, sha, buildtime):
    env = {}
    # Include our existing environment settings to ensure values like
    # HOME and other git settings are propagated.
    env.update(os.environ)
    timestamp = buildtime.strftime('%Y-%m-%d %H:%M')
    env['GIT_COMMITTER_DATE'] = timestamp
    print(f'GIT_COMMITTER_DATE={timestamp} git tag -s {tag} {sha}')
    subprocess.run(
        ['git', 'tag', '-s', '-m', tag, tag, sha],
        env=env,
        check=True,
    )


# git-tag man page shows how to backdate tags using GIT_COMMITTER_DATE variable
if __name__ == '__main__':
    main()
