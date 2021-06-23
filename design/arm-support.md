# Design for ARM64 Support

## Problem Statement
Currently Microshift is built on x86_64 platforms and uses x86_64 OKD container images.
Give that ARM chips are widely used in Edge computing devices, Microshift needs to support
ARM platforms by the following improvements:

* Cross compilation. Microshift needs to be compiled into ARM support.
* Use ARM based container images.

## Design

The goal is to enable both ARM32 and ARM64 but priority given to 64 bits first.

### Cross compilation

Microshift itself can be cross compiled to ARM64 platform by adding `GOARCH=arm64`. 
This can be done in the build script. 

### Using ARM based container image
The container images for OpenShift components such as `service-ca`, `openshift-ingress`, and `openshift-router` come from x86_64 based OKD repository. 
The images are:
```console
# grep okd assets/ -r
assets/apps/0000_80_openshift-router-deployment.yaml:          image: quay.io/openshift/okd-content@sha256:5908265eb0041cea9a9ec36ad7b2bc82dd45346fc9e0f1b34b0e38a0f43f9f18
assets/apps/0000_70_dns_01-daemonset.yaml:        image: quay.io/openshift/okd-content@sha256:fb7eafdcb7989575119e1807e4adc2eb29f8165dec5c148b9c3a44d48458d8a7
assets/apps/0000_70_dns_01-daemonset.yaml:        image: quay.io/openshift/okd-content@sha256:1aa5bb03d0485ec2db2c7871a1eeaef83e9eabf7e9f1bc2c841cf1a759817c99
assets/apps/0000_70_dns_01-daemonset.yaml:        image: quay.io/openshift/okd-content@sha256:b20d195c721cd3b6215e5716b5569cbabbe861559af7dce07b5f8f3d38e6d701
assets/apps/0000_60_service-ca_05_deploy.yaml:        image: quay.io/openshift/okd-content@sha256:d5ab863a154efd4014b0e1d9f753705b97a3f3232bd600c0ed9bde71293c462e
```

To support more platforms, the images paths needs to parameterized and be conditionally compiled:

* The OKD image names used under `assets` are to be replaced by golang template parameters.
* The manifests are rendered in runtime in the [render functions](https://github.com/redhat-et/microshift/blob/main/pkg/components/render.go)
* Values for images paths are defined in the [constant pkg](https://github.com/redhat-et/microshift/blob/main/pkg/constant/constant.go) under build tags, i.e. `// +build !arm64` for x86_64 platforms, and `// +build !amd64` for ARM targets.

### ARM based container images
As of now, there are no ARM based containers images for the OpenShift components. More investigations are required to get these images or build from source in their current forms.

## Exit criteria

A successfully deployed Microshift binary or container can run on a Raspberry Pi, preferrably 64 bits.
