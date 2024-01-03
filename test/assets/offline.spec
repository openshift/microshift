Name: microshift-offline-config
Version: 1
Summary: Host and MicroShift service configuration for offline environments
Source: microshift-offline-config.tar.gz
Release: 1
License: ASL 2.0
URL: https://github.com/openshift/microshift
BuildArch: noarch

Requires: NetworkManager

%global nmscriptdir %{_sysconfdir}/NetworkManager

%description
Host and MicroShift service configuration for offline environments

%prep
%setup

%files
%config(noreplace) %{nmscriptdir}/system-connections/lo.connection
%config(noreplace) %{nmscriptdir}/system-connections/enp0s1.connection
%config(noreplace) %{_sysconfdir}/microshift/config.yaml
%config(noreplace) %{_sysconfdir}/resolv.conf
%config(noreplace) %{_sysconfdir}/hosts

%build

%install
install -m 0755 %{_builddir} %{buildroot}