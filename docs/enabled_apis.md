# Enabled OpenShift APIs

In addition to the standard Kubernetes APIs, MicroShift supports the following OpenShift API resources.

| GroupVersion             | Kind  |
|--------------------------|-------|
| route.openshift.io/v1    | Route |
| security.openshift.io/v1 | SecurityContextConstraints |

When an unsupported API is used via the `oc` command line utility, an error message is generated about a resource that cannot be found. The following listing demonstrates typical error messages generated in this case.

```
$ oc new-project test
Error from server (NotFound): the server could not find the requested resource (get projectrequests.project.openshift.io)

$ oc get projects
error: the server doesn't have a resource type "projects"
```
