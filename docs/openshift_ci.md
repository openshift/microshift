## PoV: MicroShift Contributor
### CI Operations Overview
Openshift-CI will conditionally run end-to-end (e2e) testing whenever a PR is opened or updated.  Depending on the changes introduced, only unit and validation checks may be run. This is determined per test job according to patterns for the `run_if_changed` key defined for each test ([see](https://GitHub.com/openshift/release/blob/9c9a3e99c5985eec57464ea137e8db2534df5f1f/ci-operator/config/redhat-et/microshift/redhat-et-microshift-main.yaml#L114)).  As a rule of thumb, changes that affect the MicroShift compile-time or runtime (i.e. changes to source code, build scripts, or container image definitions) will trigger e2e test runs.

CI jobs are represented by a GitHub status check on PRs and will report a pending/pass/fail state per job.  Jobs are loosely broken up by their maintaining OKD/K8S SIG, which is embedded in the status check’s name.  Failed jobs will toggle their respective git status as failed and block the PR from merging.  Blocks may still be overridden by maintainers.

E2e jobs are executed in parallel, each against its own isolated MicroShift cluster.  That is, there are as many MicroShift clusters deployed as there are SIG status checks.  Expect a complete PR test run to take approximately 45 minutes.

### Debugging
> Test runtime objects (namespace, pods, build images) are garbage collected after 1 hour from test completion.  The build log remains, but does not provide a holistic view of the job’s life.

Prow/Openshift-CI provide CI bot management via GitHub comments. See [here](https://github.com/kubernetes/test-infra/blob/c341b223083a7e5766be99620eba58cc3c8142f1/prow/jobs.md#triggering-jobs-with-comments) for a complete list. Typically all you need is `/retest` to rerun Prow CI jobs.  

CI related GitHub statuses can be identified by their `ci/prow/*` prefix.  When a test fails, the first step should always be to check the status `Details` link.

![Details](./images/openshift_ci_details.png)

The `Details` link will bring you to that test’s combined logs for the build, deploy, test, and teardown phases for that e2e job.  Most of the time, this log can be invaluable for debugging test errors and exposing regressions. 

Occasionally, deeper investigation is warranted.  Errors can be masked in the logs for a variety of reasons, or CI is experiencing some instability (not uncommon) that isn’t immediately obvious in the build log.  

### Debugging CI Cluster Workloads
> Only the PR author is authorized to access a job’s OCP namespace and build artifacts.

OpenShift GitHub org members may gain access to the OCP namespace and backend workloads.  To do this, you must first ensure you’ve completed the [OpenShift onboarding checklist for GitHub](https://source.redhat.com/groups/public/atomicopenshift/atomicopenshift_wiki/openshift_onboarding_checklist_for_github).

To access the OCP CI Cluster namespace for a given GitHub status/e2e suite, go to the status Details link and fully expand the build log.  Each build log prints a URL to the OCP Console namespace for that CI run, near the very top of the log.

![Build Log](./images/openshift_ci_buildlog.png)

The console will ask you to login with either a `kubeadmin` or company SSO. Select Company SSO. After authenticating, you’ll be presented with the overview page for the OCP project.  Expand the left-hand sidebar menu and select `Workloads -> Pods`.  
> You may have to toggle the Developer view to Administrator, located at the top of the sidebar.   

Openshift-CI uses [Prow](https://github.com/kubernetes/test-infra/tree/master/prow#) under the hood to drive CI operations.  Without going too into detail, a CI job is divided into a sequence of steps, executed in parallel if possible.  Each step is handled by a pod, named for the step’s CI configuration.  Logs of the “test” step can always be found in the pod named `e2e-openshift-conformance-sig-[COMPONENT]-openshift-microshift-e2e-run`. Select that pod and examine its logs for the raw test output.

### Examining CI Build Artifacts
When logs aren’t enough to expose a failure’s cause, it may be useful to examine the job’s container artifacts. It is possible to access the CI registry and pull down the job’s images. First, login to the CI cluster by following the steps in [Debugging CI Cluster Workloads].  At the console page, click your name in the upper right corner and select `Copy Login Command`. 
>You may be asked to authenticate again.

![Copy login command](./images/openshift_ci_copylogin.png)

Click `Display Token` and copy and execute the `oc login --token=...` command. Then login to the CI registry with `oc registry login`.

CI build imageStreams are named `pipeline:bin`. Pull the image down:
```
podman pull registry.build02.ci.openshift.org/[PROJECT]/pipeline:bin
```

> The `build02` may not be the current build cluster. This may be changed by OpenShift CI maintainers without warning. 
