Name:       microshift-test-agent
Version:    0.0.1
Release:    1
Summary:    MicroShift Test Failure Agent
License:    ASL 2.0
BuildArch:  noarch
Source0:    test-agent

%description
todo

%install
install -d %{buildroot}%{_bindir}
install -p -m755  %{SOURCE0}/microshift-test-agent.sh %{buildroot}%{_bindir}/microshift-test-agent.sh
restorecon -v %{buildroot}%{_bindir}/microshift-test-agent.sh

install -d %{buildroot}%{_unitdir}
install -p -m644 %{SOURCE0}/microshift-test-agent.service %{buildroot}%{_unitdir}/microshift-test-agent.service

%files
%{_bindir}/microshift-test-agent.sh
%{_unitdir}/microshift-test-agent.service

%changelog
# TODO
