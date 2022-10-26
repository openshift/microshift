package components

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"text/template"

	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/lvmd"
	"github.com/openshift/microshift/pkg/release"
)

func Test_renderLvmdParams(t *testing.T) {

	iToP := func(i int) *uint64 { r := uint64(i); return &r }

	type args struct {
		l *lvmd.Lvmd
	}
	tests := []struct {
		name    string
		args    args
		want    assets.RenderParams
		wantErr bool
	}{
		{
			name: "should pass",
			args: args{
				l: &lvmd.Lvmd{
					SocketName: "/run/lvmd/lvmd.socket",
					DeviceClasses: []*lvmd.DeviceClass{
						{
							Name:        "test",
							VolumeGroup: "vg",
							Default:     true,
							SpareGB:     iToP(5),
						},
					}},
			},
			wantErr: false,
			want: assets.RenderParams{
				"SocketName": "/run/lvmd/lvmd.socket",
				"lvmd":       `"device-classes:\n- default: true\n  lvcreate-options: null\n  name: test\n  spare-gb: 5\n  stripe: null\n  stripe-size: \"\"\n  thin-pool: null\n  type: \"\"\n  volume-group: vg\nsocket-name: /run/lvmd/lvmd.socket\n"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderLvmdParams(tt.args.l)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderLvmdParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("renderLvmdParams() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renderLvmdConfig(t *testing.T) {

	defL, _ := lvmd.NewLvmdConfigFromFileOrDefault("")

	cmWrap := func(d []byte) []byte {
		base :=
			`# Source: topolvm/templates/lvmd/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: lvmd
  namespace: openshift-storage
data:
  lvmd.yaml:`
		return []byte(fmt.Sprintf("%s %q\n", base, d))
	}

	type args struct {
		tb   []byte
		data assets.RenderParams
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "should render topolvm configMap",
			args: args{
				tb:   embedded.MustAsset("components/odf-lvm/topolvm-lvmd-config_configmap_v1.yaml"),
				data: func() assets.RenderParams { p, _ := renderLvmdParams(defL); return p }(),
			},
			want:    cmWrap([]byte("device-classes:\n- default: true\n  lvcreate-options: null\n  name: default\n  spare-gb: 10\n  stripe: null\n  stripe-size: \"\"\n  thin-pool: null\n  type: \"\"\n  volume-group: rhel\nsocket-name: /run/lvmd/lvmd.socket\n")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate(tt.args.tb, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("renderTemplate() got = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}

func Test_renderTopolvmDaemonsetTemplate(t *testing.T) {
	tb := embedded.MustAsset("components/odf-lvm/topolvm-node_daemonset.yaml")

	fm := template.FuncMap{
		"Dir": filepath.Dir,
		"Sha256sum": func(s string) string {
			return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
		},
	}
	tpl := template.New("")
	tpl.Funcs(fm)

	template.Must(tpl.Parse(string(tb)))

	wantBytes := func(tpl *template.Template, data map[string]interface{}) []byte {
		buf := new(bytes.Buffer)
		err := template.Must(tpl.Clone()).Execute(buf, data)
		if err != nil {
			t.Fatalf("test table init: %v", err)
		}
		return buf.Bytes()
	}

	type args struct {
		tb   []byte
		data assets.RenderParams
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "renders lvmd-socket-name path",
			args: args{
				tb:   tb,
				data: renderParamsFromConfig(config.NewMicroshiftConfig(), assets.RenderParams{"SocketName": "/run/lvmd/lvmd.socket", "lvmd": "foobar"}),
			},
			want: wantBytes(tpl, map[string]interface{}{
				"ReleaseImage": release.Image,
				"SocketName":   "/run/lvmd/lvmd.socket",
				"lvmd":         "foobar",
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate(tt.args.tb, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("renderTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
