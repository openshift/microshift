The following guiding principles should be used when adding more artifacts to
each layer or group under `test/image-blueprints-bootc` directory.

> Important: Keep balanced build times within each group and maximize caching
> of artifacts independent of the current source code.

## Bootc Base Layer

Artifacts build in this layer are cached.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Short    | Basic prerequisites
|group2| Long     | Artifacts independent of current sources

## Bootc Source Layer

Artifacts build in this layer cannot be cached as they depend on the current sources.

|Group |Build Time|Description|
|------|----------|-----------|
|group1| Average  | Current source prerequisites (RHEL and CentOS)
|group2| Average  | Current source artifacts on RHEL
|group3| Average  | Current source artifacts on CentOS (with 1 exception)

> Note: The `rhel94-bootc-source-isolated.image-bootc` file is an exception in
> `group4` as it uses RHEL. This is done not to create a group with a single
> file.
