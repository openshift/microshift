lang en_US.UTF-8
keyboard us
timezone UTC
text
reboot

# Configure network to use DHCP and activate on boot
network --bootproto=dhcp --device=link --activate --onboot=on --hostname=microshift-starter.local --noipv6

# Partition disk with a 1GB boot XFS partition and a 8GB LVM volume containing system root
# The remainder of the volume will be used by the CSI driver for storing data
zerombr
clearpart --all --initlabel
part /boot --fstype=xfs --asprimary --size=1024
part pv.01 --grow
volgroup rhel pv.01
logvol / --vgname=rhel --fstype=xfs --size=8192 --name=root

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

# Update selinux-policy packages from CentOS 8 Stream
CENTOS8BASE=http://mirror.centos.org/centos/8-stream/BaseOS/x86_64/os/Packages
curl -LO -s $CENTOS8BASE/selinux-policy-3.14.3-96.el8.noarch.rpm
curl -LO -s $CENTOS8BASE/selinux-policy-devel-3.14.3-96.el8.noarch.rpm
curl -LO -s $CENTOS8BASE/selinux-policy-targeted-3.14.3-96.el8.noarch.rpm
dnf localinstall -y selinux-policy*.rpm
rm -f selinux-policy*.rpm

# Install MicroShift testing package
dnf copr enable -y @redhat-et/microshift-testing
dnf install -y microshift

# MicroShift service should be enabled later after setting up CRI-O with the pull secret

# Configure firewalld
firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16
firewall-offline-cmd --zone=trusted --add-source=169.254.169.1

# Install the oc and kubectl utilities (need a 4.12 dev-preview version to have the new functionality support)
curl -LO -s https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp-dev-preview/4.12.0-ec.4/openshift-client-linux-4.12.0-ec.4.tar.gz
tar zxf openshift-client-linux-4.12.0-ec.4.tar.gz -C /usr/local/bin/
rm -f openshift-client-linux-4.12.0-ec.4.tar.gz

%end
