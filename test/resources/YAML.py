"""Expose YAML parser to robot tests.
"""
import yaml
from robot.utils import DotDict


def yaml_parse(data):
    """Parse input string as YAML and return DotDict instance."""
    parsed = yaml.safe_load(data)
    return DotDict(parsed)
