# Gateway API
Gateway API is included since MicroShift 4.18 as a dev preview feature, meaning limited suport. In order to use the feature let's see the basic use case to enable it in your applications.
The example listed here is the same as [basic north/south use case](https://gateway-api.sigs.k8s.io/concepts/use-cases/#basic-northsouth-use-case) except for TLS and DNS, not included for simplicity.

> Gateway API is based on OpenShift Service Mesh 3. There is no support for service mesh capabilities outside the ones mentioned here.

## Installing
Gateway API functionality is released as a separate RPM in the same way as other optional components, such as OLM or Multus. To install it you only need to do:
```bash
$ sudo dnf install -y microshift-gateway-api
```
There is also a separate RPM package with the release information:
```bash
$ sudo dnf install -y microshift-gateway-api-release-info
```
For the purpose of this document only the first RPM is required.

> Please note the RPM was released alongside MicroShift 4.18.

## Verifying the install
Once installed you should be able to see the gateway api resources in the api server:
```bash
# Available resources for gateway api functionality. Note that not all of these are used/supported.
$ oc api-resources  | egrep "istio|gateway"
wasmplugins                                             extensions.istio.io/v1alpha1        true         WasmPlugin
gatewayclasses                      gc                  gateway.networking.k8s.io/v1        false        GatewayClass
gateways                            gtw                 gateway.networking.k8s.io/v1        true         Gateway
grpcroutes                                              gateway.networking.k8s.io/v1        true         GRPCRoute
httproutes                                              gateway.networking.k8s.io/v1        true         HTTPRoute
referencegrants                     refgrant            gateway.networking.k8s.io/v1beta1   true         ReferenceGrant
destinationrules                    dr                  networking.istio.io/v1              true         DestinationRule
envoyfilters                                            networking.istio.io/v1alpha3        true         EnvoyFilter
gateways                            gw                  networking.istio.io/v1              true         Gateway
proxyconfigs                                            networking.istio.io/v1beta1         true         ProxyConfig
serviceentries                      se                  networking.istio.io/v1              true         ServiceEntry
sidecars                                                networking.istio.io/v1              true         Sidecar
virtualservices                     vs                  networking.istio.io/v1              true         VirtualService
workloadentries                     we                  networking.istio.io/v1              true         WorkloadEntry
workloadgroups                      wg                  networking.istio.io/v1              true         WorkloadGroup
istiocnis                                               sailoperator.io/v1alpha1            false        IstioCNI
istiorevisions                      istiorev            sailoperator.io/v1alpha1            false        IstioRevision
istios                                                  sailoperator.io/v1alpha1            false        Istio
remoteistios                                            sailoperator.io/v1alpha1            false        RemoteIstio
authorizationpolicies               ap                  security.istio.io/v1                true         AuthorizationPolicy
peerauthentications                 pa                  security.istio.io/v1                true         PeerAuthentication
requestauthentications              ra                  security.istio.io/v1                true         RequestAuthentication
telemetries                         telemetry           telemetry.istio.io/v1               true         Telemetry

# A gateway class for applications to use gateways on
$ oc get gatewayclass openshift-gateway-api -o yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  creationTimestamp: "2024-11-05T13:15:50Z"
  generation: 1
  name: openshift-gateway-api
  resourceVersion: "1518"
  uid: 2490e78e-9000-42cf-8234-d3e5482825a9
spec:
  controllerName: openshift.io/gateway-controller
status:
  conditions:
  - lastTransitionTime: "2024-11-05T13:18:02Z"
    message: Handled by Istio controller
    observedGeneration: 1
    reason: Accepted
    status: "True"
    type: Accepted

# Service mesh operator and istio controller
$ oc get deployment -n openshift-gateway-api
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
istiod-openshift-gateway-api   1/1     1            1           26h
servicemesh-operator3          1/1     1            1           26h
```
There are also additional roles and role bindings, not shown here for brevity.

## Configuring Gateways
In order to use the feature we need an application that we want to expose through a gateway, a gateway, and some routes.

First we create a sample application:
```bash
$ oc create -f test/assets/hello-microshift.yaml 
$ oc create -f test/assets/hello-microshift-service.yaml 
$ oc get pod,svc hello-microshift
NAME                   READY   STATUS    RESTARTS   AGE
pod/hello-microshift   1/1     Running   0          48s

NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/hello-microshift   ClusterIP   10.43.216.183   <none>        8080/TCP   13s
```

Then we can create a gateway for it:
```bash
$ cat gatewayapi/gateway.yaml 
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: demo-gateway
  namespace: openshift-gateway-api
spec:
  gatewayClassName: openshift-gateway-api
  listeners:
  - name: demo
    hostname: "*.microshift-9"
    port: 8080
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: All

$ oc create -f gatewayapi/gateway.yaml 
gateway.gateway.networking.k8s.io/demo-gateway created

# Output truncated to show relevant parts for this doc.
$ oc get gateway -n openshift-gateway-api demo-gateway -o yaml | yq '.status'
addresses:
  - type: IPAddress
    value: 192.168.113.117
conditions:
  - ...
  - lastTransitionTime: "2024-11-06T16:03:20Z"
    message: Resource programmed, assigned to service(s) demo-gateway-openshift-gateway-api.openshift-gateway-api.svc.cluster.local:8080
    observedGeneration: 1
    reason: Programmed
    status: "True"
    type: Programmed
listeners:
  - attachedRoutes: 0
    conditions:
      ...
    name: demo
    supportedKinds:
      ...
```
The gateway defined above is using the GatewayClass that MicroShift created, it includes a listener called `demo` listening in port 8080 for any request with hostname `*.microshift-9` and is able to accept routes from any namespace in the cluster.
Note there are no attached routes to the gateway because we have not created any yet.

The gateway definition will kickstart the creation of the resources needed to accept and route traffic. A deployment (using the gateway name with the gateway class as the resource name), and a service to expose it outside of the cluster:
```bash
$ oc get deployments.apps -n openshift-gateway-api demo-gateway-openshift-gateway-api
NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
demo-gateway-openshift-gateway-api   1/1     1            1           3m12s

$ oc get svc -n openshift-gateway-api demo-gateway-openshift-gateway-api 
NAME                                 TYPE           CLUSTER-IP    EXTERNAL-IP       PORT(S)                          AGE
demo-gateway-openshift-gateway-api   LoadBalancer   10.43.69.88   192.168.113.117   15021:32736/TCP,8080:30877/TCP   2m36s
```

And to be able to use it, a route. Following is a simple route to redirect all traffic from a specific hostname into a service. Note how the status reveals information about whether the route was accepted by a gateway (and which one):
```bash
$ cat gatewayapi/httproute.yaml 
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: http
  namespace: default
spec:
  parentRefs:
  - name: demo-gateway
    namespace: openshift-gateway-api
  hostnames: ["test.microshift-9"]
  rules:
  - backendRefs:
    - name: hello-microshift
      namespace: default
      port: 8080

$ oc create -f  gatewayapi/httproute.yaml 
httproute.gateway.networking.k8s.io/http created

$ oc get httproutes.gateway.networking.k8s.io http -o yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  creationTimestamp: "2024-11-06T16:10:41Z"
  generation: 1
  name: http
  namespace: default
  resourceVersion: "133629"
  uid: ab8deb11-458b-43a8-9ff3-b63e47b3d5c3
spec:
  hostnames:
  - test.microshift-9
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: demo-gateway
    namespace: openshift-gateway-api
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: hello-microshift
      namespace: default
      port: 8080
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /
status:
  parents:
  - conditions:
    - lastTransitionTime: "2024-11-06T16:10:41Z"
      message: Route was valid
      observedGeneration: 1
      reason: Accepted
      status: "True"
      type: Accepted
    - lastTransitionTime: "2024-11-06T16:10:41Z"
      message: All references resolved
      observedGeneration: 1
      reason: ResolvedRefs
      status: "True"
      type: ResolvedRefs
    controllerName: openshift.io/gateway-controller
    parentRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: demo-gateway
      namespace: openshift-gateway-api
```

Note that the route is in a different namespace than the gateway but is still accepted due to its configuration. This is also noted in the gateway status:
```bash
$ oc get gateway -n openshift-gateway-api demo-gateway -o yaml | yq '.status.listeners'
- attachedRoutes: 1
  conditions:
    - lastTransitionTime: "2024-11-06T16:03:18Z"
      message: No errors found
      observedGeneration: 1
      reason: Accepted
      status: "True"
      type: Accepted
    - lastTransitionTime: "2024-11-06T16:03:18Z"
      message: No errors found
      observedGeneration: 1
      reason: NoConflicts
      status: "False"
      type: Conflicted
    - lastTransitionTime: "2024-11-06T16:03:18Z"
      message: No errors found
      observedGeneration: 1
      reason: Programmed
      status: "True"
      type: Programmed
    - lastTransitionTime: "2024-11-06T16:03:18Z"
      message: No errors found
      observedGeneration: 1
      reason: ResolvedRefs
      status: "True"
      type: ResolvedRefs
  name: demo
  supportedKinds:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
    - group: gateway.networking.k8s.io
      kind: GRPCRoute
```
Which is now showing 1 attached route, as in the `parentRef` from the `HTTPRoute` above.

We should now be able to query the service through the gateway:
```bash
$ curl http://test.microshift-9:8080 --connect-to test.microshift-9::192.168.113.117: -i
HTTP/1.1 200 OK
content-length: 16
x-envoy-upstream-service-time: 1
date: Wed, 06 Nov 2024 16:12:30 GMT
server: istio-envoy

Hello MicroShift
```
Note the headers from the gateway (`server` and `x-envoy-upstream-service-time`), proving that the request went through it (a curl to the service IP does not yield those headers). If we were to delete the route we should get a 404:
```bash
$ oc delete -f gatewayapi/httproute.yaml 
httproute.gateway.networking.k8s.io "http" deleted

$ curl http://test.microshift-9:8080 --connect-to test.microshift-9::192.168.113.117: -i
HTTP/1.1 404 Not Found
date: Wed, 06 Nov 2024 16:17:07 GMT
server: istio-envoy
content-length: 0
```

## Uninstall
Gateway API is installed through RPM packages, therefore removing them will erase the files for all the new manifests.
```bash
$ sudo dnf remove -y microshift-gateway-api-release-info microshift-gateway-api
```

However, this is not enough because the resources remain installed in etcd. To completely remove Gateway API resources we need to do the following:
* Remove the user defined `Gateway`, `HTTPRoute` and `GRPCRoute` resources. These depend on your application.
* Remove `openshift-gateway-api` namespace. This will remove all the namespaced resources.
  ```bash
  $ oc delete namespace openshift-gateway-api
  ```
* Remove `ClusterRole` and `ClusterRoleBindings`.
  ```bash
  $ oc get clusterrole | grep -E "openshift-gateway-api|servicemesh-operator" | awk '{print $1}' | xargs oc delete clusterrole
  $ oc delete clusterrolebinding | grep -E "openshift-gateway-api|servicemesh-operator" | awk '{print $1}' | xargs oc delete clusterrolebinding
  ```
* Remove CRD definitions.
  ```bash
  $ oc get crd | grep -E "gateway|istio" | awk '{print $1}' | xargs oc delete crd
  ```

