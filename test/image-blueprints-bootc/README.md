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

## Bootc Source Layer

Artifacts built in this layer cannot be cached as they depend on the current sources.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Average  | Current source prerequisites (RHEL and CentOS)
|group2| Average  | Current source artifacts on RHEL
|group3| Average  | Current source artifacts on CentOS (with 1 exception)

One of the isolated source image files is an exception in `group3` as it uses
RHEL. This helps avoid creation of another group with a single file.

> Note: Total build times are up to 10 minutes.
