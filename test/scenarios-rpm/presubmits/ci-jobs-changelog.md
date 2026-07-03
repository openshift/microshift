# CI Jobs Changelog

Changes to MicroShift CI jobs in `openshift/release` PR #81195.

## `openshift-microshift-main.yaml` (presubmits + on-demand periodic triggers)

| Status | Old Name | New Name | Notes |
|--------|----------|----------|-------|
| **NEW** | — | `e2e-aws-tests-rpm-presubmit-el9-x86` | RPM presubmit (replaces ostree `e2e-aws-tests`) |
| **NEW** | — | `e2e-aws-tests-rpm-presubmit-el10-x86` | RPM presubmit (replaces ostree `e2e-aws-tests-arm`) |
| **NEW** | — | `e2e-aws-tests-ostree-periodic-el9-functional-x86` | Split from `e2e-aws-tests-periodic` |
| **NEW** | — | `e2e-aws-tests-ostree-periodic-el9-functional-arm` | Split from `e2e-aws-tests-periodic-arm` |
| **NEW** | — | `e2e-aws-tests-bootc-periodic-el9-functional-x86` | Split from `e2e-aws-tests-bootc-periodic-el9` |
| **NEW** | — | `e2e-aws-tests-bootc-periodic-el10-functional-x86` | Split from `e2e-aws-tests-bootc-periodic-el10` |
| **NEW** | — | `e2e-aws-tests-bootc-periodic-el9-functional-arm` | Split from `e2e-aws-tests-bootc-periodic-arm-el9` |
| **NEW** | — | `e2e-aws-tests-bootc-periodic-el10-functional-arm` | Split from `e2e-aws-tests-bootc-periodic-arm-el10` |
| **DELETED** | `e2e-aws-tests` | — | Ostree presubmit → moved to nightly |
| **DELETED** | `e2e-aws-tests-arm` | — | Ostree presubmit → moved to nightly |
| **DELETED** | `e2e-aws-tests-bootc-el9` | — | Bootc presubmit → moved to nightly |
| **DELETED** | `e2e-aws-tests-bootc-el10` | — | Bootc presubmit → moved to nightly |
| **DELETED** | `e2e-aws-tests-bootc-arm-el9` | — | Bootc presubmit → moved to nightly |
| **DELETED** | `e2e-aws-tests-bootc-arm-el10` | — | Bootc presubmit → moved to nightly |
| RENAMED | `e2e-aws-tests-bootc-upstream-periodic` | `e2e-aws-tests-bootc-upstream-periodic-x86` | Added `-x86` |
| RENAMED | `e2e-aws-tests-cache-nightly` | `e2e-aws-tests-cache-nightly-x86` | Added `-x86` |
| RENAMED | `e2e-aws-tests-cache` | `e2e-aws-tests-cache-x86` | Added `-x86` |
| RENAMED | `e2e-aws-tests-periodic` | `e2e-aws-tests-ostree-periodic-el9-lifecycle-x86` | Lifecycle split + naming convention |
| RENAMED | `e2e-aws-tests-periodic-arm` | `e2e-aws-tests-ostree-periodic-el9-lifecycle-arm` | Lifecycle split + naming convention |
| RENAMED | `e2e-aws-tests-bootc-periodic-el9` | `e2e-aws-tests-bootc-periodic-el9-lifecycle-x86` | Lifecycle split + `-x86` |
| RENAMED | `e2e-aws-tests-bootc-periodic-el10` | `e2e-aws-tests-bootc-periodic-el10-lifecycle-x86` | Lifecycle split + `-x86` |
| RENAMED | `e2e-aws-tests-bootc-periodic-arm-el9` | `e2e-aws-tests-bootc-periodic-el9-lifecycle-arm` | Reordered + lifecycle |
| RENAMED | `e2e-aws-tests-bootc-periodic-arm-el10` | `e2e-aws-tests-bootc-periodic-el10-lifecycle-arm` | Reordered + lifecycle |
| RENAMED | `e2e-aws-tests-release` | `e2e-aws-tests-ostree-release-el9-x86` | Naming convention |
| RENAMED | `e2e-aws-tests-release-arm` | `e2e-aws-tests-ostree-release-el9-arm` | Naming convention |
| RENAMED | `e2e-aws-tests-bootc-release-el9` | `e2e-aws-tests-bootc-release-el9-x86` | Added `-x86` |
| RENAMED | `e2e-aws-tests-bootc-release-el10` | `e2e-aws-tests-bootc-release-el10-x86` | Added `-x86` |
| RENAMED | `e2e-aws-tests-bootc-release-arm-el9` | `e2e-aws-tests-bootc-release-el9-arm` | Reordered `-arm` |
| RENAMED | `e2e-aws-tests-bootc-release-arm-el10` | `e2e-aws-tests-bootc-release-el10-arm` | Reordered `-arm` |
| RENAMED | `e2e-aws-tests-bootc-upstream` | `e2e-aws-tests-bootc-presubmit-upstream-x86` | Added trigger + arch |
| RENAMED | `e2e-aws-tests-bootc-upstream-arm` | `e2e-aws-tests-bootc-presubmit-upstream-arm` | Added trigger |
| RENAMED | `e2e-aws-tests-bootc-c2cc` | `e2e-aws-tests-bootc-presubmit-c2cc-x86` | Added trigger + arch |
| RENAMED | `e2e-aws-tests-bootc-c2cc-arm` | `e2e-aws-tests-bootc-presubmit-c2cc-arm` | Added trigger |

