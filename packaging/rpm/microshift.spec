# parameters that must be provided via --define , or fixed into the spec file:
# global version 4.7.0
# global release 2021_08_31_224727
# global commit 81264d0ebb17fef06eff9ec7d4f2a81631c6b34a

%{!?commit:
# DO NOT MODIFY: the value on the line below is sed-like replaced by openshift/doozer
%global commit 0000000000000000000000000000000000000000
}

# golang specifics
# Needs to match go.mod go directive
%global golang_version 1.20
#debuginfo not supported with Go
%global debug_package %{nil}
# modifying the Go binaries breaks the DWARF debugging
%global __os_install_post %{_rpmconfigdir}/brp-compress

# SELinux specifics
%global selinuxtype targeted
%define selinux_policyver 3.14.3-67
%define microshift_relabel_files() \
   mkdir -p /var/run/kubelet; \
   mkdir -p /var/lib/kubelet/pods; \
   mkdir -p /etc/microshift; \
   mkdir -p /usr/lib/microshift; \
   mkdir -p /var/run/secrets/kubernetes.io/serviceaccount; \
   mkdir -p /var/lib/microshift-backups; # Creating folder to avoid GreenBoot race condition so that correct label is applied \
   restorecon -R /var/run/kubelet; \
   restorecon -R /var/lib/kubelet/pods; \
   restorecon -R /var/run/secrets/kubernetes.io/serviceaccount; \
   restorecon -R /var/lib/microshift-backups; \
   restorecon -R /etc/microshift; \
   restorecon -R /usr/lib/microshift
%define microshift_relabel_exes() \
   restorecon -v /usr/bin/microshift; \
   restorecon -v /usr/bin/microshift-etcd

# Git related details
%global shortcommit %(c=%{commit}; echo ${c:0:7})

Name: microshift
Version: %{version}
Release: %{release}%{dist}
Summary: MicroShift service
License: ASL 2.0
URL: https://github.com/openshift/microshift
Source0: https://github.com/openshift/microshift/archive/%{commit}/microshift-%{shortcommit}.tar.gz

ExclusiveArch: x86_64 aarch64

BuildRequires: gcc
BuildRequires: make
BuildRequires: policycoreutils
BuildRequires: systemd

# DO NOT REMOVE
#
# Without the dependency on the compiler via this file, the production
# build pipeline does not work. Sometimes that pipeline requires
# compiling with a version of go that is not released into public RHEL
# repos, yet, so we do not specify a version here. Instead, the
# version is checked in the prep step below.
#
BuildRequires: golang
# DO NOT REMOVE

Requires: cri-o >= 1.25
Requires: cri-tools >= 1.25
Requires: iptables
Requires: microshift-selinux = %{version}
Requires: microshift-networking = %{version}
Requires: microshift-greenboot = %{version}
Requires: conntrack-tools
Requires: sos
Requires: crun
Requires: openshift-clients

%{?systemd_requires}

%description
The microshift package provides an OpenShift Kubernetes distribution optimized for small form factor and edge computing.


%package release-info
Summary: Release information for MicroShift
BuildArch: noarch
BuildRequires: jq
BuildRequires: gettext

%description release-info
The microshift-release package provides release information files for this
release. These files contain the list of container image references used by
MicroShift and can be used to embed those images into osbuilder blueprints.
An example of such osbuilder blueprints for x86_64 and aarch64 platforms are
also included in the package.


%package selinux
Summary: SELinux policies for MicroShift
BuildRequires: selinux-policy >= %{selinux_policyver}
BuildRequires: selinux-policy-devel >= %{selinux_policyver}
Requires: container-selinux
BuildArch: noarch
Requires: microshift = %{version}
Requires: selinux-policy >= %{selinux_policyver}

%description selinux
The microshift-selinux package provides the SELinux policy modules required by MicroShift.


%package networking
Summary: Networking components for MicroShift
Requires: microshift = %{version}
Requires: (openvswitch3.1 or openvswitch >= 3.1)
Requires: NetworkManager
Requires: NetworkManager-ovs
Requires: jq

