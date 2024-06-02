package templatingdata

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/klog/v2"
)

// Shim is a struct that works with legacy templates using gomplate
type Shim struct {
	Env map[string]string
}

func NewShim(tplData *TemplatingData) *Shim {
	s := &Shim{
		Env: map[string]string{
			"UNAME_M": tplData.Arch,

			"LOCAL_REPO":            tplData.Source.Repository,
			"NEXT_REPO":             tplData.FakeNext.Repository,
			"BASE_REPO":             tplData.Base.Repository,
			"CURRENT_RELEASE_REPO":  tplData.Current.Repository,
			"PREVIOUS_RELEASE_REPO": tplData.Previous.Repository,

			"FAKE_NEXT_MINOR_VERSION": strconv.Itoa(tplData.FakeNext.Minor),
			"MINOR_VERSION":           strconv.Itoa(tplData.Source.Minor),
			"PREVIOUS_MINOR_VERSION":  strconv.Itoa(tplData.Previous.Minor),
			"YMINUS2_MINOR_VERSION":   strconv.Itoa(tplData.YMinus2.Minor),

			"SOURCE_VERSION":           tplData.Source.Version,
			"SOURCE_VERSION_BASE":      tplData.Base.Version,
			"CURRENT_RELEASE_VERSION":  tplData.Current.Version,
			"PREVIOUS_RELEASE_VERSION": tplData.Previous.Version,
			"YMINUS2_RELEASE_VERSION":  tplData.YMinus2.Version,

			"SOURCE_IMAGES": strings.Join(tplData.Source.Images, ","),
		},
	}

	if tplData.RHOCPMinorY != 0 {
		s.Env["RHOCP_MINOR_Y"] = strconv.Itoa(tplData.RHOCPMinorY)
	}
	if tplData.RHOCPMinorY1 != 0 {
		s.Env["RHOCP_MINOR_Y1"] = strconv.Itoa(tplData.RHOCPMinorY1)
	}
	if tplData.RHOCPMinorY2 != 0 {
		s.Env["RHOCP_MINOR_Y2"] = strconv.Itoa(tplData.RHOCPMinorY2)
	}

	return s
}

type FakeEnv struct {
	s *Shim
}

func (fe *FakeEnv) Getenv(k string, def string) string {
	v, ok := fe.s.Env[k]
	if !ok {
		return def
	}
	return v
}

func (s *Shim) Template(name, data string) (string, error) {
	klog.InfoS("Templating input text", "template", name, "preTemplating", data)

	fe := &FakeEnv{s: s}
	funcs := map[string]any{
		"hasPrefix": strings.HasPrefix,
		"env":       func() interface{} { return fe },
	}

	tpl, err := template.New(name).Funcs(funcs).Parse(data)
	if err != nil {
		klog.ErrorS(err, "Failed to parse template file", "template", name)
		return "", fmt.Errorf("failed to parse template %q: %w", name, err)
	}

	b := &strings.Builder{}
	err = tpl.Execute(b, s)
	if err != nil {
		klog.ErrorS(err, "Executing template failed", "template", name)
		return "", fmt.Errorf("failed to execute template %q: %w", name, err)
	}

	result := b.String()
	klog.InfoS("Templating successful", "template", name, "postTemplating", result)

	return result, nil
}
