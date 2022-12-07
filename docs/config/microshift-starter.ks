lang en_US.UTF-8
keyboard us
timezone UTC
text
reboot

# Configure network to use DHCP and activate on boot
network --bootproto=dhcp --device=link --activate --onboot=on --hostname=microshift-starter.local --noipv6

# Partition disk with a 1GB boot XFS partition and a 10GB LVM volume containing system root
# The remainder of the volume will be used by the CSI driver for storing data
zerombr
clearpart --all --initlabel
part /boot/efi --fstype=efi --size=200
part /boot --fstype=xfs --asprimary --size=800
part pv.01 --grow
volgroup rhel pv.01
logvol / --vgname=rhel --fstype=xfs --size=10240 --name=root

# Configure users
rootpw --lock
user --plaintext --name=redhat --password=redhat

# Minimal package setup
cdrom
%packages
@^minimal-environment
bash-completion
cockpit
conmon
conntrack-tools
containernetworking-plugins
containers-common
container-selinux
criu
git
jq
make
NetworkManager-ovs
python36
selinux-policy-devel
%end

# Post install configuration
%post --log=/var/log/anaconda/post-install.log --erroronfail

# Allow the default user to run sudo commands without password
echo -e 'redhat\tALL=(ALL)\tNOPASSWD: ALL' > /etc/sudoers.d/redhat

tee /etc/yum.repos.d/rhocp-4.12-el8-beta-$(uname -i)-rpms.repo >/dev/null <<EOF
[rhocp-4.12-el8-beta-$(uname -i)-rpms]
name=Beta rhocp-4.12 RPMs for RHEL8
baseurl=https://mirror.openshift.com/pub/openshift-v4/\$basearch/dependencies/rpms/4.12-el8-beta/
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF

# Install MicroShift testing package
dnf copr enable -y @redhat-et/microshift-testing
dnf install -y microshift
dnf install -y openshift-clients

# MicroShift service should be enabled later after setting up CRI-O with the pull secret

# Configure firewalld
firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16
firewall-offline-cmd --zone=trusted --add-source=169.254.169.1

%end