%description networking
The microshift-networking package provides the networking components necessary for the MicroShift default CNI driver.

%package greenboot
Summary: Greenboot components for MicroShift
BuildArch: noarch
Requires: microshift = %{version}
Requires: greenboot

%description greenboot
The microshift-greenboot package provides the Greenboot scripts used for verifying that MicroShift is up and running.

%package olm
Summary: Operator Lifecycle Manager components for MicroShift
ExclusiveArch: x86_64 aarch64
Requires: microshift = %{version}

%description olm
The microshift-olm package provides the required manifests for the Operator Lifecycle Manager to be installed on MicroShift.

%package olm-release-info
Summary: Release information for Operator Lifecycle Manager components for MicroShift
BuildArch: noarch
Requires: microshift-release-info = %{version}

%description olm-release-info
The microshift-olm-release-info package provides release information files for this
release. These files contain the list of container image references used by
the Operator Lifecycle Manager for MicroShift and can be used to embed those
images into osbuilder blueprints.

%package multus
Summary: Multus CNI for MicroShift
ExclusiveArch: x86_64 aarch64
Requires: microshift = %{version}

%description multus
The microshift-multus package provides the required manifests for the Multus CNI to be installed on MicroShift.

%package multus-release-info
Summary: Release information for Multus CNI for MicroShift
BuildArch: noarch
Requires: microshift-release-info = %{version}

%description multus-release-info
The microshift-multus-release-info package provides release information files for this
release. These files contain the list of container image references used by
the Multus CNI for MicroShift and can be used to embed those images into osbuilder blueprints.

%prep
# Dynamic detection of the available golang version also works for non-RPM golang packages
golang_detected=$(go version | awk '{print $3}' | tr -d '[a-z]' | cut -f1-2 -d.)
golang_required=%{golang_version}
if [[ "${golang_detected}" < "${golang_required}" ]] ; then
  echo "The detected go version ${golang_detected} is less than the required version ${golang_required}" > /dev/stderr
  exit 1
fi

%setup -n microshift-%{commit}

%build

GOOS=linux

%ifarch %{arm} aarch64
GOARCH=arm64
%endif

%ifarch x86_64
GOARCH=amd64
%endif

# if we have git commit/tag/state to be embedded in the binary pass it down to the makefile
%if %{defined embedded_git_commit}
make _build_local GOOS=${GOOS} GOARCH=${GOARCH} EMBEDDED_GIT_COMMIT=%{commit} EMBEDDED_GIT_TAG=%{embedded_git_tag} EMBEDDED_GIT_TREE_STATE=%{embedded_git_tree_state} MICROSHIFT_VERSION=%{version}
%else
make _build_local GOOS=${GOOS} GOARCH=${GOARCH} MICROSHIFT_VERSION=%{version} EMBEDDED_GIT_COMMIT=%{commit}
%endif

cp ./_output/bin/${GOOS}_${GOARCH}/microshift ./_output/microshift
cp ./_output/bin/${GOOS}_${GOARCH}/microshift-etcd ./_output/microshift-etcd

# SELinux modules build

make --directory packaging/selinux

# osbuilder sample blueprints build
function create_blueprint() {
  local -r larch="$1"

  REPLACE_USHIFT_VERSION="%{version}" \
    REPLACE_USHIFT_ARCH="${larch}" \
    envsubst < "packaging/blueprint/blueprint.toml.template" > "packaging/blueprint/blueprint-${larch}.toml"

  jq -r \
    '.images | .[] | ("[[containers]]\nsource = \"" + . + "\"\n")' \
    "assets/release/release-${larch}.json" \
    >> "packaging/blueprint/blueprint-${larch}.toml"
}

create_blueprint x86_64
create_blueprint aarch64

%install

