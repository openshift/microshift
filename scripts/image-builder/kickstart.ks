lang en_US.UTF-8
keyboard us
timezone UTC
zerombr
clearpart --all --initlabel
autopart --type=plain --fstype=xfs --nohome
reboot
text
network --bootproto=dhcp --device=link --activate --onboot=on

ostreesetup --nogpg --osname=rhel --remote=edge --url=file:///run/install/repo/ostree/repo --ref=rhel/8/x86_64/edge

%post --log=/var/log/anaconda/post-install.log --erroronfail

echo -e 'url=http://REPLACE_OSTREE_SERVER_IP/repo/' >> /etc/ostree/remotes.d/edge.conf

useradd -m -d /home/redhat -p \$5\$XDVQ6DxT8S5YWLV7\$8f2om5JfjK56v9ofUkUAwZXTxJl3Sqnc9yPnza4xoJ0 -G wheel redhat

echo -e 'redhat\tALL=(ALL)\tNOPASSWD: ALL' >> /etc/sudoers

firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16

%end
