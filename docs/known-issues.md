# On EC2 RHEL 8.4

If you want to run `microshift` on EC2 REHL 8.4(`cat /etc/os-release`), you might find [`ingress and service-ca will not stay online`](https://github.com/redhat-et/microshift/issues/270).

Inside the failing pods, you might find errors as: `10.43.0.1:443: read: connection timed out`.

This a [known issue](https://bugzilla.redhat.com/show_bug.cgi?id=1912236#c30) on RHEL 8.4 and will be resolved in 8.5.

In order to work on RHEL 8.4, you may disable the networkManager and reboot to resolve this issue.

Eg:

```
systemctl disable nm-cloud-setup.service nm-cloud-setup.timer
reboot
```

You can find the details of this EC2 networkManage issue tracked at [issue](https://gitlab.freedesktop.org/NetworkManager/NetworkManager/-/issues/740).



