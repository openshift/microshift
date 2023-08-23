# MicroShift SOS Report to Grafana Loki

Set up a Grafana+Loki+Promtail stack for a MicroShift SOS Report.

## Usage

```sh
./sos-to-loki.sh SOS-REPORT
```

`SOS-REPORT` can be:
- local directory (extracted SOS report)
- local SOS report archive file
- remote SOS report archive file

Example:
```sh
$ ./sos-to-loki.sh https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_microshift/2225/pull-ci-openshift-microshift-main-microshift-metal-tests/1692522601659764736/artifacts/microshift-metal-tests/openshift-microshift-e2e-metal-tests/artifacts/scenario-info/el92-src@fdo/vms/host1/sos/sosreport-el92-src-fdo-host1-2023-08-18-vaafzhm.tar.xz
Fetching SOS report
Path to local SOS report: /tmp/sos-to-loki/sosreport-el92-src-fdo-host1-2023-08-18-vaafzhm
Starting Grafana+Loki+Promtail stack

                 Timezone     First journal entry       Last entry
SOS              EDT          Aug 18 10:25:02           Aug 18 10:43:36
Local            CEST         Aug 18 16:25:02           Aug 18 16:43:36
URL timestamps   CEST         1692368402000             1692370116000

Link to query with journal:
http://localhost:3000/explore?orgId=1&left=%7B%22datasource%22:%22P8E80F9AEF21F6940%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22expr%22:%22%7Bfilename%3D%5C%22%2Flogs%2Fsos_commands%2Flogs%2Fjournalctl_--no-pager%5C%22%7D%20%7C%3D%20%60%60%22,%22queryType%22:%22range%22,%22datasource%22:%7B%22type%22:%22loki%22,%22uid%22:%22P8E80F9AEF21F6940%22%7D,%22editorMode%22:%22builder%22%7D%5D,%22range%22:%7B%22from%22:%221692368402000%22,%22to%22:%221692370116000%22%7D%7D
```

Query link points to `sos_commands/logs/journalctl_--no-pager` file with time range from first and last journal entries (with a 5 minute margin both ways).
