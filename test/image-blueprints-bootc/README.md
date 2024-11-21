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

The `.container-encapsulate` files depend on the `rpm-ostree` commits built in
the `image-blueprints/layer1-base` layer. These commits must be downloaded from
cache or pre-built locally for `bootc` base layer builds to be successful.

> Note: Total build times are up to 15 minutes.

## Bootc Source Layer

Artifacts built in this layer cannot be cached as they depend on the current sources.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Average  | Current source prerequisites (RHEL and CentOS)
|group2| Average  | Current source artifacts on RHEL
|group3| Average  | Current source artifacts on CentOS (with 1 exception)

The `rhel94-bootc-source-isolated.image-bootc` file is an exception in `group3`
as it uses RHEL. This helps avoid creation of another group with a single file.

> Note: Total build times are up to 15 minutes.
