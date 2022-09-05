package components

import (
	"bytes"
	"text/template"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/release"
)

var templateFuncs = map[string]interface{}{}

func renderParamsFromConfig(cfg *config.MicroshiftConfig, extra assets.RenderParams) assets.RenderParams {
	params := map[string]interface{}{
		"ReleaseImage":  release.Image,
		"NodeName":      cfg.NodeName,
		"NodeIP":        cfg.NodeIP,
		"ClusterCIDR":   cfg.Cluster.ClusterCIDR,
		"ServiceCIDR":   cfg.Cluster.ServiceCIDR,
		"ClusterDNS":    cfg.Cluster.DNS,
		"ClusterDomain": cfg.Cluster.Domain,
	}
	for k, v := range extra {
		params[k] = v
	}
	return params
}

func renderTemplate(tb []byte, data assets.RenderParams) ([]byte, error) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(string(tb))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderServiceCAController(b []byte, p assets.RenderParams) ([]byte, error) {
	data := struct {
		ReleaseImage           map[string]string
		CAConfigMap, TLSSecret string
	}{
		ReleaseImage: release.Image,
		CAConfigMap:  p["ConfigMap"].(string),
		TLSSecret:    p["Secret"].(string),
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
		ReleaseImage map[string]string
		ClusterIP    string
	}{
		ReleaseImage: release.Image,
		ClusterIP:    p["ClusterDNS"].(string),
	}
	tpl := template.Must(template.New("svc").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}

func renderOVNKManifests(b []byte, p assets.RenderParams) ([]byte, error) {
	data := struct {
		ReleaseImage   map[string]string
		ClusterCIDR    string
		ServiceCIDR    string
		KubeconfigPath string
		KubeconfigDir  string
	}{
		ReleaseImage:   release.Image,
		ClusterCIDR:    p["ClusterCIDR"].(string),
		ServiceCIDR:    p["ServiceCIDR"].(string),
		KubeconfigPath: p["KubeconfigPath"].(string),
		KubeconfigDir:  p["KubeconfigDir"].(string),
	}
	tpl := template.Must(template.New("cm").Parse(string(b)))
	var byteBuff bytes.Buffer

	if err := tpl.Execute(&byteBuff, data); err != nil {
		return nil, err
	}
	return byteBuff.Bytes(), nil
}

func renderReleaseImage(b []byte, p assets.RenderParams) ([]byte, error) {
	data := struct {
		ReleaseImage map[string]string
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
