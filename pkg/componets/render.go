package render

import (
	"bytes"
	"text/template"
)

// from https://github.com/openshift/okd/releases/tag/4.5.0-0.okd-2020-10-15-235428
const (
	imageCvoOperator       = "quay.io/openshift/okd-content@sha256:b807a37a507ad2919b353a7c5d7af353f7bcceddd166cb352531ba6336bfe7ed"
	imageDNSOperator       = "quay.io/openshift/okd-content@sha256:04ed5d9d6fb14c0005c7074b659d6587c117da0e7e3b98f506f6cf3440c45358"
	versionOperator        = "4.5.0-0.okd-2020-10-15-235428"
	imageCoreDNS           = "quay.io/openshift/okd-content@sha256:cf916e742d1a01f5632b42e1cdc7d918385eb85408faf32d3c92796f9faa6234"
	imageOC                = "quay.io/openshift/okd-content@sha256:e87ccf31c42554e2b62d5a441a68307043576dfc1c509b8e72fadc815591d3d2"
	imageKubeRbacProxy     = "quay.io/openshift/okd-content@sha256:1aa5bb03d0485ec2db2c7871a1eeaef83e9eabf7e9f1bc2c841cf1a759817c99"
	imageServiceCAOperator = "quay.io/openshift/okd-content@sha256:d5ab863a154efd4014b0e1d9f753705b97a3f3232bd600c0ed9bde71293c462e"
)

func RenderSCController(b []byte) ([]byte, error) {
	data := struct {
		ImageSCOperator, KeyDir, CADir string
	}{
		ImageSCOperator: imageServiceCAOperator,
		CADir:           "",
		KeyDir:          "",
	}
	tpl := template.Must(template.New("sc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}

func RenderDNSService(b []byte) ([]byte, error) {
	data := struct {
		ClusterIP string
	}{
		ClusterIP: constants.ClusterDNSClusterIP,
	}
	tpl := template.Must(template.New("svc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}
