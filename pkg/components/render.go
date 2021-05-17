package components

import (
	"bytes"
	"text/template"

	"github.com/openshift/microshift/pkg/constant"
)

func renderSCController(b []byte) ([]byte, error) {
	data := struct {
		ImageSCOperator, KeyDir, CADir string
	}{
		ImageSCOperator: constant.ImageServiceCAOperator,
		KeyDir:          "/etc/kubernetes/ushift-resources/service-ca/secrets/service-ca",
		CADir:           "/etc/kubernetes/ushift-certs/ca-bundle",
	}
	tpl := template.Must(template.New("sc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}

func renderDNSService(b []byte) ([]byte, error) {
	data := struct {
		ClusterIP string
	}{
		ClusterIP: constant.ClusterDNS,
	}
	tpl := template.Must(template.New("svc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}
