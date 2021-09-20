%{!?github_tag: %global github_tag 4.7.0-0.microshift-2021-08-31-224727}
%{!?version: %global version 4.7.0}
%{!?release: %global release 2021_08_31_224727}

# golang specifics
%global golang_version 1.15
#debuginfo not supported with Go
%global debug_package %{nil}
# modifying the Go binaries breaks the DWARF debugging
%global __os_install_post %{_rpmconfigdir}/brp-compress

Name: microshift
Version: %{version}
Release: %{release}%{dist}
# this can be %{timestamp}.git%{short_hash} later for continous main builds
Summary: Microshift binary
License: ASL 2.0
URL: https://github.com/redhat-et/microshift

%if ! 0%{?local_build:1}
Source0: https://github.com/redhat-et/microshift/archive/refs/tags/%{github_tag}.tar.gz
%endif

%if 0%{?go_arches:1}
ExclusiveArch: %{go_arches}
%else
ExclusiveArch: x86_64 aarch64 ppc64le s390x
%endif

BuildRequires: gcc
BuildRequires: glibc-static
BuildRequires: golang >= %{golang_version}
BuildRequires: make

Requires: cri-o
Requires: cri-tools
%{?systemd_requires}

%description
Microshift is a research project that is exploring how OpenShift Kubernetes
can be optimized for small form factor and edge computing.

Edge devices deployed out in the field pose very different operational,
environmental, and business challenges from those of cloud computing.
These motivate different engineering
trade-offs for Kubernetes at the far edge than for cloud or near-edge
scenarios. Microshift's design goals cater to this:

make frugal use of system resources (CPU, memory, network, storage, etc.),
tolerate severe networking constraints, update (resp. roll back) securely,
safely, speedily, and seamlessly (without disrupting workloads), and build on
and integrate cleanly with edge-optimized OSes like Fedora IoT and RHEL for Edge,
while providing a consistent development and management experience with standard
OpenShift.

We believe these properties should also make Microshift a great tool for other
use cases such as Kubernetes applications development on resource-constrained
systems, scale testing, and provisioning of lightweight Kubernetes control planes.

Note: Microshift is still early days and moving fast. Features are missing.
Things break. But you can still help shape it, too.

%prep

# Unpack the sources, unless it's a localbuild
%if ! 0%{?local_build:1}
%setup -n microshift-%{github_tag}
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

make _build_local GOOS=${GOOS} GOARCH=${GOARCH}
cp ./_output/bin/${GOOS}_${GOARCH}/microshift ./_output/microshift

%install

install -d %{buildroot}%{_bindir}
install -p -m755 ./_output/microshift %{buildroot}%{_bindir}/microshift
restorecon -v %{buildroot}%{_bindir}/microshift

install -d -m755 $RPM_BUILD_ROOT/%{_unitdir}
install -p -m644 systemd/microshift.service %{buildroot}%{_unitdir}/microshift.service

mkdir -p -m755 %{buildroot}/var/run/flannel
mkdir -p -m755 %{buildroot}/var/run/kubelet
mkdir -p -m755 %{buildroot}/var/lib/kubelet/pods
mkdir -p -m755 %{buildroot}/var/run/secrets/kubernetes.io/serviceaccount
mkdir -p -m755 %{buildroot}/var/hpvolumes
restorecon -v %{buildroot}/var/hpvolumes

%post

%systemd_post microshift.service

%preun

%systemd_preun microshift.service


%files

%license LICENSE
%{_bindir}/microshift
%{_unitdir}/microshift.service

%changelog
* Mon Sep 20 2021 Miguel Angel Ajo <majopela@redhat.com> . 4.7.0-2021_08_31_224727
- Initial packaging
