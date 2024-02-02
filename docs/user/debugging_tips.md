# Debugging Tips

## Checking the MicroShift Version

From the command line, use `microshift version` to check the version
information.

```bash
$ microshift version
MicroShift Version: 4.10.0-0.microshift-e6980e25
Base OCP Version: 4.10.18
```

Through the API, access the `kube-public/microshift-version` ConfigMap
to retrieve the same information.

```bash
$ oc get configmap -n kube-public microshift-version -o yaml
apiVersion: v1
data:
  major: "4"
  minor: "10"
  version: 4.10.0-0.microshift-e6980e25
kind: ConfigMap
metadata:
  creationTimestamp: "2022-08-08T21:06:11Z"
  name: microshift-version
  namespace: kube-public
```

## Checking the LVMS Version

Like the MicroShift version, the LVM version is available via a configmap. To get the version, run:

```bash
$ oc get configmap -n kube-public lvms-version -ojsonpath='{.data.version}'
```

## Generating an SOS Report

The MicroShift RPMs have an explicit dependency on the `sos` utility allowing to collect
configuration, diagnostic, and troubleshooting data to be provided to Red Hat Technical Support.

> See [Generating sos reports for technical support](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/getting_the_most_from_your_support_experience/generating-an-sos-report-for-technical-support_getting-the-most-from-your-support-experience) for more information on the `sos` utility usage.

> See [documentation](https://github.com/sosreport/sos) for further information on `sos` utility.

Sos tool is composed of different plugins for tailored retrieval of information from different applications.
A MicroShift specific plugin has been added from sos version 4.5.1, and it can gather the following data:
* MicroShift configuration and version
* YAML output for cluster-wide and system namespaced resources
* OVNK information

Plugin is automatically enabled as soon as it detects any of the following conditions:
* MicroShift RPMs are installed.
* MicroShift systemd service is present, whether running or not.
* MicroShift ovnk pods are running.

For more information on the plugin design and output format, please check the [enhancement proposal](https://github.com/openshift/enhancements/blob/master/enhancements/microshift/microshift-supportability-tools.md).

The plugin is also available through the usual options to enable/disable them individually or by profile.
```bash
$ sos report --list-profiles

sosreport (version 4.3)

The following profiles are available:

...
 microshift      microshift, microshift_ovn
...

 24 profiles, 91 plugins
```

Log into the host running MicroShift and execute the following command to generate a report (in this example only 
microshift plugins are enabled. Your system's output may vary):

```bash
$ sudo microshift-sos-report

sosreport (version 4.5.6)


Your sosreport has been generated and saved in:
        /tmp/sosreport-microshift-2-2023-03-23-ylwbkjc.tar.xz

 Size   7.74MiB
 Owner  root
 sha256 850ecd95897441e0ed6ff4595a0e2d46aaa5582b67ce84b32625041498dd0e1d

Please send this file to your support representative.

```
> Sos must always run with root privileges.

The output file and its checksum are generated in `/var/tmp/` directory.
```bash
$ sudo ls -tr /var/tmp/sosreport-* | tail -2
/var/tmp/sosreport-microshift-2-2023-03-23-ylwbkjc.tar.xz
/var/tmp/sosreport-microshift-2-2023-03-23-ylwbkjc.tar.xz.sha256
```

Upload the archives to Red Hat Technical Support as described in [this section](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/getting_the_most_from_your_support_experience/generating-an-sos-report-for-technical-support_getting-the-most-from-your-support-experience#methods-for-providing-an-sos-report-to-red-hat-technical-support_generating-an-sosreport-for-technical-support)

The `sos` archives may consume significant disk space. Make sure to delete the report files after uploading them.

```bash
sudo rm -f /var/tmp/sosreport-*
```

## Pod Security Admission and Security Context Constraints

MicroShift limits the SecurityContextConstraint of new namespaces to
`restricted-v2` by default. This can mean that workloads that run on
OpenShift result in pod security admission errors when run on
MicroShift. Refer to the [pod security HowTo](howto_pod_security.md)
for a detailed example of configuring a workload to run with a custom
security context.

## Configure Log Verbosity for OVN-Kubernetes

The default log verbosity level in MicroShift is set to `4` for OVN-Kubernetes
and `info` for OVN. You have the option to configure the log verbosity for
debugging and troubleshooting purposes. To do so, you can create a ConfigMap
named `env-overrides` with specific keys in the `openshift-ovn-kubernetes`
namespace and restart the corresponding OVN-Kubernetes pods.

Here's an example:

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: env-overrides
  namespace: openshift-ovn-kubernetes
  annotations:
data:
  # To set the log levels for the ovnkube-node pod
  # replace this with the node's name (from oc get nodes)
  ip-10-0-135-96.us-east-2.compute.internal: |
    # To enable debug logging for ovn-controller:
    # Logging verbosity level: off, emer, err, warn, info, or  dbg (default: info)
    OVN_LOG_LEVEL=dbg
  # To adjust the log levels for the ovnkube-master pod, use _master
  _master: |
    # This sets the log level for the ovn-kubernetes process
    # Logging verbosity level: 5=debug, 4=info, 3=warn, 2=error, 1=fatal (default: 4).
    OVN_KUBE_LOG_LEVEL=5
    # To enable debug logging for OVN northd, nbdb and sbdb processes:
    # Logging verbosity level: off, emer, err, warn, info, or  dbg (default: info)
    OVN_LOG_LEVEL=dbg
```
