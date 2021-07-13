Component Loader Design
=======================

# Background
MicroShift is designed to provide a turnkey solution that can readily serve the target workloads 
with little intervention, probably in an airgapped environment. This requires MicroShift to
be able to provides the most, if not all, infrastructure components during bootstrapping. 

In its current form, MicroShift has an predefined asset repo and a sequence of 
loading the manifests to the cluster. Although new assets can be added programmatically, 
this workload clearly suffers from scalability and flexibilty. 

Specifically, currently the core components are: Service-ca, OpenShift DNS, OpenShift Router, and HostPath provisioner.
For non-interactive use cases, few of these components are required. While for Data Services oriented workloads 
may prefer a more reliable and scale-out storage is beyond the capacity of the HostPath volume. 

So it is clear that MicroShift must have a way to customize which components to load during bootstrapping and normal startup.

# Existing Solutions

## Operator
Operator aims to manage the entire lifecycles of complicated components. Although Operator solves the issue
of managing inter-dependecy, it is deemed heavier than the MicroShift use case. In many perceived occasions, the MicroShift clusters are deployed just once: they are either stay on offline mode for life (e.g. an Edge device deployed in isolated environment) or destroyed without care (e.g. a developer's test environment). There is no
need to keep the Operator running to keep the API objects in the desired states. On the other hand, Operators' all-on pattern does not serve MicroShift's small footprint design goal.

## Kubernetes Addon
Kubernetes Addon mechanism loads cluster wide infrastructure after cluster initialization. Addons such as networking, storage, and management are curated in the Kubernetes repo. Such arrangement, however, does not fit 
into the MicroShift use case:
- MicroShift may need more than one components. For instance, Ceph bucket notification, a use case for processing
Object Store objects on demand, involves Rook/Ceph and Knative (Eventing, Serving, and Networking). 
- These components are often inter-dependent. Kubernetes Addons cannot ensure the components are created or started in the desired order.

## Helm
Helm packages multiple components in a chart form and deploy them in a declarative manner. Once deployed, Helm exists till the next it upgrades or removes the chart from the cluster. The non-interactive and low footprint features fit into MicroShift use well. Still the chart must be accessible by the MicroShift cluster when operating
in an airgapped environment.

# Design
As discussed, MicroShift Component Loader must be able to:
- Customizable for different workloads targeting varying use cases
- Managed multiple, sometimes inter-dependent component
- Allow MicroShift to invoke when the cluster is ready
- Allow MicroShift to monitor readiness status
- Low, preferrably zero footprint
- Able to operate in offline mode

## Component Loader Workflow

When MicroShift starts up the cluster, it reads from the configuration file and finds the Container image that Component Load Job (i.e. running as a Kubernetes Job) uses, constructs a Kubernetes Job using this Container images, creates one ConfigMap for the Job to report progress, creates another ConfigMap for parameters required by the Job (if exists), then starts and watches the Job.

When the Job starts, it reads its configuration from the parameter ConfigMap, creates components, reports readiness status by updating the predefined ConfigMap.

Once the Job exits, MicroShift catches the status and checks the ConfigMap and gets the status.

# Issues

## Repair Components
It is likely that after startup, some of the components, such as Pods and Deployments, degrade or are even killed. To cope with this case, the Component Loader can be started as a CronJob so it can start periodically to repair the components.

## Offline Operation Support

### Container Image
The Container images of the Component Loader and the compnonents it created are packaged or pre-downloaded when integrated with MicroShift.

### Reuse Helm Chart
Similarly, if Helm Charts are used as or by a Component Loader, the Charts are also locally stored together with the airgapped Container images.




