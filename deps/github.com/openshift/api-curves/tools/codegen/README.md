# Codegen

This tool simplifies the process of running API generation across a large number of API groups and versions.
It's aim is to encode specific knowledge about the API generators we use and make generating OpenShift APIs
easier.

## Usage

The tool can be compiled in the normal way using either `go build` or `make codegen` from the root of the tools module.
When using make the tool will output to `$(TOOLS_MODULE)/_output/bin/$(GOOS)/$(GOARCH)/bin/codegen`.

The root command has the following arguments:
- `--api-group-versions` - A comma separated list of group versions to generate, all groups must be fully qualified.
  e.g. apps.openshift.io/v1,machine.openshift.io/v1,machine.openshift.io/v1beta1.
- `--base-dir` - The path to the root of the API folders, this directory will be recursively searched to find the group
  versions specified in `--api-group-versions`. When no group versions are specified, all discovered group versions
  will be generated.
- `--verify` - Rather than updating existing files, the generators will verify that the existing files are up to date.
  The generator will exit with a non-zero exit code if any file needs to be regenerated.
- `--versions` - Prints the version of the utility and exits.

When using the root command nakedly, all default generators will be executed on the API groups unless a configuration file
is found which disables a particular generator.

As a full example, from the root of the OpenShift API repository, you may run:
```
codegen --base-dir /go/src/github.com/openshift/api --api-group-versions apps.openshift.io/v1,config.openshift.io/v1,operator.openshift.io/v1
```

## Inclusion in other repositories

To use this tool in another repository, you should make sure to add the tools Go submodule to your dependency magnet
or `tools.go` file.

```go
//go:build tools
// +build tools

package tools

import (
  _ "github.com/openshift/api/tools"
  _ "github.com/openshift/api/tools/codegen/cmd"
)
```

You will also need to replace controller-tools with the OpenShift version within your `go.mod`:
```go
replace sigs.k8s.io/controller-tools => github.com/openshift/controller-tools v0.9.3-0.20220912174723-cf3ef054f3dd // v0.9.2+openshift-0.2
```

Then ensure your vendored dependencies are up to date:

```bash
go mod tidy
go mod vendor
```

In your top level Makefile, you can add a target like below, substituting the path to your own APIs base folder and
group versions:
```Make
.PHONY: update-codegen
update-codegen:
  make -C vendor/github.com/openshift/api/tools run-codegen  BASE_DIR="${PWD}/pkg/apis" API_GROUP_VERSIONS="autoscaling.openshift.io/v1,autoscaling.openshift.io/v1beta1"
```

To generate the same group versions but with the TechPreviewNoUpgrade FeatureSet, you would add the FeatureSet to the
end:
```Make
.PHONY: update-codegen-crds
update-codegen-crds:
  make -C vendor/github.com/openshift/api/tools run-codegen-crds  BASE_DIR="${PWD}/pkg/apis" API_GROUP_VERSIONS="autoscaling.openshift.io/v1,autoscaling.openshift.io/v1beta1" OPENSHIFT_REQUIRED_FEATURESETS="TechPreviewNoUpgrade"
```

You may also want to add the `_output` directory to your `.gitignore` to avoid checking in compiled binaries created
by this make target.

## Generators

The following section describes the individual generators included within the `codegen` utility.

