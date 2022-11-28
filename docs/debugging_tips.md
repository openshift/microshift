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

> See [Generating sos reports for technical support](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html-single/generating_sos_reports_for_technical_support/index) for more information on the `sos` utility usage.

Log into the host running MicroShift and execute the following command to generate an obfuscated
report that should not contain sensitive information.

```bash
sudo sos report --batch --clean
```

The report archives can be found in the `/var/tmp/sosreport-*` files.

```bash
$ sudo ls -tr /var/tmp/sosreport-* | tail -2
/var/tmp/sosreport-host0-2022-11-24-pvbcaji-obfuscated.tar.xz
/var/tmp/sosreport-host0-2022-11-24-pvbcaji-obfuscated.tar.xz.sha256
```

Upload the archives to Red Hat Technical Support as described in [this section](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html-single/generating_sos_reports_for_technical_support/index#methods-for-providing-an-sos-report-to-red-hat-technical-support_generating-an-sosreport-for-technical-support)

The `sos` archives may consume significant disk space. Make sure to delete the report files after uploading them.

```bash
sudo rm -f /var/tmp/sosreport-*
```
