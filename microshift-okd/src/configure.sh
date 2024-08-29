#!/bin/bash


    cat > /etc/microshift/config.yaml <<EOF
storage:
    driver: "none"
EOF

# disable ovsdb
rm -rf /usr/lib/systemd/system/ovs-vswitchd.service
rm -rf /usr/lib/systemd/system/openvswitch.service

#removes ExecStart=/usr/bin/grub2-editenv - set boot_success=1
rm -rf /usr/lib/systemd/system/greenboot-grub2-set-success.service

systemctl enable microshift