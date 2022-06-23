# Enabled OpenShift APIs
MicroShift supports the following sets of OpenShift API resources. 

| GroupVersion          | Kind  |
|-----------------------|-------|
| route.openshift.io/v1 | Route |


| GroupVersion             | Kind                       |
|--------------------------|----------------------------|
| security.openshift.io/v1 | SecurityContextConstraints |


| GroupVersion                  | Kind                      |
|-------------------------------|---------------------------|
| authorization.openshift.io/v1 | ClusterRoleBinding        |
|                               | ClusterRole               |
|                               | LocalResourceAccessReview |
|                               | LocalSubjectAccessReview  |
|                               | ResourceAccessReview      |
|                               | RoleBindingRestriction    |
|                               | RoleBinding               |
|                               | Role                      |
|                               | SelfSubjectRulesReview    |
|                               | SubjectAccessReview       |
|                               | SubjectRulesReview        |
