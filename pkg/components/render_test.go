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
	"github.com/openshift/microshift/pkg/release"
)

func Test_renderTopolvmDaemonsetTemplate(t *testing.T) {
	tb := embedded.MustAsset("components/lvms/topolvm-node_daemonset.yaml")

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
				data: renderParamsFromConfig(config.NewDefault(), assets.RenderParams{"SocketName": "/run/lvmd/lvmd.socket", "lvmd": "foobar"}),
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
