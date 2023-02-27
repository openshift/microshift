# MicroShift RPM Packages for Development and Testing

MicroShift Release Candidates (RC) and Engineering Candidates (EC) are publicly available for following architectures::
- [x86_64](https://mirror.openshift.com/pub/openshift-v4/x86_64/microshift/ocp-dev-preview/)
- [aarch64](https://mirror.openshift.com/pub/openshift-v4/aarch64/microshift/ocp-dev-preview/)

MicroShift RPM development packages are intended to be used internally at Red Hat to facilitate the development and QE processes, **never** to be distributed to customers. The packages are produced every time a corresponding OpenShift nightly build is successful.

> For more information, see the `MicroShift Pre-Release RPM Packages` internal Red Hat Confluence page.

## Building binaries from source RPM

Following steps are applicable for source RPM from both release build system and developer's environment.

> Hint: To build source RPM in developer environment run: `make srpm` - artifact will be in `_output/rpmbuild/SRPMS/`

Assuming there's only one srpm in current directory, following commands will extract and enter the source directory:
```bash
mkdir -p ~/microshift-srpm-builds/
cd ~/microshift-srpm-builds/
cp ~/microshift/_output/rpmbuild/SRPMS/microshift*.src.rpm .
rpm2cpio microshift-*.src.rpm  | cpio -idmv
tar xf microshift-*.tar.gz
cd microshift-*/
```

> Note: Source RPM does not include `.git/` directory.
>
> This means that commands (such as `make rpm`) expecting git repository will fail and 
> version strings embedded in `microshift` binary are not fully populated.


To build `microshift` and `microshift-etcd` binaries run:
```bash
make GO_MOD_FLAGS='-buildvcs=false -mod=vendor'
```
