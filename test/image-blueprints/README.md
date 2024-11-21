The following guiding principles should be used when adding more artifacts to
each layer or group under `test/image-blueprints` directory.

> Important: Keep balanced build times within each group and maximize caching
> of artifacts independent of the current source code.

## OSTree Base Layer

Artifacts built in this layer are cached.

Groups 1-to-3 enforce an ordered build chain, necessary to satisfy a mandatory
layer dependency of `rhel94 os-only -> rhel94 y-2 -> rhel94 y-1`, which is needed
for testing upgrades.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Short    | RHEL 9.4 OS-only base layer
|group2| Short    | RHEL 9.4 layer with MicroShift `y-2` packages
|group3| Short    | RHEL 9.4 layer with MicroShift `y-1` packages
|group4| Average  | Other artifacts independent of current sources

> Note: Total build times are up to 30 minutes.

## OSTree Presubmit Layer

Artifacts built in this layer cannot be cached as they depend on the current sources.
The artifacts are used by pre-submit and periodic CI jobs.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Average  | Current source prerequisites used in pre-submits and periodics

> Note: Total build times are up to 10 minutes.

## OSTree Periodic Layer

Artifacts built in this layer cannot be cached as they depend on the current sources.
The artifacts are only used by periodic CI jobs.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Average  | Current source prerequisites used only in periodics

> Note: Total build times are up to 15 minutes.
