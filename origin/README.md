Origin MicroShift
=================
This directory is a copy of [origin](https://github.com/openshift/origin)
repository.

Its purpose is twofold:
* Have a stable and controlled openshift-tests binary.
* Add MicroShift specific checks and skips to enable more tests to run as part
  of the OCP conformance suite.

The changes introduced in this copy will be copied as PRs into
[origin](https://github.com/openshift/origin) once they are verified to work
with automatic CI jobs.

This directory is temporary. Once the [origin](https://github.com/openshift/origin)
tests are on par with the same results in CI, this copy will cease to exist.

For more info on which origin version this code is synchronized to, check
[NOTES](https://github.com/openshift/microshift/blob/master/origin/NOTES)

For more info on [origin](https://github.com/openshift/origin), please check
the [README](https://github.com/openshift/origin/blob/master/README.md).
