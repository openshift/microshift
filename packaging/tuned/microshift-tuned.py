#!/usr/bin/env python

import argparse
import os.path
import subprocess
import configparser
import sys
import yaml
import logging

logging.basicConfig(level=logging.DEBUG, format='L#%(lineno)-4d %(levelname)-8s %(message)s')


class Config:
    PATH = "/etc/microshift/tuned.yaml"
    PROFILE = "profile"
    REBOOT_AFTER_APPLY = "reboot_after_apply"
    EXPECTED_FIELDS = [PROFILE, REBOOT_AFTER_APPLY]

    def __init__(self, profile, reboot_after_apply):
        self.profile = profile
        self.reboot_after_apply = reboot_after_apply

    @staticmethod
    def load():
        if not os.path.exists(Config.PATH):
            logging.error(f"Configuration file '{Config.PATH}' does not exist, cannot proceed.")
            sys.exit(1)

        with open(Config.PATH) as cfg_file:
            cfg = yaml.safe_load(cfg_file)

        logging.debug(f"Loaded config file: {cfg}")

        valid = True
        for field in Config.EXPECTED_FIELDS:
            if field not in cfg:
                logging.error(f"Incorrect configuration file: '{field}' is missing")
                valid = False

        # Early exit if some variables are missing
        if not valid:
            sys.exit(1)

        if cfg[Config.PROFILE].strip() == "":
            logging.error(f"Invalid config: empty '{Config.PROFILE}' value: '{cfg[Config.PROFILE]}'")
            valid = False

        if not isinstance(cfg[Config.REBOOT_AFTER_APPLY], bool):
            logging.error(f"Invalid config: '{Config.REBOOT_AFTER_APPLY}' must be True or False")
            valid = False

        if not valid:
            sys.exit(1)

        return Config(cfg[Config.PROFILE], cfg[Config.REBOOT_AFTER_APPLY])


class Checksums:
    PATH = "/var/lib/microshift-tuned.yaml"

    def __init__(self, profile_checksum, variables_checksum):
        self.profile_checksum = profile_checksum
        self.variables_checksum = variables_checksum

    def __eq__(self, other):
        return ((self.profile_checksum, self.variables_checksum)
                == (other.profile_checksum, other.variables_checksum))

    def __repr__(self):
        return f"Checksums(profile: '{self.profile_checksum}', variables: '{self.variables_checksum}')"

    @staticmethod
    def load_from_cache():
        if not os.path.exists(Checksums.PATH):
            logging.debug("Cache does not exist")
            return None

        with open(Checksums.PATH, 'r') as cache_file:
            cache = yaml.safe_load(cache_file)
            checksums = Checksums(cache["profile_checksum"], cache["variables_checksum"])
            logging.debug(f"Loaded cache: {checksums}")
            return checksums

    def write_to_cache(self):
        logging.debug(f"Updating cache: {self}")
        cache = {
            "profile_checksum": self.profile_checksum,
            "variables_checksum": self.variables_checksum,
        }
        with open(Checksums.PATH, 'w') as cache_file:
            yaml.dump(cache, cache_file)


def get_active_profile() -> tuple[str, bool]:
    stdout, success = run_command(["tuned-adm", "active"])
    if not success:
        logging.debug("No active TuneD profile")
        return ("", False)
    profile = stdout.split(": ")[1].strip()
    logging.debug(f"Active TuneD profile: '{profile}'")
    return (profile, True)


def run_command(cmd: list[str], failure_fatal=False) -> tuple[str, bool]:
    logging.debug(f"Executing command: '{' '.join(cmd)}'")
    result = subprocess.run(cmd, capture_output=True)
    stdout = result.stdout.decode('utf-8')
    to_log = {"stdout": stdout, "stderr": result.stderr.decode('utf-8'), "rc": result.returncode}
    logging.debug(f"Results of '{' '.join(cmd)}': {to_log}")
    if failure_fatal and result.returncode != 0:
        logging.error(f"Command '{' '.join(cmd)}' failed")
        sys.exit(1)
    return (stdout, result.returncode == 0)