## `openshift-microshift-release-5.0__periodics.yaml` (nightly crons)

| Status | Old Name | New Name | Notes |
|--------|----------|----------|-------|
| **NEW** | — | `e2e-aws-tests-bootc-x86-nightly-el9-functional` | Split from `e2e-aws-tests-bootc-nightly-el9` |
| **NEW** | — | `e2e-aws-tests-bootc-x86-nightly-el10-functional` | Split from `e2e-aws-tests-bootc-nightly-el10` |
| **NEW** | — | `e2e-aws-tests-bootc-arm-nightly-el9-functional` | Split from `e2e-aws-tests-bootc-arm-nightly-el9` |
| **NEW** | — | `e2e-aws-tests-bootc-arm-nightly-el10-functional` | Split from `e2e-aws-tests-bootc-arm-nightly-el10` |
| **NEW** | — | `e2e-aws-tests-ostree-x86-nightly-el9-functional` | Split from `e2e-aws-tests-nightly` |
| **NEW** | — | `e2e-aws-tests-ostree-arm-nightly-el9-functional` | Split from `e2e-aws-tests-arm-nightly` |
| **NEW** | — | `e2e-aws-tests-bootc-x86-nightly-el10-lifecycle` | New (el10 was 1 job, now split) |
| **NEW** | — | `e2e-aws-tests-bootc-arm-nightly-el10-lifecycle` | New (el10 was 1 job, now split) |
| RENAMED | `e2e-aws-tests-bootc-nightly-el9` | `e2e-aws-tests-bootc-x86-nightly-el9-lifecycle` | Naming convention + lifecycle |
| RENAMED | `e2e-aws-tests-bootc-nightly-el10` | `e2e-aws-tests-bootc-x86-nightly-el10-lifecycle` | Naming convention + lifecycle |
| RENAMED | `e2e-aws-tests-bootc-arm-nightly-el9` | `e2e-aws-tests-bootc-arm-nightly-el9-lifecycle` | Added `-lifecycle` |
| RENAMED | `e2e-aws-tests-bootc-arm-nightly-el10` | `e2e-aws-tests-bootc-arm-nightly-el10-lifecycle` | Added `-lifecycle` |
| RENAMED | `e2e-aws-tests-bootc-c2cc-nightly` | `e2e-aws-tests-bootc-x86-nightly-c2cc` | Naming convention |
| RENAMED | `e2e-aws-tests-bootc-c2cc-arm-nightly` | `e2e-aws-tests-bootc-arm-nightly-c2cc` | Reordered |
| RENAMED | `e2e-aws-tests-nightly` | `e2e-aws-tests-ostree-x86-nightly-el9-lifecycle` | Naming convention |
| RENAMED | `e2e-aws-tests-arm-nightly` | `e2e-aws-tests-ostree-arm-nightly-el9-lifecycle` | Naming convention |
| RENAMED | `e2e-aws-tests-release-periodic` | `e2e-aws-tests-ostree-x86-release-periodic-el9` | Naming convention |
| RENAMED | `e2e-aws-tests-release-arm-periodic` | `e2e-aws-tests-ostree-arm-release-periodic-el9` | Naming convention |
| RENAMED | `e2e-aws-tests-bootc-release-periodic-el9` | `e2e-aws-tests-bootc-x86-release-periodic-el9` | Naming convention |
| RENAMED | `e2e-aws-tests-bootc-release-periodic-el10` | `e2e-aws-tests-bootc-x86-release-periodic-el10` | Naming convention |
| RENAMED | `e2e-aws-tests-bootc-release-arm-periodic-el9` | `e2e-aws-tests-bootc-arm-release-periodic-el9` | Reordered |
| RENAMED | `e2e-aws-tests-bootc-release-arm-periodic-el10` | `e2e-aws-tests-bootc-arm-release-periodic-el10` | Reordered |

