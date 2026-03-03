# API Integration tests

This module provides an integration test suite that can be driven by test suite definitions alongside
our API definitions.

The aim of this suite is to allow testing of API validations pre-merge, and then continously as changes
are made to the API to ensure we catch any breaking changes before those changes are merged.

## How is the suite configured?

The test suite looks through the openshift API repository for test suite files (`<something>.testsuite.yaml`)
which contain test definition for a particular API.

We expect for stable APIs, a file would exist alongside the types called `stable.<api-object-name>.testsuite.yaml`.
For tech preview APIs, the file is instead `techpreview.<api-object-name>.testsuite.yaml`.
You may also include additional test suite files to break down specific test suite cases, in which case you can
add additional context after the object name, before the testsuite suffix.

As an example, for the Machine API type `ControlPlaneMachineSet`, it's suite will be called `stable.controlplanemachineset.testsuite.yaml`.
This API also include AWS specific test cases which are in a file called `stable.controlplanemachineset.aws.testsuite.yaml`.

Each API type within OpenShift API requires at least one test suite file, the verify step will fail if one is not configured.

## How does the suite work?

The suite loads all of the test suites configured in the API repo.
It then iterates through these, installing the relevant CRD into a temporary API server provided by the
[envtest](https://book.kubebuilder.io/reference/envtest.html) project.

The individual test cases are then executed against this API server to check the success/failure
scenarios defined within each suite.

## Writing a test suite

To write a test suite, first you'll need to set up the basic configuration for the suite.
Once you have the basic configuration provided, you can start adding individual test cases.

There are two types of tests cases available, `onCreate` and `onUpdate` test cases.

### Basic configuration

Create a file alongside your API `types.go` file with the expected name.
For an API object called `foo` (ie `oc get foo`), the expected format would be:
- For a stable API, `stable.foo.testsuite.yaml`
- For a tech preview API, `techpreview.foo.testsuite.yaml`

Some APIs may need both stable and techpreview suites.

Add the basic configuration to this file for the test suite:
```yaml
apiVersion: apiextensions.k8s.io/v1 # Hack because controller-gen complains if we don't have this
name: "[Stable] Foo"
crd: foo.crd.yaml
```
- Include the `apiVersion` line as this is required due to a bug in controller-gen
- Include the `name` with a `[Stable]` or `[TechPreview]` label as appropriate. The name should be the name of the API type.
- Include the `crd` with the name of the CRD defintion file. This should be in the same folder as your API and test suite.

This sets up the configuration for the base test suite.

A script is available that can set up the basic suite for you.
```bash
./tests/hack/gen-minimal-test.sh $FOLDER $VERSION
```

Substitute the folder and API version as appropriate.

### OnCreate test cases

OnCreate tests allow you to install a CRD and check either for the persisted data,
or a validation error during the create operation.

Each API version should have a minimal test that creates the object
using the absolute minimum configuration, and checks for defaulted fields.

For example, this might look like:
```yaml
tests:
  onCreate:
  - name: Should be able to create a minimal Foo
    initial: |
      apiVersion: example.openshift.io/v1 # API version should be provided
      kind: Foo # Kind should be provided
      spec:
        requiredField: blah # This is the minimal requirement for this API
    expected: |
      apiVersion: example.openshift.io/v1
      kind: Foo
      spec:
        requiredField: blah
        defaultInt: 3 # This field is defaulted by OpenAPI
```

The above test case shows a successful create of the object.
The `initial` object is the object that is sent via a create call to the API server.
The `expected` object is the object that is compared against the stored object in the API server.

In this case, because the schema defaults the `defaultInt` field, this is added to the `expected`
object in addition to the `initial` schema.

Note, no `metadata` is required. The `name` and `namespace` for the test object will be generated
when they are not provided.

#### Checking validation failures on create

Most validations we write for OpenShift APIs can be tested when creating the object.
This is because the validations do not rely on the previous state of the object.
For example, pattern validations for fields or maximum and minimum values are static
and therefore fail validation on both create and update operators.

The most concise way to add a test for these validations is to add a failure based
OnCreate style test.

For example, in an API that has a field that must be a positive integer, we can write
a test that checks the field validation.

```go
type MyAPI struct {
    // positiveInt is an integer value that must be positive.
    // +kubebuilder:validation:Minimum=0
    // +optional
    PositiveInt int `json:"positiveInt,omitempty"`
}
```

The test for this field would be:

```yaml
tests:
  onCreate:
  - name: Should not allow negative values for positiveInt
    initial: |
      apiVersion: example.openshift.io/v1 # API version should be provided
      kind: Foo # Kind should be provided
      spec:
        positiveInt: -1
    expectedError: "spec.positiveInt: Invalid value: -1: spec.positiveInt in body should be greater than or equal to 0"
```

This case differs from the previous example in that the `expectedError` field specifies a substring within the error
expected to be returned by the API server.
This should contain the portion of the API error that is specific to the test case you are implementing.

As the API error contains strings that vary (eg the resource name), we use a substring match for the error here.

The easiest way to work out what the value for the `expectedError`, is to use a dummy value and execute the test.
For example, using `expectedError: TODO` works well and while the test will fail, the output of the test failure
will show you the errorfrom the API server.
Extract from this the appropriate value for the `expectedError` and update the test case.

### OnUpdate tests

OnUpdate tests are similar to the OnCreate tests however, they are used to test validations that only apply to
updates. For example, they are useful for testing mutation between one state and another, or for testing
immutability of a particular API field.

For a valid update, the test case will look something like:

```yaml
tests:
  onUpdate:
  - name: Should be able to update the required field
    initial: |
      apiVersion: example.openshift.io/v1 # API version should be provided
      kind: Foo # Kind should be provided
      spec:
        requiredField: blah
    updated: |
      apiVersion: example.openshift.io/v1 # API version should be provided
      kind: Foo # Kind should be provided
      spec:
        requiredField: bar
    expected: |
      apiVersion: example.openshift.io/v1
      kind: Foo
      spec:
        requiredField: bar
```

This test differs from the OnCreate test in a few ways.
Firstly the initial object must be valid and any errors in creating the initial object will cause the test case
to fail.
Secondly, a new field, `updated` is introduced. This should represent the update you wish to attempt.
The test will create the `initial` object, fetch it, apply the `update` and then compare the stored, updated object
against the `expected` object.

For an invalid update, the case will look something like:

```yaml
tests:
  onUpdate:
  - name: Should be able to update the required field
    initial: |
      apiVersion: example.openshift.io/v1 # API version should be provided
      kind: Foo # Kind should be provided
      spec:
        immutableField: blah
    updated: |
      apiVersion: example.openshift.io/v1 # API version should be provided
      kind: Foo # Kind should be provided
      spec:
        immutableField: bar
    expectedError: "spec.immutableField: Invalid value: \"string\": immutableField is immutable"
```

This is similar again to the invalid test cases for the OnCreate style tests, but the `updated`
object in this case is the invalid change. The `expectedError` works in the same way.

Again, when checking invalid updates, the `initial` object must be valid.

#### Testing status

Because objects cannot be created with status fields when a `status` subresource is used, to test
status you must use an update test.

The test suite will create any object in the `initial` field and then, if any status is present,
update the object with the status defined in the `initial` field. This means the status must be
valid within the `initial` field.

It will then apply the `updated` field, first by `spec`, and then by `status`.

To define a `status` update error, use `expectedStatusError` instead of `expectedError`, otherwise
the behaviour is the same.

A full example is included below:
```yaml
- name: Should not allow changing an immutable status field
    initial: |
      apiVersion: example.openshift.io/v1
      kind: Foo
      status:
        immutableField: foo
    updated: |
      apiVersion: example.openshift.io/v1
      kind: StableConfigType
      status:
        immutableField: bar
    expectedStatusError: "status.immutableField: Invalid value: \"string\": immutableField is immutable"
```

#### Testing validation ratcheting

Kubernetes now supports [validation ratcheting][validation-ratcheting].
This means that we can now evolve the validations of APIs over time, without immediately breaking stored APIs.
Note, any changes to validations that may be breaking still need careful consideration, this is not a 'get out of jail free' card.

Ratcheting can be tested using `onUpdate` tests and the `initialCRDPatches` option.
`initialCRDPatches` is a list of JSON Object Patches ([RFC6902][rfc6902]) that will be applied temporarily to the CRD prior to the
initial object being created.
This allows you to revert a newer change to an API validation, apply an object that would be invalid with the newer validation,
and then test how the object behaves with the new, current schema.

For example, if a field does not include a maximum, and we decide to enforce a new maximum, a patch such like below could be used
to remove the maximum temporarily to then create an object which exceeds the new maximum.

```
onUpdate:
- name: ...
  initialCRDPatches:
      - op: remove
        path: /spec/versions/0/schema/openAPIV3Schema/properties/spec/properties/fieldWithNewMaxLength/maximum # Maximum was not originally set
  ...
```

Once the patch is applied, three tests should be added for each ratcheting validation:
1. Test that other values can be updated, while the invalid value is persisted
1. Test that the value itself cannot be updated to an alternative, newly invalid value (e.g. making the value longer in this case)
1. Test that the value can be updated to a valid value (e.g. in this case, a value shorted than the new maximum)

[validation-ratcheting]: https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#validation-ratcheting
[rfc6902]: https://datatracker.ietf.org/doc/html/rfc6902