def get_profile_path(profile: str) -> str:
    paths = [f"/etc/tuned/{profile}", f"/usr/lib/tuned/{profile}"]
    for path in paths:
        if os.path.exists(path):
            logging.debug(f"Found profile '{profile}' in '{path}'")
            return path
    logging.error(f"Could not find profile '{profile}' in paths: {paths}")
    sys.exit(1)


def get_variables_file_path(profile_path: str) -> str:
    profile = configparser.ConfigParser()
    profile.read(os.path.join(profile_path, "tuned.conf"))
    if "variables" in profile and "include" in profile["variables"]:
        variables_file = profile["variables"]["include"]
        logging.debug(f"Profile '{profile_path}' includes '{variables_file}'")
        if not os.path.exists(variables_file):
            logging.error(f"'{variables_file}' doesn't exist")
            sys.exit(1)
        return variables_file
    logging.debug(f"Profile '{profile_path}' does not include variables")
    return ""


def get_profile_checksum(profile_path: str, variables_path: str) -> Checksums:
    # Get md5sum of /{etc,usr/lib}/tuned/PROFILE/* contents.
    # Alternative would be to use `tar c` to get contents, ownership, permissions, and timestamps,
    # but that could cause unnecessary reboots if files' timestamps got updated without changes to contents.
    profile_checksum, _ = run_command(["bash", "-c", f"set -o pipefail && cat {os.path.join(profile_path, '*')} | md5sum"], failure_fatal=True)
    if variables_path != "":
        variables_checksum, _ = run_command(["bash", "-c", f"set -o pipefail && cat {variables_path} | md5sum"], failure_fatal=True)
    else:
        variables_checksum = ""
    checksums = Checksums(profile_checksum.split(' ')[0], variables_checksum.split(' ')[0])
    logging.info(f"Calculated checksums of requested profile: {checksums}")
    return checksums


def activate_profile(profile: str) -> None:
    run_command(["tuned-adm", "profile", profile], failure_fatal=True)


def reboot() -> None:
    run_command(["systemctl", "--message='Reboot to fully activate tuned profile'", "reboot"])


def tuned_daemon_should_be_running():
    _, success = run_command(["systemctl", "is-active", "tuned.service"])
    if not success:
        logging.error("TuneD service is not running")
        sys.exit(1)


def should_run_as_root():
    if os.getuid() != 0:
        logging.error("Program must run with root privileges")
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        prog='MicroShift TuneD',
        formatter_class=argparse.RawTextHelpFormatter,
        description=f"""Daemon for unattended TuneD profile activation.

When program starts, it compares configuration and system state.
If the requested profile is not activate, it will be activated,
and if reboot_after_apply is True, the node will be rebooted.
When profile is being activated, checksums of the profile contents
and its variables are stored in separate location.
Later, when program starts again, it compares stored checksums with
current state of the profile and only reactivate and (optionally)
reboot the host if the profile has changed.

Configuration file is at {Config.PATH} and it has following schema:
    profile: <name of the tuned profile>
    reboot_after_apply: {{True|False}}"""
    )
    parser.add_argument("--live-run",
                        action='store_true',
                        help="Allows program to reboot the node if it was requested in the configuration file (reboot_after_apply)")

    args = parser.parse_args()

    should_run_as_root()
    tuned_daemon_should_be_running()
    cfg = Config.load()

    profile_path = get_profile_path(cfg.profile)
    vars_path = get_variables_file_path(profile_path)
    checksums = get_profile_checksum(profile_path, vars_path)

    active_profile, active = get_active_profile()
    if active and cfg.profile == active_profile:
        logging.info(f"Active profile and requested profile are the same: '{active_profile}'.")
        cache = Checksums.load_from_cache()
        if cache is not None and cache == checksums:
            logging.info("No changes to profile or variables detected. Exiting...")
            sys.exit(0)

    activate_profile(cfg.profile)
    checksums.write_to_cache()

    if cfg.reboot_after_apply:
        if args.live_run:
            logging.info("Rebooting the host")
            reboot()
        else:
            logging.info("Reboot is skipped because --live-run was not provided.")


if __name__ == "__main__":
    main()
