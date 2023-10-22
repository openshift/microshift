# Using Advanced Cluster Security with MicroShift

Red Hat Advanced Cluster Security (RHACS) for Kubernetes helps improve the
security of the application build process, protects the application platform
and configurations, detects runtime issues, and facilitates response.

RHACS Cloud Service provides built-in controls for enforcement to reduce
operational risk, and uses a Kubernetes-native approach that supports built-in
security across the entire software development life cycle, facilitating greater
developer productivity.

See the following documents for more information:
* [RHACS Overview](https://console.redhat.com/application-services/acs/overview)
* [RHACS Product Documentation](https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_security_for_kubernetes/)

The remainder of this document explains how RHACS Cloud Service can be used with MicroShift.

## Create ACS Cloud Instance

Open the [ACS Instances](https://console.redhat.com/application-services/acs/instances)
page and click on `Create ACS Instance` to open the creation form. Fill in the
required values and click on `Create Instance` to initiate the process.

It may take a few minutes until the cloud instance creation is complete. Open the
instance page and save the following instance details:
* Central API endpoint (Sensor mTLS)
* Central instance (UI, roxctl)

> The `Central API endpoint` is required for registering MicroShift clusters with
> the cloud service.

Open the `Central instance` URL and sign in using your Red Hat SSO credentials.

## Download Cluster Init Bundle

Log into the `Central instance` URL and click on the `Manage Tokens` button to
open the `Authenticaton Tokens` integration.

Click on `Cluster Init Bundle` and then on the `Generate bundle` button to start
the wizard. Enter the bundle name and click `Generate`. Once the generation is
complete, click on the `Download Helm values file` button to download the init
bundle file named `<acs_instance_name>-cluster-init-bundle.yaml`.

> Securely store the init bundle file as it contains sensitive information.

## Register MicroShift with ACS

Log into a host with `Helm` installation and access to a MicroShift instance
to be registered with RHACS.

```
$ helm version
version.BuildInfo{Version:"v3.13.1", GitCommit:"3547a4b5bf5edb5478ce352e18858d8a552a4110", GitTreeState:"clean", GoVersion:"go1.20.8"}

$ oc get nodes
NAME                    STATUS   ROLES                         AGE   VERSION
microshift-dev-rhel92   Ready    control-plane,master,worker   40s   v1.27.4
```

### Add Helm Chart Repository

Add and update the `Helm` RHACS charts repository by running the following
commands.

```
helm repo add rhacs https://mirror.openshift.com/pub/rhacs/charts/
helm repo update rhacs
```

Verify that the chart repository was properly added.

```
helm search repo -l rhacs/
```

### Configure Cluster SCC

Run the following commands to configure MicroShift cluster SCC settings for
RHACS pod security.

```
oc adm policy add-scc-to-group privileged system:authenticated system:serviceaccounts
oc adm policy add-scc-to-group anyuid     system:authenticated system:serviceaccounts
```

### MicroShift Cluster Registration

Set variables denoting the information about your environment.

```
HELM_BUNDLE=microshift-cluster-init-bundle.yaml
CLUSTER_NAME=microshift
CENTRAL_ENDPOINT=acs-data-ckj5lnnjhq9and1rulng.acs.rhcloud.com:443
```

Run the following command to register your cluster with RHACS.

```
helm install -n stackrox \
    --create-namespace stackrox-secured-cluster-services rhacs/secured-cluster-services \
    -f "${HELM_BUNDLE}" \
    --set clusterName="${CLUSTER_NAME}" \
    --set centralEndpoint="${CENTRAL_ENDPOINT}"
```

> Examine the `Helm` command output to make sure no warnings
> or errors are reported.

Watch the `stackrox` namespace to make sure all the pods are running
successfully.

```
watch oc get pods -n stackrox
```

> Regular MicroShift troubleshooting techniques apply in case any problems
> are encountered. Start by examining the `sensor` deployment and pod data
> before other RHACS configuration.

## View ACS Security Information

Open the `Central instance` URL and sign in using your Red Hat SSO credentials.

Click on `Platform Configuration > Clusters` to make sure your cluster was
properly registered. Continue to `Dashboard` and other RHACS pages to review
the security information collected by the tool.

> Allow for some time for the information collection to be complete.