The generators enabled by default are:
- [Compatibility](#compatibility)
- [Deepcopy](#deepcopy)
- [Schemapatch](#schemapatch)
- [Swaggerdocs](#swaggerdocs)

### Compatibility

To generate API compatibility comments, use the `compatibility` subcommand.

The `compatibility` subcommand generates a compatibility level comment for each API defintion.
The generation is controlled by a marker applied to the CRD struct defintiion.
For example, this annotation would be `+openshift:compatibility-gen:level=1` for a level 1 API.
	
Valid API levels are 1, 2, 3 and 4. Version 1 is required for all stable APIs.
Version 2 should be used for beta level APIs. Levels 3 and 4 may be used for alpha APIs.

### Deepcopy

To generate Deepcopy functions, use the `deepcopy` subcommand.

The deepcopy subcommand uses the Kubernetes [deepcopy generator](https://github.com/kubernetes/code-generator/tree/master/cmd/deepcopy-gen)
to generate `DeepCopy()` functions for API types.
These are then used in projects to create completely independent copies of resourcres that share
no memory with the original object.

These additoinal arguments may be provider when using the `deepcopy` generator:
- `--deepcopy:header-file-path` - Path to file containing boilerplate header text. The string YEAR will be replaced with the current 4-digit year.
  When omitted, no header is added to the generated files.
- `--deepcopy:output-file-base-name` - Base name of the output file. When omitted, `zz_generated.deepcopy` is used as the default.

### Schemapatch

To generate CRD schemas, use the `schemapatch` subcommand.

The schemapatch subcommand uses the [controller-tools][controller-tools] project to generate CRD schemas.
It requires a stub CRD to exist before the schema will be generated. Copying an existing OpenShift CRD
and updating the metadata is the easiest way to generate a stub.

In addition to generating schemas, it also allows patches to be applied to the schema post generation.
The [yaml-patch][yaml-patch] library accepts RFC-6092-ish patches in a YAML format and applies them to
the CRD.
YAML patch files must be named identically to the CRD they apply to, but with a `yaml-patch` extension
instead of the standard `yaml` extension.

These additional arguments may be provided when using the `schemapatch` generator:
- `--controller-gen` - optionally use a particular controller-gen binary. When not specified, the tool will use the
  built in generator.
  Note, you must use a `controller-gen` built from the [OpenShift fork](https://github.com/openshift/kubernetes-sigs-controller-tools) of controller-tools.
- `--required-feature-sets` - optionally generate based on the OpenShift feature sets annotations.
  This will update only CRDs with a matching value for the `release.openshift.io/feature-set` annotation.

For example, to generate only TechPreviewNoUpgrade versions of CRDs:
```
codegen schemapatch --base-dir /go/src/github.com/openshift/api --api-group-versions apps.openshift.io/v1,config.openshift.io/v1,operator.openshift.io/v1 --require-feature-sets TechPreviewNoUpgrade
```

[controller-tools]: https://github.com/kubernetes-sigs/controller-tools
[yaml-patch]: https://github.com/vmware-archive/yaml-patch

### Swaggerdocs

To generate SwaggerDoc functions for types within an API, use the `swaggerdocs` subcommand.

This generator inspects the documentation within the API defintions and generates functions
that return the Go documenation for each struct and field within an API.

The generator can also be used to verify that each API field and struct has appropriate
documentation, using the `--swagger:enforce-comments` and `--verify` flags in conjunction.

These additional arguments may be provided when using the `swaggerdocs` generator:
- `--swagger:output-file-name` -  Defines the file name to use for the swagger generated docs for each group version.
  When omitted, defaults to `zz_generated.swagger_doc_generated.go`.
- `--swagger:comment-policy` - Defines the policy to use when a field is missing documentation. Valid values are 'Ignore', 'Warn' and 'Enforce'.
  The default policy is 'Warn'. Missing comments will be ignored when the policy is set to 'Ignore', a warning will be produced when the policy is set to 'Warn',
  and the generator will fail when the policy is set to 'Enforce'. Only effective when combined with `--verify`. 

## Using configuration files

The `codegen` utility will search for a file called `.codegen.yaml` at the API group level.
These files allow individual API groups to configure the generation of their paritcular API files.

It also enables the enablement and disablement of generators with a goal that the usage of the generator
should require only the `--base-dir` argument.

The schema of the configuration file is outlined below, all fields are optional and default values will apply
when they are omitted.

Each generator option relates to a flag attached to the generator, review the [generators](#generators) section
for more details on each option within the config file structure.

```yaml
# Configuration for the compatibility generator.
compatibility:
  disabled: false # Defaults to false, set to true to disable.
# Configuration for the deepcopy generator.
deepcopy:
  disabled: false # Defaults to false, set to true to disable.
  headerFilePath: "" # Defaults to empty, no header will be appended.
  outputFileBaseName: zz_generated.deepcopy # Change this if you want to rename the output file. Do not include a file type suffix.
# Configuration for the schemapatch generator.
schemapatch:
  disabled: false # Defaults to true, set to false to disable.
  requiredFeatureSets: # Each entry will be matched against the value of the required feature set annotation.
  - "" # This matches any manifest that does not have the required feature set annotation.
  - "Default"
  - "TechPreviewNoUpgrade"
# Configuration for the swaggerdocs generator.
swaggerdocs:
  disabled: false # Defaults to false, set to true to disable.
  commentPolicy: Warn # Valid values are `Ignore`, `Warn` and `Enforce`.
  outputFileName: zz_generated.swagger_doc_generated.go # Change this if you want to rename the output file.
```