install -d %{buildroot}%{_bindir}
install -p -m755 ./_output/microshift %{buildroot}%{_bindir}/microshift
install -p -m755 ./_output/microshift-etcd %{buildroot}%{_bindir}/microshift-etcd
install -p -m755 scripts/microshift-cleanup-data.sh %{buildroot}%{_bindir}/microshift-cleanup-data
install -p -m755 scripts/microshift-sos-report.sh %{buildroot}%{_bindir}/microshift-sos-report

install -d -m755 %{buildroot}%{_sharedstatedir}/microshift
install -d -m755 %{buildroot}%{_sharedstatedir}/microshift-backups

install -d -m755 %{buildroot}%{_sysconfdir}/crio/crio.conf.d

%ifarch %{arm} aarch64
install -p -m644 packaging/crio.conf.d/10-microshift_arm64.conf %{buildroot}%{_sysconfdir}/crio/crio.conf.d/10-microshift.conf
%endif

%ifarch x86_64
install -p -m644 packaging/crio.conf.d/10-microshift_amd64.conf %{buildroot}%{_sysconfdir}/crio/crio.conf.d/10-microshift.conf
%endif

install -p -m644 packaging/crio.conf.d/11-microshift-ovn.conf %{buildroot}%{_sysconfdir}/crio/crio.conf.d/11-microshift-ovn.conf

install -d -m755 %{buildroot}/%{_unitdir}
install -p -m644 packaging/systemd/microshift.service %{buildroot}%{_unitdir}/microshift.service

install -d -m755 %{buildroot}/%{_sysconfdir}/microshift
install -d -m755 %{buildroot}/%{_sysconfdir}/microshift/manifests
install -d -m755 %{buildroot}/%{_sysconfdir}/microshift/manifests.d
install -p -m644 packaging/microshift/config.yaml %{buildroot}%{_sysconfdir}/microshift/config.yaml.default
install -p -m644 packaging/microshift/lvmd.yaml %{buildroot}%{_sysconfdir}/microshift/lvmd.yaml.default
install -p -m644 packaging/microshift/ovn.yaml %{buildroot}%{_sysconfdir}/microshift/ovn.yaml.default

# /usr/lib/microshift manifest directories for other packages to add to
install -d -m755 %{buildroot}/%{_prefix}/lib/microshift
install -d -m755 %{buildroot}/%{_prefix}/lib/microshift/manifests
install -d -m755 %{buildroot}/%{_prefix}/lib/microshift/manifests.d

# release-info files
mkdir -p -m755 %{buildroot}%{_datadir}/microshift/release
install -p -m644 assets/release/release-{x86_64,aarch64}.json %{buildroot}%{_datadir}/microshift/release
mkdir -p -m755 %{buildroot}%{_datadir}/microshift/blueprint
install -p -m644 packaging/blueprint/blueprint*.toml %{buildroot}%{_datadir}/microshift/blueprint

# spec validation files
mkdir -p -m755 %{buildroot}%{_datadir}/microshift/spec
install -p -m644 cmd/generate-config/config/config-openapi-spec.json %{buildroot}%{_datadir}/microshift/spec/config-openapi-spec.json

# Memory tweaks to the OpenvSwitch services
mkdir -p -m755 %{buildroot}%{_sysconfdir}/systemd/system/ovs-vswitchd.service.d
mkdir -p -m755 %{buildroot}%{_sysconfdir}/systemd/system/ovsdb-server.service.d
install -p -m644 packaging/systemd/microshift-ovs-vswitchd.conf %{buildroot}%{_sysconfdir}/systemd/system/ovs-vswitchd.service.d/microshift-cpuaffinity.conf
install -p -m644 packaging/systemd/microshift-ovsdb-server.conf %{buildroot}%{_sysconfdir}/systemd/system/ovsdb-server.service.d/microshift-cpuaffinity.conf

