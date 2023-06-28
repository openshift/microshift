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

Log into the host running MicroShift and execute the following command to generate an obfuscated
report that should not contain sensitive information (in this example only microshift plugins are
enabled. Your system's output may vary):
```bash
$ sos report --batch --clean --all-logs --profile microshift

sosreport (version 4.3)

...

 Setting up archive ...
 Setting up plugins ...
 Running plugins. Please wait ...

  Starting 1/1   microshift      [Running: microshift]

  Finished running plugins

Successfully obfuscated 1 report(s)

Creating compressed archive...

A mapping of obfuscated elements is available at
	/var/tmp/sosreport-microshift-2-2023-03-23-ylwbkjc-private_map

Your sosreport has been generated and saved in:
	/var/tmp/sosreport-microshift-2-2023-03-23-ylwbkjc-obfuscated.tar.xz

...
```
> Sos must always run with root privileges.

The output file and its checksum are generated in `/var/tmp/` directory.
```bash
$ sudo ls -tr /var/tmp/sosreport-* | tail -2
/var/tmp/sosreport-microshift-2-2023-03-23-ylwbkjc-obfuscated.tar.xz
/var/tmp/sosreport-microshift-2-2023-03-23-ylwbkjc-obfuscated.tar.xz.sha256
```

Upload the archives to Red Hat Technical Support as described in [this section](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html-single/generating_sos_reports_for_technical_support/index#methods-for-providing-an-sos-report-to-red-hat-technical-support_generating-an-sosreport-for-technical-support)

The `sos` archives may consume significant disk space. Make sure to delete the report files after uploading them.

```bash
sudo rm -f /var/tmp/sosreport-*
```
