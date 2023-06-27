"""Provides helpers for common data formats
"""
import json

import yaml
from robot.utils import DotDict


def json_parse(data):
    """Parse input string as JSON and return DotDict instance.

    If the input data is empty, return an empty DotDict.
    """
    if not data:
        return DotDict()
    parsed = json.loads(data)
    return DotDict(parsed)


def yaml_parse(data):
    """Parse input string as YAML and return DotDict instance.

    If the input data is empty, return an empty DotDict.
    """
    if not data:
        return DotDict()
    parsed = yaml.safe_load(data)
    return DotDict(parsed)


def _merge(dest, addition):
    """Implements data structure merge.

    The dest value is modified in place by adding missing values found
    in the addition parameter.  Existing containers (dict, list) are
    updated. Scalar values (int, string) are replaced.

    >>> _merge({}, {})
    {}

    >>> _merge({}, {'new-key': 'value'})
    {'new-key': 'value'}

    >>> _merge({'same-key': 'old-value'}, {'same-key': 'new-value'})
    {'same-key': 'new-value'}

    >>> _merge({'nested': {}}, {'nested': {'new-key': 'new-value'}})
    {'nested': {'new-key': 'new-value'}}

    >>> _merge({'list': ['a']}, {'list': ['b']})
    {'list': ['a', 'b']}

    >>> _merge({'nested-list': {'list': ['a']}}, {'nested-list': {'list': ['b']}})
    {'nested-list': {'list': ['a', 'b']}}

    """
    for key, value in addition.items():
        if key in dest:
            if isinstance(value, dict):
                _merge(dest[key], value)
                continue
            if isinstance(value, list):
                dest[key].extend(value)
                continue
        # Either a new item or replacing a non-container type
        dest[key] = value
    # The return value is for doctest
    return dest


def yaml_merge(base, addition):
    """Return combination of both YAML data structures, additively."""
    if not base:
        combined = {}
    else:
        combined = yaml.safe_load(base)
    if not addition:
        parsed_addition = {}
    else:
        parsed_addition = yaml.safe_load(addition)
    _merge(combined, parsed_addition)
    return yaml.dump(combined)


def update_kubeconfig_server_url(kubeconfig_text, new_url):
    """Change the server URL and return the new kubeconfig text."""
    parsed = yaml.safe_load(kubeconfig_text)
    parsed['clusters'][0]['cluster']['server'] = new_url
    return yaml.dump(parsed)


# lvmd_merge idempotently merges the provided patch into the base lvmd config. If the patch is already present in the
# device class list, the base config is returned without mutation. Necessary because LVMD does not tolerate device-classes
# of the same name.
def lvmd_merge(base, patch):
    if base is None:
        raise ValueError('cannot process empty base')
    if patch is None:
        return base

    dev_classes = yaml.safe_load(base)['device-classes']
    dc_patch_name = yaml.safe_load(patch)['device-classes'][0]['name']

    if any(d['name'] == dc_patch_name for d in dev_classes):
        return base
    return yaml_merge(base, patch)
