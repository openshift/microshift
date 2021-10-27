package components

import (
	"bytes"
	"text/template"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/release"
)

func renderSCController(b []byte, p assets.RenderParams) ([]byte, error) {
	data := struct {
		ReleaseImage  assets.RenderParams
		KeyDir, CADir string
	}{
		ReleaseImage: release.Image,
		KeyDir:       p["DataDir"] + "/resources/service-ca/secrets/service-ca",
		CADir:        p["DataDir"] + "/certs/ca-bundle",
	}
	tpl := template.Must(template.New("sc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}

func renderDNSService(b []byte, p assets.RenderParams) ([]byte, error) {
	data := struct {
		ReleaseImage assets.RenderParams
		ClusterIP    string
	}{
		ReleaseImage: release.Image,
		ClusterIP:    p["ClusterDNS"],
	}
	tpl := template.Must(template.New("svc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}

func renderReleaseImage(b []byte, p assets.RenderParams) ([]byte, error) {
	data := struct {
		ReleaseImage assets.RenderParams
	}{
		ReleaseImage: release.Image,
	}
	tpl := template.Must(template.New("dp").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}
