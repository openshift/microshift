# How To Configure a Workload with Custom Security Context

MicroShift limits the SecurityContextConstraint of new namespaces to
`restricted-v2` by default. This can mean that workloads that run on
OpenShift result in pod security admission errors when run on
MicroShift. This HowTo demonstrates a simple application that uses
escalated privileges defined in a custom SecurityContextConstraint.

## SecurityContextConstraint

This SecurityContextConstraint enables host directory mounts host
ports, and running as any user, among other settings.

```bash
oc apply -f - <<EOF
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: scc-demo
allowHostDirVolumePlugin: true   # allow mounting host directories
allowHostNetwork: true
allowHostPID: true
allowHostPorts: true             # allow using host network ports
allowPrivilegedContainer: true
allowPrivilegeEscalation: true
readOnlyRootFilesystem: false
runAsUser:
  type: RunAsAny                 # allow running as any user
supplementalGroups:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
users: []
EOF
```

## Namespace

This namespace is configured to disable the pod security label syncer
by setting the `security.openshift.io/scc.podSecurityLabelSync` label
to `"false"`. It also has the enforcement, audit, and warning labels
set to `privileged`.

```bash
oc apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: scc-demo
  labels:
    name: scc-demo
    security.openshift.io/scc.podSecurityLabelSync: "false"
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged
EOF
```

## Cluster Role

After the SCC and namespace are created, the next step is to define a
ClusterRole that can grant permission to use the SCC.

```bash
oc apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: scc-demo
  namespace: scc-demo
rules:
- apiGroups:
  - security.openshift.io
  resources:
  - securitycontextconstraints
  resourceNames:
  - scc-demo
  verbs:
  - use
EOF
```

## Service Account

Next, create a ServiceAccount and use a ClusterRoleBinding to give it
the ClusterRole.

```bash
oc apply -f - <<EOF
apiVersion: v1
automountServiceAccountToken: false
kind: ServiceAccount
metadata:
  name: scc-demo
  namespace: scc-demo
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: scc-demo
  namespace: scc-demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: scc-demo
subjects:
- kind: ServiceAccount
  name: scc-demo
  namespace: scc-demo
EOF
```

## Deployment

Finally, define a Deployment that uses the service account and takes
advantage of the extra privileges to run the container as root and
mount the host's root filesystem.

```bash
oc apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: scc-demo
  namespace: scc-demo
spec:
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      privileged: true         # run privileged
      containers:
      - name: busybox
        image: busybox:1.35
        command:
          - sleep
          - "3600"
        securityContext:
          # privileged: true
          runAsUser: 0         # run as root
        volumeMounts:
        - mountPath: /host/root
          mountPropagation: HostToContainer
          name: root
          readOnly: true
      serviceAccountName: scc-demo  # use the service account
      volumes:
      - hostPath:
          path: /
        name: root
EOF
```

## References

* OpenShift Documentation
  * [Important Changes to Pod Security Standards](https://connect.redhat.com/en/blog/important-openshift-changes-pod-security-standards)
  * [Pod Security Admission in OpenShift 4.11](https://cloud.redhat.com/blog/pod-security-admission-in-openshift-4.11)
  * [Complying with pod security admission](https://docs.openshift.com/container-platform/4.13/operators/operator_sdk/osdk-complying-with-psa.html)
  * [About security context constraints](https://docs.openshift.com/container-platform/4.13/authentication/managing-security-context-constraints.html#security-context-constraints-about_configuring-internal-oauth)
  * [Security context constraint synchronization with pod security standards](https://docs.openshift.com/container-platform/4.13/authentication/understanding-and-managing-pod-security-admission.html#security-context-constraints-psa-synchronization_understanding-and-managing-pod-security-admission)
* Kubernetes Documentation
  * [Enforcing Pod Security Standards with Namespace Labels](https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-namespace-labels/)
  * [Important OpenShift changes to Pod Security Standards](https://connect.redhat.com/en/blog/important-openshift-changes-pod-security-standards)
