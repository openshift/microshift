# Managing MicroShift Certificates

The `microshift certs` commands inspect and renew locally managed certificates.
All commands require root, including status and renewal dry-runs.

## Inspecting Certificate Status

Run `microshift certs status` as root to inspect the certificates managed by
MicroShift, including CAs, serving certificates, client certificates, and peer
certificates:

```bash
sudo microshift certs status
```

The command reads the local configuration and existing certificates under
`/var/lib/microshift/certs`. During normal operation it does not change certificate
material. It never restarts services or requires the Kubernetes API to be available.
An interrupted renewal is recovered before reporting status; see [Recovery](#recovery).
MicroShift must
have initialized its certificates first. This is not an inventory of certificates
managed by workloads or supplied externally by users.

## Output Formats

Without an output flag, the command prints a table with `SERVICE`, `CERTIFICATE`,
`STATUS`, `EXPIRY`, and `MESSAGE` columns. Expiry times are UTC. Status is
`Healthy`, `ExpiresSoon`, `ExpirationImminent`, `Expired`, or `NotYetValid`;
the message distinguishes certificates that are not expiring, expiring, expired,
or not yet valid.

Healthy certificates use a concise message, for example `Valid for 300 days`.
`ExpiresSoon` reports that remaining validity is at or below the warning
threshold; `ExpirationImminent`
reports the critical threshold instead. Thresholds use the certificate's actual
lifetime and rotation policy, not a fixed number of days. Displayed day counts
are rounded up; status comparisons use the unrounded values. Expired and
not-yet-valid messages do not include thresholds.

For automation, select JSON or YAML with `-o` or `--output`:

```bash
sudo microshift certs status -o json
sudo microshift certs status --output=yaml
```

Both formats emit one `CertificateStatusList` document on stdout, with
`apiVersion: microshift.openshift.io/v1alpha1`. They contain the same fields:

- `generatedAt`: the timestamp used for the report's calculations.
- `config`: the certificate policy resolved from the current configuration files,
  including `forceRestartOnExpirationImminent` and serving/CA validity durations.
  This is not a query of the running service's startup configuration.
- `items`: certificates sorted by service and then name. Each item contains
  `service`, `name`, `role`, `rotationPolicy`, `status`, `notBefore`, `notAfter`,
  and `remainingSeconds`. The `status` field uses the same five state names as the
  table. Remaining seconds are zero at expiry and negative afterward.
- `warnings`: configuration warnings, or an empty array when none apply.

For currently valid certificates, the `standard` rotation policy is `Healthy`
above 58.3% remaining validity, `ExpiresSoon` above 33.3%, and
`ExpirationImminent` otherwise. The `extended` policy uses 15% and 10% thresholds.
At or after `notAfter`, certificates are `Expired`. Before `notBefore`, they are
`NotYetValid`, with a `Valid in ...` message in the table. `NotYetValid` is an
unhealthy state: check clock synchronization and certificate issuance before
deciding whether renewal is needed. It does not mean expiration is imminent.
A successful report exits with code 0 regardless of certificate state;
automation should inspect the reported `status` values.

The report contains certificate metadata, not certificate PEM data or private
keys. Go consumers can use the exported types and `AddToScheme` in
`github.com/openshift/microshift/pkg/apis/certificates/v1alpha1` to decode the
versioned documents.

## Renewing Certificates

Select exactly one renewal mode:

- `--serving` renews all managed leaf certificates: serving, client, and peer.
  CA certificates and keys remain unchanged.
- `--ca` renews every managed CA and cascades to all of its descendants.

Preview either operation while MicroShift is running:

```bash
sudo microshift certs renew --serving --dry-run
sudo microshift certs renew --ca --dry-run -o yaml
```

Dry-run validates the existing certificate/key pairs and signing relationships,
then reports current and proposed expiry times without generating keys or staging
new material. Renewal uses the configured serving and CA validity periods and
caps each descendant's expiry at the earliest expiry in its signing chain. If a
CA is expired or not yet valid, leaf-only renewal is refused; use `--ca` to renew
the chain. With `forceRestartOnExpirationImminent` enabled, leaf-only renewal
also refuses CAs that have reached their policy-derived rotation deadline.

### Optional Pre-Renewal Backup

Renewal retains recoverable originals during activation, but removes them after
successful completion. This is not a permanent backup of MicroShift data.

If you need a full pre-renewal backup, schedule a separate maintenance window
**before preparing renewal**. Unlike renewal preparation, `microshift backup`
requires both `microshift.service` and `microshift-etcd.scope` to be stopped.
Choose a backup destination that does not already exist:

```bash
sudo systemctl stop microshift
sudo microshift backup /var/lib/microshift-backups/before-certificate-renewal
```

The backup command checks that both units are stopped. If etcd is still stopping,
wait for it to stop before retrying. After the backup succeeds, start MicroShift
again:

```bash
sudo systemctl start microshift
```

You can then prepare renewal online and choose a later activation window. A
backup is optional, it is not a prerequisite for the online workflow below.

### Prepare and Activate Renewal

Prepare renewal while MicroShift is running, then activate it by restarting
MicroShift during your maintenance window:

```bash
sudo microshift certs renew --serving
sudo microshift certs status -o yaml
# At the chosen maintenance window:
sudo systemctl restart microshift
```

Replace `--serving` with `--ca` to renew the whole chain. Renewal saves a validated
pending generation without changing active certificates, bundles, or kubeconfigs.
It never stops or starts a service. If MicroShift is already stopped, preparation
has the same behavior: start it to activate the pending generation. Successful
table output explicitly describes the pending certificates, not those in use.

Status continues to report active files in `items`. When renewal is pending,
JSON/YAML also includes `pendingRenewal`, a `CertificateRenewalResult` containing
the prepared expiry dates, and all formats warn that a restart is required.
Preparation does not extend active certificate lifetimes or reset the running
service's automatic rotation deadline. Restart before the active certificates
expire; the pending certificates' validity also starts at generation, not activation.

Generated kubeconfigs under `/var/lib/microshift/resources/kubeadmin` are updated in the
same transaction at activation. After **CA activation**, redistribute those kubeconfigs to any
external locations where you previously copied them. Existing external
kubeconfigs remain trusted after leaf-only renewal until their own certificates
expire. Applications that cache certificates or CA bundles may need a reload or
restart after either operation.

At startup, MicroShift updates the pod templates of its directly deployed
API-consuming components when their active certificate generation changes.
OVN is reconciled first so that it remounts the renewed admin kubeconfig. After
CA renewal, DNS, ingress, service-ca, the LVMS operator, the CSI snapshot
controller, and Multus (when enabled) also receive updated pod templates.
MicroShift waits for the current service-account CA bundle before requesting
these replacements. Kubernetes recreates the pods using each controller's
existing update strategy; this can temporarily interrupt networking and ingress
during the maintenance window.

On an existing installation, startup waits up to two minutes for OVN's updated
pods. If they are still unavailable, MicroShift logs a warning and continues
deploying the other components while Kubernetes keeps reconciling the rollout.

Preparing renewal does not change pod templates. An ordinary restart with
unchanged certificates does not request another certificate rollout, and an
interrupted rollout is reconciled again on the next startup. MicroShift does not
restart customer workloads or workloads deployed by other operators; their
owners remain responsible for reloading cached certificates and CA bundles.

Both modes accept `-o json` and `-o yaml`. A successful command emits one versioned
`CertificateRenewalResult` with `items` sorted lexicographically by service, then
certificate name, with inventory path as a tie-breaker, and an `impact` summary.
This is presentation order, not the order in which certificates are regenerated.
Dry-run reports `status: validated`, `dryRun: true`, and `changed: false`
for every item. Prepared renewal reports `status: pending`, `dryRun: false`, and
`changed: true`, with expiry dates read back from the pending certificates.
`changed` refers to the prepared material, not active files. Running renewal again
prepares a fresh replacement for the pending generation, using the active PKI as
its base. A generation or validation failure preserves the previous pending set.
Dry-run likewise plans against active files and does not change a pending renewal.

### Example Output

These excerpts use illustrative timestamps and show one certificate. Actual
reports include all selected certificates; additional table rows and table-mode
warnings on stderr are omitted here.

`sudo microshift certs renew --serving --dry-run` previews the expiry change:

```text
DRY RUN: Validated certificate renewal; no certificates or kubeconfigs were changed.
SERVICE         ROLE     PARENT CA                        CERTIFICATE                       CURRENT EXPIRY        PROPOSED EXPIRY
kube-apiserver  serving  kube-apiserver-localhost-signer  kube-apiserver-localhost-serving  2027-01-01T00:00:00Z  2027-10-23T00:00:00Z
...
```

After `sudo microshift certs renew --serving`, the table describes the prepared
certificates, not those currently active. Excerpt:

```text
SERVICE         CERTIFICATE                       STATUS   EXPIRY                MESSAGE
kube-apiserver  kube-apiserver-localhost-serving  Healthy  2027-10-23T00:00:00Z  Valid for 366 days
...
CA certificates are unchanged by leaf renewal.
Active certificates and kubeconfigs are unchanged. Start or restart microshift.service to activate a pending renewal.
```

Before restarting, `sudo microshift certs status -o yaml` still shows the old
expiry in `items`. The new expiry appears only in `pendingRenewal`, alongside
`status: pending`, `changed: true`, and the restart impact. The same fields are
available with `-o json`:

```yaml
apiVersion: microshift.openshift.io/v1alpha1
config:
  caValidity: 87600h
  forceRestartOnExpirationImminent: true
  servingValidity: 8760h
generatedAt: "2026-10-22T12:00:00Z"
items:
- name: kube-apiserver-localhost-serving
  notAfter: "2027-01-01T00:00:00Z"
  notBefore: "2026-01-01T00:00:00Z"
  remainingSeconds: 6091200
  role: serving
  rotationPolicy: standard
  service: kube-apiserver
  status: ExpirationImminent
kind: CertificateStatusList
pendingRenewal:
  apiVersion: microshift.openshift.io/v1alpha1
  dryRun: false
  generatedAt: "2026-10-22T12:00:00Z"
  impact:
    applicationReloadMayBeRequired: true
    kubeconfigRedistributionRequired: false
    serviceRestartRequired: true
  items:
  - changed: true
    currentNotAfter: "2027-01-01T00:00:00Z"
    name: kube-apiserver-localhost-serving
    newNotAfter: "2027-10-23T00:00:00Z"
    parentCA: kube-apiserver-localhost-signer
    role: serving
    service: kube-apiserver
  kind: CertificateRenewalResult
  mode: serving
  status: pending
  warnings:
  - Applications that cache certificates or CA bundles may need to be reloaded or
    restarted after activation.
  - Pending renewal does not extend the lifetime of active certificates or change
    the running service's automatic rotation deadline.
warnings:
- Renewed certificates are pending activation. Status items describe active files;
  restart MicroShift to activate the pending renewal.
```

### Recovery

Renewal stages certificates, keys, bundles, and generated resources in the
root-only transaction directory `/var/lib/microshift/.cert-renewal/`. Active
files remain under `/var/lib/microshift/certs/` and
`/var/lib/microshift/resources/` until activation. The private
`.cert-renewal/pending.json` file identifies the complete prepared generation so
startup can validate it before activation. Do not edit this file manually.

Allow enough free space on the data filesystem to stage `certs/` and `resources/`.
Renewal rejects symlinks and special files in these directories instead of
following them outside staging. It does not copy or replace the etcd database.
For the internal layout, metadata, and resource-refresh behavior, see
[Certificate Renewal Transactions](../contributor/certificates.md).

#### Startup Validation

Startup revalidates pending material after any data restore and before components
read certificates. Activation requires the previous etcd scope to be stopped.
If its state cannot be determined, activation fails without replacing active files.
Certificate-affecting configuration and active-PKI hashes must still match
preparation. If either has changed, or pending certificates have expired, startup
discards the pending generation, logs a warning, and continues with normal
certificate initialization using the current configuration and active material.
Run renewal again if still needed. Unrelated settings, such as logging, do not
invalidate a pending renewal.

When `forceRestartOnExpirationImminent` is enabled, startup also checks the pending
generation's policy-derived rotation deadline. If that deadline has already
arrived, it discards the generation and uses normal
certificate initialization instead of activating it and immediately restarting.
The warning notes that normal startup may rotate CAs, even if the discarded
renewal selected only `--serving`. Prepare and activate renewal before the
deadline; a pending generation does not postpone automatic rotation.
With the setting disabled, reaching this threshold alone does not block
activation; certificate validity and integrity checks still apply.

#### Corrupt Renewal Metadata

Malformed or incomplete `pending.json` does not block startup if MicroShift can
quarantine it safely. MicroShift moves the rejected transaction into the directory
`/var/lib/microshift/.cert-renewal-rejected/`, logs a warning, and continues normal
certificate initialization. This does not force a CA renewal; the existing
startup rotation policy still applies. Rejected certificates are never activated.
`microshift certs status` continues to report active certificates, with a warning
on stderr for table output or in `warnings` for JSON/YAML. Before startup has
quarantined corrupt metadata, status also reports the problem without changing
any files and directs the administrator to restart and prepare renewal again.
The rejection warning and quarantined files survive restarts and new renewal
preparation. They are removed after a replacement renewal is successfully
activated. Quarantined files include private keys and remain root-only.

Filesystem errors, unsupported pending-record versions, invalid certificate
material, and failed interrupted-activation recovery still block startup.

#### Pending Certificates Are Not Yet Valid

Startup refuses to activate a pending certificate when the host's current time
is earlier than its `notBefore` timestamp. For this specific failure, check host
clock synchronization and certificate issuance timestamps; a backward clock
adjustment after preparation can make a previously valid certificate appear
not yet valid. The pending generation is retained. This differs from expired
pending certificates, which startup discards as described above.

#### Interrupted Activation

Failures before active certificates and resources are replaced leave them
untouched. Failed replacement or validation triggers restoration of the originals.
If the process or host is interrupted during activation or rollback, the next
startup restores the originals before using the data. The next `certs status` or
`certs renew` can also recover an interrupted activation while MicroShift is
stopped. This recovery also applies to a subsequent dry-run:

```bash
sudo systemctl stop microshift
sudo microshift certs status
```

Recovery of an interrupted activation discards its pending generation; use status
to inspect the restored material and prepare renewal again if needed.
Startup refuses to use an unrecoverable transaction. Do not delete the transaction
directory manually: it may contain the only recoverable copies. If recovery
fails, keep MicroShift stopped and preserve that directory for troubleshooting.
Status and dry-run share an operation lock at
`/var/lib/microshift-backups/certs.lock`; startup writes, preparation and recovery
exclude concurrent certificate operations. The running service instead holds
`/var/lib/microshift-backups/certs-runtime.lock`, which prevents live recovery
without blocking preparation. Both files stay outside the data directory so
backup restoration cannot replace them. Do not delete them: the files persist,
but locks are released when the owning processes close them or exit.
Startup waits up to 30 seconds for an existing certificate operation to release
the operation lock. If it remains busy, startup fails with a timeout error.
Certificate CLI commands continue to fail immediately when their lock is busy.

## Configuring Certificate Lifetimes and Restart Policy

Set these fields in `/etc/microshift/config.yaml` or a YAML drop-in under
`/etc/microshift/config.d/`:

```yaml
certificates:
  servingValidity: 1008h  # Six weeks; default is 8760h (365 days).
  caValidity: 87600h      # Default is 3650 days.
  forceRestartOnExpirationImminent: false  # Default is true.
```

Durations use Go syntax (`h`, `m`, `s`), not days or years. Both must be positive,
and `caValidity` must be strictly greater than `servingValidity`. Validate the
configuration with `sudo microshift show-config` before preparing renewal.

The validity settings apply to newly issued managed serving certificates and
CAs, including intermediate CAs. Client and peer certificates retain their
existing issuance durations and explicit rotation policies. Every newly issued
descendant is bounded by its signing chain. User-provided named certificates
are not managed by these settings. As before, expiry is aligned to midnight on
the next day plus the configured duration, so the issued lifetime can be up to
one day longer than configured (or shorter when capped by an existing CA).

Changing configuration does not rewrite existing certificates, activate pending
renewal, or restart MicroShift. Each `certs status` or `certs renew` invocation
reads the current configuration, so renewal can be prepared with new validity
settings while MicroShift is running. Activation still requires a later service
start. A validity change after preparation invalidates that pending generation;
prepare renewal again before restarting.

MicroShift does **not** hot-reload its main configuration or drop-ins. The running
service's `forceRestartOnExpirationImminent` policy changes only on its next
start. Setting it to `false` makes the critical certificate threshold warn-only;
it does not extend certificate validity, disable other restart causes, or disable
certificate regeneration on startup. Existing DNS/hosts file watchers are
unaffected. The status table labels the restart policy as configured, not as a
live service setting.

Startup renewal and the running service's critical deadline use percentages of
the validity encoded in each certificate, not fixed-month windows or the new
configured lifetime. Short-lived custom certificates therefore do not immediately
trigger renewal. Existing certificates keep their encoded dates until explicit
renewal or normal startup regeneration is required. All CAs use the extended
policy, including legacy CAs originally issued for one year.

After an upgrade, a legacy one-year CA becomes eligible for startup regeneration
at 15% remaining validity (about 55 days), while its running-service critical
deadline is at 10% remaining (about 37 days). A standard serving certificate
becomes critical at 33.3% remaining validity. If a renewed serving certificate
is capped by a retained CA's expiry, it can trigger a restart before the CA is
eligible for regeneration, then be reissued again when the CA rotates. Additional
certificate-driven restarts are therefore possible during this transition when
automatic restart is enabled.

An upgrade alone does not force CA replacement. To renew the chain together,
prepare `microshift certs renew --ca` and restart during a maintenance window,
following the CA trust and kubeconfig redistribution steps above.

## Warnings and Errors

In table mode, configuration warnings are written to stderr with a `WARNING:`
prefix. In JSON/YAML mode, warnings are included in the document's `warnings`
array, leaving stderr empty on success.

Failures exit non-zero. In table mode, stderr contains a human-readable error.
With `-o json` or `-o yaml`, failures leave stdout empty and write exactly one
versioned `Error` document to stderr, without usage text or additional diagnostics.
For example, invalid configuration produces this JSON shape:

```json
{
  "apiVersion": "microshift.openshift.io/v1alpha1",
  "kind": "Error",
  "generatedAt": "2026-09-25T12:00:00Z",
  "code": "InvalidConfiguration",
  "message": "failed to load MicroShift configuration; check /etc/microshift/config.yaml and /etc/microshift/config.d",
  "details": null
}
```

Use `code`, rather than parsing `message`, to classify failures. Certificate error
codes include `InvalidArguments`, `InsufficientPrivileges`,
`InvalidConfiguration`, `CertificateInventoryFailed`, `MicroShiftRunning`,
`RenewalFailed`, `RecoveryFailed`, and `InternalError`.
Configuration errors deliberately omit the underlying loader diagnostic because
it can contain raw configuration or credentials. Inspect the configuration files
locally without copying sensitive values into logs.

Only `json` and `yaml` are accepted values for the output flag; omit the flag for
the table. An unsupported output format is rejected with a human-readable error.