# this script and systemd service configures openvswitch to properly operate with OVN
install -p -m644 packaging/systemd/microshift-ovs-init.service %{buildroot}%{_unitdir}/microshift-ovs-init.service
install -p -m755 packaging/systemd/configure-ovs.sh %{buildroot}%{_bindir}/configure-ovs.sh
install -p -m755 packaging/systemd/configure-ovs-microshift.sh %{buildroot}%{_bindir}/configure-ovs-microshift.sh

# Avoid firewalld manipulation and flushing of iptable rules,
# this is a workaround for https://issues.redhat.com/browse/NP-641
# It will trigger some warnings on the selinux audit log when restarting firewalld.
# In the future firewalld should stop flushing iptables unless we use any firewalld rule with direct
# iptables rules, once that's available in RHEL we can remove this workaround
# see https://github.com/firewalld/firewalld/issues/863#issuecomment-1407059938

mkdir -p -m755 %{buildroot}%{_sysconfdir}/systemd/system/firewalld.service.d
install -p -m644 packaging/systemd/firewalld-no-iptables.conf %{buildroot}%{_sysconfdir}/systemd/system/firewalld.service.d/firewalld-no-iptables.conf

mkdir -p -m755 %{buildroot}/var/run/kubelet
mkdir -p -m755 %{buildroot}/var/lib/kubelet/pods
mkdir -p -m755 %{buildroot}/var/run/secrets/kubernetes.io/serviceaccount

install -d %{buildroot}%{_datadir}/selinux/packages/%{selinuxtype}
install -m644 packaging/selinux/microshift.pp.bz2 %{buildroot}%{_datadir}/selinux/packages/%{selinuxtype}

# Greenboot scripts
install -d -m755 %{buildroot}%{_datadir}/microshift/functions
install -p -m644 packaging/greenboot/functions.sh %{buildroot}%{_datadir}/microshift/functions/greenboot.sh

install -d -m755 %{buildroot}%{_sysconfdir}/greenboot/check/required.d
install -p -m755 packaging/greenboot/microshift-running-check.sh %{buildroot}%{_sysconfdir}/greenboot/check/required.d/40_microshift_running_check.sh

install -d -m755 %{buildroot}%{_sysconfdir}/greenboot/red.d
install -p -m755 packaging/greenboot/microshift-pre-rollback.sh %{buildroot}%{_sysconfdir}/greenboot/red.d/40_microshift_pre_rollback.sh

# OLM manifests
install -d -m755 %{buildroot}/%{_prefix}/lib/microshift/manifests.d/001-microshift-olm
# Copy all the OLM manifests except the arch specific ones
install -p -m644 assets/optional/operator-lifecycle-manager/0000* %{buildroot}/%{_prefix}/lib/microshift/manifests.d/001-microshift-olm
install -p -m644 assets/optional/operator-lifecycle-manager/kustomization.yaml %{buildroot}/%{_prefix}/lib/microshift/manifests.d/001-microshift-olm
install -p -m755 packaging/greenboot/microshift-running-check-olm.sh %{buildroot}%{_sysconfdir}/greenboot/check/required.d/50_microshift_running_check_olm.sh

%ifarch %{arm} aarch64
cat assets/optional/operator-lifecycle-manager/kustomization.aarch64.yaml >> %{buildroot}/%{_prefix}/lib/microshift/manifests.d/001-microshift-olm/kustomization.yaml
%endif

%ifarch x86_64
cat assets/optional/operator-lifecycle-manager/kustomization.x86_64.yaml >> %{buildroot}/%{_prefix}/lib/microshift/manifests.d/001-microshift-olm/kustomization.yaml
%endif

# olm-release-info
mkdir -p -m755 %{buildroot}%{_datadir}/microshift/release
install -p -m644 assets/optional/operator-lifecycle-manager/release-olm-{x86_64,aarch64}.json %{buildroot}%{_datadir}/microshift/release/

