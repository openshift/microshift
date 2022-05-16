# parameters that must be provided via --define , or fixed into the spec file:
# global version 4.7.0
# global release 2021_08_31_224727
# global github_tag 4.7.0-0.microshift-2021-08-31-224727
# global git_commit 81264d0ebb17fef06eff9ec7d4f2a81631c6b34a


# golang specifics
%global golang_version 1.17
#debuginfo not supported with Go
%global debug_package %{nil}
# modifying the Go binaries breaks the DWARF debugging
%global __os_install_post %{_rpmconfigdir}/brp-compress

# SELinux specifics
%global selinuxtype targeted
%define selinux_policyver 3.14.3-67
%define container_policyver 2.167.0-1
%define container_policy_epoch 2
%define microshift_relabel_files() \
   mkdir -p /var/hpvolumes; \
   mkdir -p /var/run/flannel; \
   mkdir -p /var/run/kubelet; \
   mkdir -p /var/lib/kubelet/pods; \
   mkdir -p /var/run/secrets/kubernetes.io/serviceaccount; \
   restorecon -R /var/hpvolumes; \
   restorecon -R /var/run/kubelet; \
   restorecon -R /var/run/flannel; \
   restorecon -R /var/lib/kubelet/pods; \
   restorecon -R /var/run/secrets/kubernetes.io/serviceaccount


# Git related details
%global shortcommit %(c=%{git_commit}; echo ${c:0:7})

Name: microshift
Version: %{version}
Release: %{release}%{dist}
# this can be %{timestamp}.git%{short_hash} later for continous main builds
Summary: MicroShift binary
License: ASL 2.0
URL: https://github.com/openshift/microshift

%if ! 0%{?local_build:1}%{?git_commit:1}
Source0: https://github.com/openshift/microshift/archive/refs/tags/%{github_tag}.tar.gz
%endif

%if 0%{?git_commit:1}
Source0: https://github.com/openshift/microshift/archive/%{git_commit}/microshift-%{shortcommit}.tar.gz
%endif


%if 0%{?go_arches:1}
ExclusiveArch: %{go_arches}
%else
ExclusiveArch: x86_64 aarch64 ppc64le s390x
%endif

BuildRequires: gcc
BuildRequires: golang >= %{golang_version}
BuildRequires: make
BuildRequires: policycoreutils
BuildRequires: systemd

Requires: cri-o
Requires: cri-tools
Requires: iptables
Requires: microshift-selinux
Requires: conntrack-tools

%{?systemd_requires}

%description
MicroShift is a research project that is exploring how OpenShift Kubernetes
can be optimized for small form factor and edge computing.

Edge devices deployed out in the field pose very different operational,
environmental, and business challenges from those of cloud computing.
These motivate different engineering
trade-offs for Kubernetes at the far edge than for cloud or near-edge
scenarios. MicroShift's design goals cater to this:

make frugal use of system resources (CPU, memory, network, storage, etc.),
tolerate severe networking constraints, update (resp. roll back) securely,
safely, speedily, and seamlessly (without disrupting workloads), and build on
and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge,
while providing a consistent development and management experience with standard
OpenShift.

We believe these properties should also make MicroShift a great tool for other
use cases such as Kubernetes applications development on resource-constrained
systems, scale testing, and provisioning of lightweight Kubernetes control planes.

Note: MicroShift is still early days and moving fast. Features are missing.
Things break. But you can still help shape it, too.

%package selinux
Summary: SELinux policies for MicroShift
BuildRequires: selinux-policy >= %{selinux_policyver}
BuildRequires: selinux-policy-devel >= %{selinux_policyver}
Requires: container-selinux >= %{container_policy_epoch}:%{container_policyver}
BuildArch: noarch
%{?selinux_requires}

%description selinux
SElinux policy modules for MicroShift.

%prep

# Unpack the sources, unless it's a localbuild (tag)
%if ! 0%{?local_build:1}%{?git_commit:1}
%setup -n microshift-%{github_tag}
%endif

# Unpack the sources, for a commit-based tarball
%if 0%{?git_commit:1}
%setup -n microshift-%{git_commit}
%endif

%build

GOOS=linux

%ifarch ppc64le
GOARCH=ppc64le
%endif

%ifarch %{arm} aarch64
GOARCH=arm64
%endif

%ifarch s390x
GOARCH=s390x
%endif

%ifarch x86_64
GOARCH=amd64
%endif

# if we have git commit/tag/state to be embedded in the binary pass it down to the makefile
%if 0%{?embedded_git_commit:1}
make _build_local GOOS=${GOOS} GOARCH=${GOARCH} EMBEDDED_GIT_COMMIT=%{embedded_git_commit} EMBEDDED_GIT_TAG=%{embedded_git_tag} EMBEDDED_GIT_TREE_STATE=%{embedded_git_tree_state}
%else
make _build_local GOOS=${GOOS} GOARCH=${GOARCH}
%endif

