Name: microshift-images

# disable dynamic rpmbuild checks
%global __os_install_post /bin/true
%global __arch_install_post /bin/true
AutoReqProv: no

# where do we want the images to be stored on the final system
%global imageStore /opt/microshift/images
%global imageStoreSed %(echo %{imageStore} | sed 's/\//\\\//g')

%define version %(echo %{baseVersion} | sed s/-/_/g)

# to-be-improved:
# avoid warnings for container layers:
#   - warning: absolute symlink:
#   - warning: Duplicate build-ids

Version: %{version}
Release: 2

Summary: MicroShift related container images
License: Apache License 2.0
URL: https://github.com/redhat-et/microshift

BuildRequires: podman
Requires: crio


%description
This rpm creates a custom RO container storage for the MicroShift container images
and pull images and add path to additional container image stores.

%prep


if [ -d  %{buildroot}%{imageStore} ]
then
   sudo rm -rf  %{buildroot}%{imageStore}
fi

%build


%install

mkdir -p %{buildroot}%{imageStore}

%define arch %{_arch}

# aarch64 is arm64 for container regisitries

%ifarch %{arm} aarch64
%define arch arm64
%endif

pull_arch="--arch %{arch}"

# for x86_64 we don't want to specify the arch otherwise quay gets grumpy

%ifarch x86_64
pull_arch=""
images=%{images_x86_64}
%endif

%ifarch %{arm}
images=%{images_arm}
%endif

%ifarch %{arm} aarch64
images=%{images_arm64}
%endif

%ifarch ppc64le
images=%{images_ppc64le}
%endif

%ifarch riscv64
images=%{images_riscv64}
%endif


for val in ${images}; do
    podman pull ${pull_arch} --root %{buildroot}%{imageStore} $val
done

# check, why do we need this?
# sudo chmod -R a+rx  %{imageStore}

%post

# only on install (1), not on upgrades (2)
if [ $1 -eq 1 ]; then
   sed -i '/^additionalimagestores =*/a "%{imageStore}",' /etc/containers/storage.conf
   # if crio was already started, restart it so it read from new imagestore
   systemctl is-active --quiet crio && systemctl restart --quiet crio
fi

%postun

# only on uninstall (0), not on upgrades(1)
if [ $1 -eq 0 ];
  sed -i '/"${imageStoreSed}",/d" /etc/containers/storage.conf
  systemctl is-active --quiet crio && systemctl restart --quiet crio

fi

%files
%{imageStore}/*

%changelog
* Wed Mar 2 2022 Miguel Angel Ajo <majopela@redhat.com> . 4.8.0_0.okd_2021_10_10_030117-2
- Automatically get architecture images and OKD base version

* Wed Feb 16 2022 Parul Singh <parsingh@redhat.com> . 4.8.0-0.microshiftr-2022_02_02_194009_3
- Initial packaging of additional RO container storage.
