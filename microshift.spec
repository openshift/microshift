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

Source0: https://github.com/redhat-et/microshift/archive/refs/tags/%{github_tag}.tar.gz

%if 0%{?go_arches:1}
ExclusiveArch: %{go_arches}
%else
ExclusiveArch: x86_64 aarch64 ppc64le s390x
%endif

BuildRequires: gcc
BuildRequires: glibc-static
BuildRequires: golang >= %{golang_version}
BuildRequires: make

%description
TBD

%prep
%setup -n microshift-%{github_tag}

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
install -p -m 755 ./_output/microshift %{buildroot}%{_bindir}/microshift

%files
%license LICENSE
%{_bindir}/microshift

%changelog