# multus
install -d -m755 %{buildroot}/%{_prefix}/lib/microshift/manifests.d/000-microshift-multus
# Copy all the Multus manifests except the arch specific ones
install -p -m644 assets/optional/multus/0* %{buildroot}/%{_prefix}/lib/microshift/manifests.d/000-microshift-multus
install -p -m644 assets/optional/multus/kustomization.yaml %{buildroot}/%{_prefix}/lib/microshift/manifests.d/000-microshift-multus
install -p -m755 packaging/greenboot/microshift-running-check-multus.sh %{buildroot}%{_sysconfdir}/greenboot/check/required.d/41_microshift_running_check_multus.sh
install -p -m755 packaging/crio.conf.d/12-microshift-multus.conf %{buildroot}%{_sysconfdir}/crio/crio.conf.d/12-microshift-multus.conf

%ifarch %{arm} aarch64
cat assets/optional/multus/kustomization.aarch64.yaml >> %{buildroot}/%{_prefix}/lib/microshift/manifests.d/000-microshift-multus/kustomization.yaml
%endif

%ifarch x86_64
cat assets/optional/multus/kustomization.x86_64.yaml >> %{buildroot}/%{_prefix}/lib/microshift/manifests.d/000-microshift-multus/kustomization.yaml
%endif

# multus-release-info
mkdir -p -m755 %{buildroot}%{_datadir}/microshift/release
install -p -m644 assets/optional/multus/release-multus-{x86_64,aarch64}.json %{buildroot}%{_datadir}/microshift/release/

%pre networking

getent group hugetlbfs >/dev/null || groupadd -r hugetlbfs
usermod -a -G hugetlbfs openvswitch

%post

# This can be called only after microshift executable files are installed
%microshift_relabel_exes

%systemd_post microshift.service

# only for install, not on upgrades
if [ $1 -eq 1 ]; then
	# if crio was already started, restart it so it will catch /etc/crio/crio.conf.d/microshift.conf
	systemctl is-active --quiet crio && systemctl restart --quiet crio || true
fi

%pre selinux
%selinux_relabel_pre -s %{selinuxtype}

%post selinux

%selinux_modules_install -s %{selinuxtype} %{_datadir}/selinux/packages/%{selinuxtype}/microshift.pp.bz2
%microshift_relabel_files

%postun selinux

if [ $1 -eq 0 ]; then
    %selinux_modules_uninstall -s %{selinuxtype} microshift
fi

%posttrans selinux

%selinux_relabel_post -s %{selinuxtype}

%post networking
# setup ovs / ovsdb optimization to avoid full pre-allocation of memory
sed -i -n -e '/^OPTIONS=/!p' -e '$aOPTIONS="--no-mlockall"' /etc/sysconfig/openvswitch
%systemd_post microshift-ovs-init.service
systemctl is-active --quiet NetworkManager && systemctl restart --quiet NetworkManager || true
systemctl enable --now --quiet openvswitch || true

%preun networking
%systemd_preun microshift-ovs-init.service

%preun

%systemd_preun microshift.service


%files
%license LICENSE
%{_bindir}/microshift
%{_bindir}/microshift-etcd
%{_bindir}/microshift-cleanup-data
%{_bindir}/microshift-sos-report
%{_unitdir}/microshift.service
%{_sysconfdir}/crio/crio.conf.d/10-microshift.conf
%{_datadir}/microshift/spec/config-openapi-spec.json
%dir %{_sysconfdir}/microshift
%dir %{_sysconfdir}/microshift/manifests
%dir %{_sysconfdir}/microshift/manifests.d
%config(noreplace) %{_sysconfdir}/microshift/config.yaml.default
%config(noreplace) %{_sysconfdir}/microshift/lvmd.yaml.default
%config(noreplace) %{_sysconfdir}/microshift/ovn.yaml.default

%dir %{_datadir}/microshift
%dir %{_datadir}/microshift/spec
%dir %{_prefix}/lib/microshift
%dir %{_prefix}/lib/microshift/manifests
%dir %{_prefix}/lib/microshift/manifests.d

%files release-info
%dir %{_datadir}/microshift
%dir %{_datadir}/microshift/release
%dir %{_datadir}/microshift/blueprint

