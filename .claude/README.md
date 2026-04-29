# Claude Code Setup

Claude Code skills for MicroShift are available in the
[edge-tooling](https://github.com/openshift-eng/edge-tooling) repository
and are distributed as plugins.

## Installing Plugins

First, add the marketplace:

```
/plugin marketplace add openshift-eng/edge-tooling
```

Then install the plugins you need:

```
/plugin install microshift-ci
/plugin install microshift-dev
/plugin install microshift-release
```

| Plugin | Description |
|---|---|
| `microshift-ci` | CI job analysis, failure triage, and bug creation |
| `microshift-dev` | Development workflows, test generation, and code analysis |
| `microshift-release` | Release version tracking and release note generation |
