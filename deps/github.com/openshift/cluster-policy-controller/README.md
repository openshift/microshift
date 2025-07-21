## cluster-policy-controller
The cluster-policy-controller is responsible for maintaining policy resources necessary to create pods in a cluster. 
Controllers managed by cluster-policy-controller are:
* cluster quota reconcilion - manages cluster quota usage
* namespace SCC allocation controller - allocates UIDs and SELinux labels for namespaces     
* cluster csr approver controller - csr approver for monitoring scraping
* podsecurity admission label syncer controller - configure the PodSecurity admission namespace label for namespaces with "security.openshift.io/scc.podSecurityLabelSync: true" label

The `cluster-policy-controller` runs as a container in the `openshift-kube-controller-manager namespace`, in the kube-controller-manager static pod.
This pod is defined and managed by the [`kube-controller-manager`](https://github.com/openshift/cluster-kube-controller-manager-operator/)
[OpenShift ClusterOperator](https://github.com/openshift/enhancements/blob/master/enhancements/dev-guide/operators.md#what-is-an-openshift-clusteroperator).
that installs and maintains the KubeControllerManager [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) in a cluster.  It can be viewed with:     
```
oc get clusteroperator kube-controller-manager -o yaml
```

Many OpenShift ClusterOperators and Operands share common build, test, deployment, and update methods.    
For more information about how to build, deploy, test, update, and develop OpenShift ClusterOperators, see      
[OpenShift ClusterOperator and Operand Developer Document](https://github.com/openshift/enhancements/blob/master/enhancements/dev-guide/operators.md#how-do-i-buildupdateverifyrun-unit-tests)

This section explains how to deploy OpenShift with your test `cluster-kube-controller-manager-operator` and `cluster-policy-controller` images:        
[Testing a ClusterOperator/Operand image in a cluster](https://github.com/openshift/enhancements/blob/master/enhancements/dev-guide/operators.md#how-can-i-test-changes-to-an-openshift-operatoroperandrelease-component)