%{_datadir}/microshift/release/release-{x86_64,aarch64}.json
%{_datadir}/microshift/blueprint/blueprint*.toml

%files selinux
/var/run/kubelet
/var/lib/kubelet/pods
/var/run/secrets/kubernetes.io/serviceaccount
%{_datadir}/selinux/packages/%{selinuxtype}/microshift.pp.bz2
%ghost %{_sharedstatedir}/selinux/%{selinuxtype}/active/modules/200/microshift

%files networking
%{_sysconfdir}/crio/crio.conf.d/11-microshift-ovn.conf
%{_sysconfdir}/systemd/system/ovs-vswitchd.service.d/microshift-cpuaffinity.conf
%{_sysconfdir}/systemd/system/ovsdb-server.service.d/microshift-cpuaffinity.conf
%{_sysconfdir}/systemd/system/firewalld.service.d/firewalld-no-iptables.conf

# OpensvSwitch oneshot configuration script which handles ovn-k8s gateway mode setup
%{_unitdir}/microshift-ovs-init.service
%{_bindir}/configure-ovs.sh
%{_bindir}/configure-ovs-microshift.sh

%files greenboot
%{_sysconfdir}/greenboot/check/required.d/40_microshift_running_check.sh
%{_sysconfdir}/greenboot/red.d/40_microshift_pre_rollback.sh
%{_datadir}/microshift/functions/greenboot.sh
%dir %{_datadir}/microshift
%dir %{_datadir}/microshift/functions

