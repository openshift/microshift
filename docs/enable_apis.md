# Enabled OpenShift APIs

In addition to the standard Kubernetes APIs, MicroShift supports the
following OpenShift API resources.

| GroupVersion          | Kind  |
|-----------------------|-------|
| route.openshift.io/v1 | Route |


| GroupVersion             | Kind                       |
|--------------------------|----------------------------|
| security.openshift.io/v1 | SecurityContextConstraints |


| GroupVersion                  | Kind                      |
|-------------------------------|---------------------------|
| authorization.openshift.io/v1 | RoleBindingRestriction    |
