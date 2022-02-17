%global imageStore %{buildroot}/var/lib/microshift/images

Name: microshift-images
Version: 1
Release: 1
# this can be %{timestamp}.git%{short_hash} later for continous main builds
Summary: Create custom container storage
License: ASL 2.0
URL: https://github.com/redhat-et/microshift

BuildRequires:  podman
BuildRequires: crio


%description
This rpm creates a custom RO container storage and pull images and add path to additional container image stores.

%prep

if [ -d  %{imageStore} ] 
then
   sudo rm -rf  %{imageStore}
fi


%build


%install

mkdir -p %{imageStore}

declare -a ListOfImages=("quay.io/openshift/okd-content@sha256:27f7918b5f0444e278118b2ee054f5b6fadfc4005cf91cb78106c3f5e1833edd" \
"quay.io/openshift/okd-content@sha256:bcdefdbcee8af1e634e68a850c52fe1e9cb31364525e30f5b20ee4eacb93c3e8" \
"quay.io/openshift/okd-content@sha256:01cfbbfdc11e2cbb8856f31a65c83acc7cfbd1986c1309f58c255840efcc0b64" \
"quay.io/coreos/flannel:v0.14.0" \
"quay.io/microshift/flannel-cni:4.8.0-0.okd-2021-10-10-030117" \
"quay.io/openshift/okd-content@sha256:459f15f0e457edaf04fa1a44be6858044d9af4de276620df46dc91a565ddb4ec" \
"quay.io/kubevirt/hostpath-provisioner:v0.8.0" \
"k8s.gcr.io/pause" \
"quay.io/openshift/okd-content@sha256:dd1cd4d7b1f2d097eaa965bc5e2fe7ebfe333d6cbaeabc7879283af1a88dbf4e")

for val in ${ListOfImages[@]}; do
   sudo podman pull  --root %{imageStore} $val
done
sudo chmod -R a+rx  %{imageStore}


%post
sudo sed -i '/^additionalimagestores =*/a "/var/lib/microshift/images",' /etc/containers/storage.conf
# if crio was already started, restart it so it read from new imagestore
systemctl is-active --quiet crio && systemctl restart --quiet crio


%files
/var/lib/microshift/images/*



%changelog
* Wed Feb 16 2022 Parul Singh <parsingh@redhat.com> . 4.8.0-0.microshiftr-2022_02_02_194009_3
- Initial packaging of additional RO container storage.