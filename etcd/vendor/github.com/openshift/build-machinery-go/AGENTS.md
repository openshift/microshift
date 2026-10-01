# AI Agent Instructions for build-machinery-go

**Audience:** AI agents editing **this repository** only — not downstream repos
that vendor these fragments. Those repos should maintain their own root-level
`AGENTS.md`.

| File | Purpose |
|------|---------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Make stack design, project layout, verification model |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Development workflow, PR expectations, external guidelines |

## What This Repo Is

Reusable GNU Make fragments and helper scripts. Downstream repos include one
vendored entry file (`golang.mk`, `default.mk`, or `operator.mk`) in their
`Makefile`. See [ARCHITECTURE.md](ARCHITECTURE.md) for stacks, layout, and
include chains.

## Critical Rules

1. **Run `make verify` before considering any change complete.**
2. **Do not hand-edit `*.log` files** — regenerate with `make update`.
3. **Add new behavior in `make/targets/`**, not in entry files.
4. **Keep backward compatibility** — every repo that vendors this module is affected.
5. **Update examples and logs together** when changing `make/targets/`.

## What NOT to Do

- Hand-edit `*.example.mk.help.log` or `Makefile.test.log`.
- Change `make/targets/` without updating examples and regenerating logs.
- Edit files under `make/examples/*/vendor/`.
- Duplicate logic across entry files.
- Modify OWNERS or OWNERS_ALIASES.
- Use AI to respond to review comments.

For workflow details (container image, branch name, commit structure), see
[CONTRIBUTING.md](CONTRIBUTING.md). For org-wide conventions, see the links in
CONTRIBUTING.md and [openshift/coderabbit](https://github.com/openshift/coderabbit).
