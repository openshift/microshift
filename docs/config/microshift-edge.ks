lang en_US.UTF-8
keyboard us
timezone UTC
text
reboot

# Configure network to use DHCP and activate on boot
network --bootproto=dhcp --device=link --activate --onboot=on

# Partition the disk with hardware-specific boot partitions, adding an LVM
# volume that contains a 10GB+ system root. The remainder of the volume will
# be used by the LVMS CSI driver for storing data.
zerombr
clearpart --all --initlabel
reqpart --add-boot
part pv.01 --grow
volgroup rhel pv.01
logvol / --vgname=rhel --fstype=xfs --size=10240 --name=root

# Lock root user account
rootpw --lock

# Deploy the ostree commit embedded in the edge-installer ISO
ostreesetup --nogpg --osname=rhel --remote=edge --url=file:///run/install/repo/ostree/repo --ref=rhel/9/x86_64/edge

# Post install configuration
%post --log=/dev/console --erroronfail

# Create a default redhat user, allowing it to run sudo commands without password
useradd -m -d /home/redhat -p \$5\$XDVQ6DxT8S5YWLV7\$8f2om5JfjK56v9ofUkUAwZXTxJl3Sqnc9yPnza4xoJ0 redhat
echo -e 'redhat\tALL=(ALL)\tNOPASSWD: ALL' > /etc/sudoers.d/microshift

# Import Red Hat public keys to allow RPM GPG check (not necessary if a system is registered)
if ! subscription-manager status >& /dev/null ; then
   rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-*
fi

# Make the KUBECONFIG from MicroShift directly available for the root user
echo -e 'export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig' >> /root/.bash_profile

# Configure systemd journal service to persist logs between boots and limit their size to 1G
sudo mkdir -p /etc/systemd/journald.conf.d
cat > /etc/systemd/journald.conf.d/microshift.conf <<EOF
[Journal]
Storage=persistent
SystemMaxUse=1G
RuntimeMaxUse=1G
EOF

# Make sure all the Ethernet network interfaces are connected automatically
# by removing autoconnect option from the configuration files
find /etc/NetworkManager -name '*.nmconnection' -print0 | while IFS= read -r -d $'\0' file ; do
    if grep -qE '^type=ethernet' "${file}" ; then
        sed -i '/autoconnect=.*/d' "${file}"
    fi
done

# Work around bootupd not installing the EFI grub.cfg during ostree deployment.
# Without these files the UEFI firmware loads grubx64.efi but GRUB drops to a
# shell because it cannot find its configuration.
if [ -d /boot/efi/EFI/redhat ] && [ -f /usr/lib/bootupd/grub2-static/grub-static-efi.cfg ] && [ -f /boot/grub2/bootuuid.cfg ]; then
    cp /usr/lib/bootupd/grub2-static/grub-static-efi.cfg /boot/efi/EFI/redhat/grub.cfg
    cp /boot/grub2/bootuuid.cfg /boot/efi/EFI/redhat/bootuuid.cfg
fi

%end