cp ./_output/bin/${GOOS}_${GOARCH}/microshift ./_output/microshift

# SELinux modules build

cd packaging/selinux
make

%install

install -d %{buildroot}%{_bindir}
install -p -m755 ./_output/microshift %{buildroot}%{_bindir}/microshift
install -p -m755 hack/cleanup.sh %{buildroot}%{_bindir}/cleanup-all-microshift-data

restorecon -v %{buildroot}%{_bindir}/microshift

install -d -m755 %{buildroot}%{_sysconfdir}/crio/crio.conf.d
install -p -m644 packaging/crio.conf.d/microshift.conf %{buildroot}%{_sysconfdir}/crio/crio.conf.d/microshift.conf

install -d -m755 %{buildroot}/%{_unitdir}
install -p -m644 packaging/systemd/microshift.service %{buildroot}%{_unitdir}/microshift.service

mkdir -p -m755 %{buildroot}/var/run/flannel
mkdir -p -m755 %{buildroot}/var/run/kubelet
mkdir -p -m755 %{buildroot}/var/lib/kubelet/pods
mkdir -p -m755 %{buildroot}/var/run/secrets/kubernetes.io/serviceaccount

install -d %{buildroot}%{_datadir}/selinux/packages/%{selinuxtype}
install -m644 packaging/selinux/microshift.pp.bz2 %{buildroot}%{_datadir}/selinux/packages/%{selinuxtype}

%post

%systemd_post microshift.service

# only for install, not on upgrades
if [ $1 -eq 1 ]; then
	# if crio was already started, restart it so it will catch /etc/crio/crio.conf.d/microshift.conf
	systemctl is-active --quiet crio && systemctl restart --quiet crio
fi

%post selinux

%selinux_modules_install -s %{selinuxtype} %{_datadir}/selinux/packages/%{selinuxtype}/microshift.pp.bz2
if /usr/sbin/selinuxenabled ; then
    %microshift_relabel_files
fi

%postun selinux

if [ $1 -eq 0 ]; then
    %selinux_modules_uninstall -s %{selinuxtype} microshift
fi

%posttrans selinux

%selinux_relabel_post -s %{selinuxtype}

%preun

%systemd_preun microshift.service


%files

%license LICENSE
%{_bindir}/microshift
%{_bindir}/cleanup-all-microshift-data
%{_unitdir}/microshift.service
%{_sysconfdir}/crio/crio.conf.d/microshift.conf

%files selinux

/var/run/flannel
/var/run/kubelet
/var/lib/kubelet/pods
/var/run/secrets/kubernetes.io/serviceaccount
%{_datadir}/selinux/packages/%{selinuxtype}/microshift.pp.bz2
%ghost %{_sharedstatedir}/selinux/%{selinuxtype}/active/modules/200/microshift

%changelog
* Fri May 7 2022 Sally O'Malley <somalley@redhat.com> . 4.10.0-0.microshift-2022-04-23-131357
- Update required golang version to 1.17

* Mon Feb 7 2022 Ryan Cook <rcook@redhat.com> . 4.8.0-0.microshiftr-2022_02_02_194009_3
- Selinux directory creation and labeling

* Wed Feb 2 2022 Ryan Cook <rcook@redhat.com> . 4.8.0-0.microshift-2022_01_04_175420_25
- Define specific selinux policy version to help manage selinux package

* Wed Feb 2 2022 Miguel Angel Ajo <majopela@redhat.com> . 4.8.0-0.microshift-2022-01-06-210147-20
- Remove the microshift-containerized subpackage, our docs explain how to download the .service file,
  and it has proven problematic to package this.
- Fix the microshift.service being overwritten by microshift-containerized, even when the non-containerized
  package only is installed.

* Thu Nov 4 2021 Miguel angel Ajo <majopela@redhat.com> . 4.8.0-nightly-14-g973b9c78
- Add microshift-containerized subpackage which contains the microshift-containerized systemd
  definition.

* Thu Nov 4 2021 Miguel Angel Ajo <majopela@redhat.com> . 4.8.0-nightly-13-g886705e5
- Include the cleanup-all-microshift-data script for convenience


* Thu Sep 23 2021 Miguel Angel Ajo <majopela@redhat.com> . 4.7.0-021_08_31_224727_40_g5c23735f
- Support commit based builds
- workaround rpmbuild with no build in place support
- add missing BuildRequires on systemd and policycoreutils

* Mon Sep 20 2021 Miguel Angel Ajo <majopela@redhat.com> . 4.7.0-2021_08_31_224727
- Initial packaging