%files olm
%dir %{_prefix}/lib/microshift/manifests.d/001-microshift-olm
%{_prefix}/lib/microshift/manifests.d/001-microshift-olm/*
%{_sysconfdir}/greenboot/check/required.d/50_microshift_running_check_olm.sh

%files olm-release-info
%{_datadir}/microshift/release/release-olm-{x86_64,aarch64}.json

%files multus
%dir %{_prefix}/lib/microshift/manifests.d/000-microshift-multus
%{_prefix}/lib/microshift/manifests.d/000-microshift-multus/*
%{_sysconfdir}/greenboot/check/required.d/41_microshift_running_check_multus.sh
%{_sysconfdir}/crio/crio.conf.d/12-microshift-multus.conf

%files multus-release-info
%{_datadir}/microshift/release/release-multus-{x86_64,aarch64}.json


# Use Git command to generate the log and replace the VERSION string
# LANG=C git log --date="format:%a %b %d %Y" --pretty="tformat:* %cd %an <%ae> VERSION%n- %s%n" packaging/rpm/microshift.spec
%changelog
* Mon Feb 26 2024 Patryk Matuszak <pmatusza@redhat.com> 4.16.0
- RPM packages for Multus CNI

* Thu Jan 25 2024 Patryk Matuszak <pmatusza@redhat.com> 4.16.0
- Rename CRI-O configs to include prefix

* Thu Jan 25 2024 Patryk Matuszak <pmatusza@redhat.com> 4.15.0
- Create microshift-olm-release-info RPM containing OLM release info files

* Thu Jan 25 2024 Gregory Giguashvili <ggiguash@redhat.com> 4.15.0
- OLM release info files are no longer included in the base release-info RPM package

* Wed Jan 24 2024 Patryk Matuszak <pmatusza@redhat.com> 4.15.0
- Add missing dependency of microshift-olm on microshift package

* Thu Dec 21 2023 Patryk Matuszak <pmatusza@redhat.com> 4.15.0
- Add OLM release info to microshift-olm RPMs

* Thu Dec 14 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.15.0
- Implement greenboot check for microshift-olm RPM

* Tue Dec 05 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.15.0
- The microshift-release-info RPM is no longer required
- The microshift-release-info RPM contains sample blueprints including container image references
- Fix package uninstall logic to clean up all the directories created on installation

* Mon Dec 04 2023 Patryk Matuszak <305846+pmtk@users.noreply.github.com> 4.15.0
- Change way of assembling microshift-olm RPM

* Tue Nov 28 2023 Joaquim Moreno Prusi <joaquim@redhat.com> 4.15.0
- Extend microshift.spec to build microshift-olm rpm

* Mon Nov 13 2023 Pablo Acevedo Montserrat <pacevedo@redhat.com> 4.15.0
- USHIFT-1872: Remove keyfile nm plugin force when installing

* Wed Nov 01 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.15.0
- Fix selinux labeling for microshift executable files

* Wed Sep 06 2023 Pablo Acevedo Montserrat <pacevedo@redhat.com> 4.14.0
- Add microshift-sos-report binary

* Thu Jul 27 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.14.0
- The microshift-greenboot package is no longer optional

* Tue Jul 25 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.14.0
- Add explicit version dependencies among MicroShift RPM packages

* Fri Jul 21 2023 Zenghui Shi <zshi@redhat.com> 4.14.0
- Add openvswitch user to hugetlbfs group

* Tue Jun 06 2023 Doug Hellmann <dhellmann@redhat.com> 4.14.0
- Restore golang BuildRequires setting

* Mon May 15 2023 Doug Hellmann <dhellmann@redhat.com> 4.14.0
- Remove version specifier for container-selinux to let the system
  make the best choice.

* Mon Apr 24 2023 Doug Hellmann <dhellmann@redhat.com> 4.14.0
- Add /etc/microshift/manifests.d and /usr/lib/microshift/manifests.d
  directories.

* Wed Apr 12 2023 Zenghui Shi <zshi@redhat.com> 4.13.0
- Upgrade openvswitch package version to 3.1

* Wed Mar 29 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.13.0
- Upgrade golang build-time dependency to 1.19 version

* Wed Mar 01 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.13.0
- Add lvmd.yaml and ovn.yaml default configuration files

* Fri Feb 24 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.13.0
- Implement MicroShift pre-rollback greenboot procedure

* Mon Feb 20 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.13.0
- Create empty manifests directory

* Tue Feb 07 2023 Gregory Giguashvili <ggiguash@redhat.com> 4.13.0
- Initial implementation of MicroShift integration with greenboot

* Mon Feb 06 2023 Ricardo Noriega de Soto <rnoriega@rehat.com> 4.13.0
- Require minimum CRIO version

* Fri Jan 27 2023 Miguel Angel Ajo Pelayo <majopela@rehat.com> 4.12.0
- Add firewalld systemd service override configuration to avoid access to iptables

* Tue Jan 24 2023 Patryk Matuszak <pmatusza@redhat.com> 4.13.0
- Include microshift-etcd in package

* Wed Dec 14 2022 Frank A. Zdarsky <fzdarsky@redhat.com> 4.12.0
- Add microshift-release-info subpackage

* Wed Dec 07 2022 Gregory Giguashvili <ggiguash@redhat.com> 4.12.0
- Update the summaries and descriptions of MicroShift RPM packages

* Tue Dec 06 2022 Patryk Matuszak <pmatusza@redhat.com> 4.12.0
- Add commit macro and embed it into binary

* Wed Nov 30 2022 Patryk Matuszak <pmatusza@redhat.com> 4.12.0
- Pass version macro to Makefile

* Wed Nov 30 2022 Gregory Giguashvili <ggiguash@redhat.com> 4.12.0
- Change the config.yaml file name to allow its overwrite by users

* Mon Nov 28 2022 Patryk Matuszak <pmatusza@redhat.com> 4.12.0
- Use commit time & sha for RPM and exec

* Fri Nov 25 2022 Gregory Giguashvili <ggiguash@redhat.com> 4.12.0
- Install sos utility with MicroShift and document its usage

* Mon Oct 24 2022 Zenghui Shi <zshi@redhat.com> 4.12.0
- Add arch specific crio conf

* Fri Sep 30 2022 Frank A. Zdarsky <fzdarsky@redhat.com> 4.12.0
- Update openswitch version to 2.17

* Wed Aug 31 2022 Doug Hellmann <dhellmann@redhat.com> 4.10.0
- Remove experimental comments from RPM description
- Add example config file to rpm

* Wed Aug 31 2022 Gregory Giguashvili <ggiguash@redhat.com> 4.10.0
- Fix RPM post install script not to return error when crio is not running
- Co-authored-by: Dan Clark <danielmclark@gmail.com>

* Wed Aug 31 2022 Patryk Matuszak <pmatusza@redhat.com> 4.10.0
- Removed hostpath-provisioner

* Tue Aug 02 2022 Zenghui Shi <zshi@redhat.com> 4.10.0
- Fix openvswitch issues when MicroShift service is disabled

* Thu Jul 28 2022 Ricardo Noriega <rnoriega@redhat.com> 4.10.0
- Add NetworkManager configuration file

* Tue Jul 26 2022 Miguel Angel Ajo Pelayo <majopela@redhat.com> 4.10.0
- Move crio.conf.d/microshift-ovn.conf to microshift-networking

* Tue Jul 26 2022 Zenghui Shi <zshi@redhat.com> 4.10.0
- Restart NetworkManager before OVS configuration

* Fri Jul 22 2022 Miguel Angel Ajo Pelayo <majopela@redhat.com> 4.10.0
- Add the jq dependency

* Thu Jul 21 2022 Miguel Angel Ajo Pelayo <majopela@redhat.com> 4.10.0
- Remove ovs duplicated services to set CPUAffinity with systemd .d dirs

* Thu Jul 21 2022 Ricardo Noriega <rnoriega@redhat.com> 4.10.0
- Adding microshift-ovn.conf with CRI-O network and workload partitioning

* Wed Jul 20 2022 Miguel Angel Ajo Pelayo <majopela@redhat.com> 4.10.0
- Add microshift-ovs-init script as oneshot during boot

* Tue Jul 19 2022 Miguel Angel Ajo <majopela@redhat.com> 4.10.0
- Adding the microshift-ovs-init systemd service and script which initializes br-ex and connects
  the main interface through it.

* Tue Jul 12 2022 Miguel Angel Ajo <majopela@redhat.com> 4.10.0
- Adding the networking subpackage to support ovn-networking
- Adding virtual openvswitch systemd files with CPUAffinity=0
- Setting OVS_USER_OPT to --no-mlockall in /etc/sysconfig/openvswitch

* Tue May 24 2022 Ricardo Noriega <rnoriega@redhat.com> 4.10.0
- Adding hostpath-provisioner.service to set SElinux policies to the volumes directory

* Fri May 6 2022 Sally O'Malley <somalley@redhat.com> 4.10.0
- Update required golang version to 1.17

* Mon Feb 7 2022 Ryan Cook <rcook@redhat.com> 4.8.0
- Selinux directory creation and labeling

* Wed Feb 2 2022 Ryan Cook <rcook@redhat.com> 4.8.0
- Define specific selinux policy version to help manage selinux package

* Wed Feb 2 2022 Miguel Angel Ajo <majopela@redhat.com> 4.8.0
- Remove the microshift-containerized subpackage, our docs explain how to download the .service file,
  and it has proven problematic to package this.
- Fix the microshift.service being overwritten by microshift-containerized, even when the non-containerized
  package only is installed.

* Thu Nov 4 2021 Miguel angel Ajo <majopela@redhat.com> 4.8.0
- Add microshift-containerized subpackage which contains the microshift-containerized systemd
  definition.

* Thu Nov 4 2021 Miguel Angel Ajo <majopela@redhat.com> 4.8.0
- Include the cleanup-all-microshift-data script for convenience

* Thu Sep 23 2021 Miguel Angel Ajo <majopela@redhat.com> 4.7.0
- Support commit based builds
- workaround rpmbuild with no build in place support
- add missing BuildRequires on systemd and policycoreutils

* Mon Sep 20 2021 Miguel Angel Ajo <majopela@redhat.com> 4.7.0
- Initial packaging
