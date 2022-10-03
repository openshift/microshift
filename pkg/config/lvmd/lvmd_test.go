package lvmd

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_newLvmdConfigFromFile(t *testing.T) {

	iToP := func(i int) *uint64 {
		r := uint64(i)
		return &r
	}

	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    *Lvmd
		wantErr bool
	}{
		{
			name: "empty string set",
			args: args{
				p: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid config path",
			args: args{
				p: "./test/lvmd.yaml",
			},
			want: &Lvmd{
				SocketName: "/run/lvmd/lvmd.socket",
				DeviceClasses: []*DeviceClass{
					{
						Name:        "gold",
						Default:     true,
						VolumeGroup: "vg_1",
						SpareGB:     iToP(5),
					},
					{
						Name:        "silver",
						Default:     false,
						VolumeGroup: "vg_2",
						SpareGB:     iToP(5),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLvmdConfigFromFile(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("newLvmdConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && !reflect.DeepEqual(*got, *tt.want) {
				g, _ := json.Marshal(got)
				w, _ := json.Marshal(tt.want)
				t.Errorf("newLvmdConfigFromFile() = %s, want %s", string(g), string(w))
			}
		})
	}
}
