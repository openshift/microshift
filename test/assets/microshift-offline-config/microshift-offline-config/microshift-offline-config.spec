Name: microshift-offline-config
Version: 1
Summary: Host and MicroShift service configuration for offline environments
Source: microshift-offline-config.tar.gz
Release: 1
License: ASL 2.0
URL: https://github.com/openshift/microshift
BuildArch: noarch

Requires: NetworkManager

%global config_root %{_builddir}/microshift-offline-config

%description
Host and MicroShift service configuration for offline environments

%prep
%setup -n microshift-offline-config

%build
%install
install -m 0755 -d %{buildroot}%{_sysconfdir}/microshift
install -m 0755 -d %{buildroot}%{_sysconfdir}/NetworkManager/system-connections
install -m 0755 -d %{buildroot}/usr/lib/NetworkManager/system-connections
cp %{config_root}%{_sysconfdir}/hosts %{buildroot}/%{_sysconfdir}/hosts
cp %{config_root}%{_sysconfdir}/resolv.conf %{buildroot}/%{_sysconfdir}/resolv.conf
cp %{config_root}%{_sysconfdir}/microshift/config.yaml %{buildroot}/%{_sysconfdir}/microshift/config.yaml
cp %{config_root}/usr/lib/NetworkManager/system-connections/enp0s1.connection %{buildroot}/usr/lib/NetworkManager/system-connections/enp0s1.connection
cp %{config_root}%{_sysconfdir}/NetworkManager/system-connections/lo.connection %{buildroot}%{_sysconfdir}/NetworkManager/system-connections/lo.connection

%files
%config(noreplace) /usr/lib/NetworkManager/system-connections/enp0s1.connection
%config(noreplace) %{_sysconfdir}/NetworkManager/system-connections/lo.connection
%config(noreplace) %{_sysconfdir}/microshift/config.yaml
%config(noreplace) %{_sysconfdir}/resolv.conf
%config(noreplace) %{_sysconfdir}/hosts

