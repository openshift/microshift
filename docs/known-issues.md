---
modified: "2021-10-25T11:10:27.304+02:00"
title: Known Issues
tags: known issues, troubleshooting
layout: page
toc: true
---

## On EC2 with RHEL 8.4

### `service-ca` can't be created

If you want to run `microshift` on EC2 RHEL 8.4(`cat /etc/os-release`), you might find [`ingress and service-ca will not stay online`](https://github.com/redhat-et/microshift/issues/270).

Inside the failing pods, you might find errors as: `10.43.0.1:443: read: connection timed out`.

This a [known issue](https://bugzilla.redhat.com/show_bug.cgi?id=1912236#c30) on RHEL 8.4 and will be resolved in 8.5.

In order to work on RHEL 8.4, you may disable the networkManager and reboot to resolve this issue.

Eg:

```sh
systemctl disable nm-cloud-setup.service nm-cloud-setup.timer
reboot
```

You can find the details of this EC2 networkManage issue tracked at [issue](https://gitlab.freedesktop.org/NetworkManager/NetworkManager/-/issues/740).

### Openshift pods restarts on `CrashLoopBackOff`

A few minutes after `microshift` started, openshift pods fall into `CrashLoopBackOff`.

If you check up the `journalctl |grep iptables`, you may see the following:

```log

Sep 21 19:12:54 ip-172-31-85-30.ec2.internal microshift[1297]: I0921 19:12:54.399365    1297 server_others.go:185] Using iptables Proxier.
Sep 21 19:13:50 ip-172-31-85-30.ec2.internal kernel: iptables[2438]: segfault at 88 ip 00007feaf5dc0e47 sp 00007fff6f2fea08 error 4 in libnftnl.so.11.3.0[7feaf5dbc000+16000]
Sep 21 19:13:50 ip-172-31-85-30.ec2.internal systemd-coredump[2442]: Process 2438 (iptables) of user 0 dumped core.
Sep 21 20:35:57 ip-172-31-85-30.ec2.internal microshift[1297]: E0921 20:35:57.914558    1297 remote_runtime.go:143] StopPodSandbox "1ae45abde0b46d8ea5176b6a00f0e5b4291e6bb496762ca25a4196a5f18d0475" from runtime service failed: rpc error: code = Unknown desc = failed to destroy network for pod sandbox k8s_service-ca-64547678c6-2nxnp_openshift-service-ca_6236deba-fc5f-4915-817d-f8699a4accfc_0(1ae45abde0b46d8ea5176b6a00f0e5b4291e6bb496762ca25a4196a5f18d0475): error removing pod openshift-service-ca_service-ca-64547678c6-2nxnp from CNI network "crio": running [/usr/sbin/iptables -t nat -D POSTROUTING -s 10.42.0.3 -j CNI-d5d0edec163ce01e4591c1c4 -m comment --comment name: "crio" id: "1ae45abde0b46d8ea5176b6a00f0e5b4291e6bb496762ca25a4196a5f18d0475" --wait]: exit status 2: iptables v1.8.4 (nf_tables): Chain 'CNI-d5d0edec163ce01e4591c1c4' does not exist
```

Also, the `openshift-ingress` pod will faild on:

```console
I0921 17:36:17.811391       1 router.go:262] router "msg"="router is including routes in all namespaces"
E0921 17:36:17.914638       1 haproxy.go:418] can't scrape HAProxy: dial unix /var/lib/haproxy/run/haproxy.sock: connect: no such file or directory
I0921 17:36:17.948417       1 router.go:579] template "msg"="router reloaded"  "output"=" - Checking http://localhost:80 ...\n - Health check ok : 0 retry attempt(s).\n"
```

As a workaround, you can follow steps below:

- delete `flannel` daemonset

  ```sh
  kubectl delete ds -n kube-system kube-flannel-ds
  ```

- restart all the openshift pods.

This workaround won't affect the single node `microshift` functionality since the `flannel` daemonset is used for multi-node microshift.

This issue is tracked at: [#296](https://github.com/redhat-et/microshift/issues/296)