## Summary

| | main.yaml | periodics.yaml | Total |
|---|---|---|---|
| **New** | 8 | 8 | **16** |
| **Deleted** | 6 | 0 | **6** |
| **Renamed** | 20 | 14 | **34** |
| Unchanged | 13 | 10 | 23 |



⏺ ┌────────────────┬─────────────┬───────────────┬──────────────────────────────────────────────────────┬────────────────────────────────────────────────────────────┬──────────────────────────────────────────┐
  │      File      │   Status    │    Trigger    │                       Old Name                       │                          New Name                          │                  Notes                   │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ NEW         │ PR-auto       │                                                      │ e2e-aws-tests-rpm-presubmit-el{9,10}-arm                   │ RPM presubmit (replaces ostree)          │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ DELETED     │ PR-auto       │ e2e-aws-tests(-arm)?                                 │                                                            │ Ostree presubmit - moved to nightly      │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ DELETED     │ PR-auto       │ e2e-aws-tests-bootc(-arm)?-el{9,10}                  │                                                            │ Bootc presubmit - moved to nightly       │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-auto       │ e2e-aws-tests-bootc-(upstream|c2cc)(-arm)?           │ e2e-aws-tests-bootc-presubmit-(upstream|c2cc)-{x86,arm}    │ Added trigger (+ arch)                   │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-auto       │ e2e-aws-tests-cache                                  │ e2e-aws-tests-cache-x86                                    │ Added -x86                               │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ NEW         │ PR-manually   │                                                      │ e2e-aws-tests-ostree-periodic-el9-functional-{x86,arm}     │ Split from e2e-aws-tests-periodic(-arm)? │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ NEW         │ PR-manually   │                                                      │ e2e-aws-tests-bootc-periodic-el{9,10}-functional-{x86,arm} │ Split from bootc-periodic jobs           │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-manually   │ e2e-aws-tests-periodic(-arm)?                        │ e2e-aws-tests-ostree-periodic-el9-lifecycle-{x86,arm}      │ Lifecycle split + naming convention      │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-manually   │ e2e-aws-tests-bootc-periodic(-arm)?-el{9,10}         │ e2e-aws-tests-bootc-periodic-el{9,10}-lifecycle-{x86,arm}  │ Lifecycle split                          │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-manually   │ e2e-aws-tests-release(-arm)?                         │ e2e-aws-tests-ostree-release-el9-{x86,arm}                 │ Naming convention                        │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-manually   │ e2e-aws-tests-bootc-release(-arm)?-el{9,10}          │ e2e-aws-tests-bootc-release-el{9,10}-{x86,arm}             │ Reordered                                │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ PR-manually   │ e2e-aws-tests-bootc-upstream-periodic                │ e2e-aws-tests-bootc-upstream-periodic-x86                  │ Added -x86                               │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ NEW         │ cron (daily)  │                                                      │ e2e-aws-tests-bootc-{x86,arm}-nightly-el{9,10}-functional  │ Split from bootc nightly jobs            │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ NEW         │ cron (daily)  │                                                      │ e2e-aws-tests-ostree-{x86,arm}-nightly-el9-functional      │ Split from ostree nightly jobs           │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ NEW/RENAMED │ cron (daily)  │ e2e-aws-tests-bootc(-arm)?-nightly-el9               │ e2e-aws-tests-bootc-{x86,arm}-nightly-el{9,10}-lifecycle   │ el9 renamed, el10 new                    │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ RENAMED     │ cron (daily)  │ e2e-aws-tests(-arm)?-nightly                         │ e2e-aws-tests-ostree-{x86,arm}-nightly-el9-lifecycle       │ Naming convention                        │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ RENAMED     │ cron (daily)  │ e2e-aws-tests-bootc-c2cc(-arm)?-nightly              │ e2e-aws-tests-bootc-{x86,arm}-nightly-c2cc                 │ Naming convention                        │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ main.yaml      │ RENAMED     │ cron (daily)  │ e2e-aws-tests-cache-nightly                          │ e2e-aws-tests-cache-nightly-x86                            │ Added -x86                               │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ RENAMED     │ cron (weekly) │ e2e-aws-tests-release(-arm)?-periodic                │ e2e-aws-tests-ostree-{x86,arm}-release-periodic-el9        │ Naming convention                        │
  ├────────────────┼─────────────┼───────────────┼──────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────┼──────────────────────────────────────────┤
  │ periodics.yaml │ RENAMED     │ cron (weekly) │ e2e-aws-tests-bootc-release(-arm)?-periodic-el{9,10} │ e2e-aws-tests-bootc-{x86,arm}-release-periodic-el{9,10}    │ Reordered                                │
  └────────────────┴─────────────┴───────────────┴──────────────────────────────────────────────────────┴────────────────────────────────────────────────────────────┴──────────────────────────────────────────┘


  ⏺ Here's the full updated table:

  ┌────────────────┬───────────────┬───────────────────────────────────────────────────────────────┬─────────────────────────────────────────────┐
  │      File      │    Trigger    │                             Name                              │                    Notes                    │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ e2e-aws-tests-rpm-presubmit-el{9,10}-arm                      │ NEW — RPM presubmit (replaces ostree)       │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ e2e-aws-tests-bootc-presubmit-(upstream|c2cc)-{x86,arm}       │ RENAMED — added trigger + arch              │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ e2e-aws-tests-cache-x86                                       │ RENAMED — added -x86                        │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ e2e-aws-tests-cache-arm                                       │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ e2e-aws-ai-model-serving                                      │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ e2e-aws-footprint-and-performance                             │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ images                                                        │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ ocp-full-conformance(-serial)?-rhel-eus                       │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ security                                                      │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ test-(rebase|rpm|unit)                                        │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-auto       │ verify(-deps)?                                                │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-ostree-periodic-el9-functional-{x86,arm}        │ NEW — split from periodic                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-bootc-periodic-el{9,10}-functional-{x86,arm}    │ NEW — split from bootc-periodic             │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-ostree-periodic-el9-lifecycle-{x86,arm}         │ RENAMED — lifecycle split                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-bootc-periodic-el{9,10}-lifecycle-{x86,arm}     │ RENAMED — lifecycle split                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-ostree-release-el9-{x86,arm}                    │ RENAMED — naming convention                 │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-bootc-release-el{9,10}-{x86,arm}                │ RENAMED — reordered                         │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ main.yaml      │ PR-manually   │ e2e-aws-tests-bootc-upstream-periodic-x86                     │ RENAMED — added -x86                        │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-tests-(ostree|bootc)-{x86,arm}-nightly-el9-functional │ NEW — split from nightly                    │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-tests-bootc-{x86,arm}-nightly-el10-functional         │ NEW — split from bootc nightly              │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-tests-(ostree|bootc)-{x86,arm}-nightly-el9-lifecycle  │ RENAMED — naming convention                 │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-tests-bootc-{x86,arm}-nightly-el10-lifecycle          │ NEW/RENAMED — el9 renamed, el10 new         │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-tests-bootc-{x86,arm}-nightly-c2cc                    │ RENAMED — naming convention                 │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-tests-cache-nightly(-arm)?                            │ -arm unchanged, -x86 renamed from main.yaml │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-ai-model-serving-nightly                              │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ e2e-aws-footprint-and-performance-nightly                     │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ clusterbot-nightly                                            │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (daily)  │ rebase-on-nightlies                                           │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (weekly) │ e2e-aws-tests-ostree-{x86,arm}-release-periodic-el9           │ RENAMED — naming convention                 │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (weekly) │ e2e-aws-tests-bootc-{x86,arm}-release-periodic-el{9,10}       │ RENAMED — reordered                         │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ cron (weekly) │ e2e-aws-ovn-ocp-conformance(-serial)?                         │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ other         │ publish-release-notes                                         │ unchanged                                   │
  ├────────────────┼───────────────┼───────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
  │ periodics.yaml │ other         │ update-versions-releases                                      │ unchanged                                   │
  └────────────────┴───────────────┴───────────────────────────────────────────────────────────────┴─────────────────────────────────────────────┘



⏺ Based on the PR changes, here are all rehearsable presubmit jobs (new or meaningfully changed):

  RPM presubmits (new, ARM):
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-rpm-presubmit-el9-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-rpm-presubmit-el10-arm

  Functional splits (new):
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-ostree-periodic-el9-functional-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-ostree-periodic-el9-functional-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el9-functional-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el10-functional-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el9-functional-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el10-functional-arm

  Lifecycle renames (renamed, but new target names):
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-ostree-periodic-el9-lifecycle-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-ostree-periodic-el9-lifecycle-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el9-lifecycle-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el10-lifecycle-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el9-lifecycle-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-periodic-el10-lifecycle-arm

  Release renames:
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-ostree-release-el9-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-ostree-release-el9-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-release-el9-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-release-el10-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-release-el9-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-release-el10-arm

  Other renames (cache, upstream, c2cc):
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-cache-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-cache-nightly-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-upstream-periodic-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-presubmit-upstream-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-presubmit-upstream-arm
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-presubmit-c2cc-x86
  /pj-rehearse pull-ci-openshift-microshift-main-e2e-aws-tests-bootc-presubmit-c2cc-arm
