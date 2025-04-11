The following guiding principles should be used when adding more artifacts to
each layer or group under `test/image-blueprints-bootc` directory.

> Important: Keep balanced build times within each group and maximize caching
> of artifacts independent of the current source code.

## Bootc Base Layer

Artifacts built in this layer are cached.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Short    | Basic prerequisites
|group2| Long     | Artifacts independent of current sources

The `y-2` and `y-1` upgrade tests depend on `ostree` commits when running
scenarios, not during container image builds. These commits must be downloaded
from cache locally for all `bootc` test scenarios to execute successfully.

> Note: Total build times are up to 10 minutes.

## Bootc Presubmit layer

Artifacts built in this layer cannot be cached as they depend on the current sources.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Short    | Current source prerequisites (RHEL and CentOS) used in presubmits and periodics
|group2| Short    | Current source artifacts on RHEL used in presubmits and periodics

> Note: Total build times are up to 5 minutes.

## Bootc Periodic Layer

Artifacts built in this layer cannot be cached as they depend on the current sources.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Average  | Current source prerequisites (RHEL and CentOS) used only in periodics
|group2| Average  | Current source artifacts on RHEL used only in periodics
|group3| Average  | Current source artifacts on CentOS used only in periodics

> Note: Total build times are up to 15 minutes.
