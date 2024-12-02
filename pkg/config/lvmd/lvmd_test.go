package lvmd

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLvmdConfigForVGs(t *testing.T) {
	tests := []struct {
		name        string
		vgNames     []string
		expected    *Lvmd
		expectedErr error
	}{
		{
			name: "no groups",
			expected: &Lvmd{
				SocketName: defaultSockName,
				Message:    errorMessageNoVolumeGroups,
			},
		},
		{
			name:    "one group",
			vgNames: []string{"choose-me"},
			expected: &Lvmd{
				SocketName: defaultSockName,
				DeviceClasses: []*DeviceClass{
					{
						Name:        "default",
						VolumeGroup: "choose-me",
						Default:     true,
						SpareGB:     func() *uint64 { s := uint64(defaultSpareGB); return &s }(),
					},
				},
				Message: statusMessageDefaultAvailable,
			},
		},
		{
			name:    "one group default",
			vgNames: []string{defaultRHEL4EdgeVolumeGroup},
			expected: &Lvmd{
				SocketName: defaultSockName,
				DeviceClasses: []*DeviceClass{
					{
						Name:        "default",
						VolumeGroup: defaultRHEL4EdgeVolumeGroup,
						Default:     true,
						SpareGB:     func() *uint64 { s := uint64(defaultSpareGB); return &s }(),
					},
				},
				Message: statusMessageDefaultAvailable,
			},
		},
		{
			name:    "default first",
			vgNames: []string{defaultRHEL4EdgeVolumeGroup, "other"},
			expected: &Lvmd{
				SocketName: defaultSockName,
				DeviceClasses: []*DeviceClass{
					{
						Name:        "default",
						VolumeGroup: defaultRHEL4EdgeVolumeGroup,
						Default:     true,
						SpareGB:     func() *uint64 { s := uint64(defaultSpareGB); return &s }(),
					},
				},
				Message: statusMessageFoundDefault,
			},
		},
		{
			name:    "default last",
			vgNames: []string{"other", defaultRHEL4EdgeVolumeGroup},
			expected: &Lvmd{
				SocketName: defaultSockName,
				DeviceClasses: []*DeviceClass{
					{
						Name:        "default",
						VolumeGroup: defaultRHEL4EdgeVolumeGroup,
						Default:     true,
						SpareGB:     uint64Ptr(defaultSpareGB),
					},
				},
				Message: statusMessageFoundDefault,
			},
		},
		{
			name:    "no default",
			vgNames: []string{"other", "choose-me"},
			expected: &Lvmd{
				SocketName: defaultSockName,
				Message:    errorMessageMultipleVolumeGroups,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getLvmdConfigForVGs(tt.vgNames)
			assert.Equal(t, tt.expected, actual, "names: %v", tt.vgNames)
			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_newLvmdConfigFromFile(t *testing.T) {
	iToP := func(i uint64) *uint64 {
		return &i
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
				Message: "Read from ./test/lvmd.yaml",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLvmdConfigFromFile(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLvmdConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && !reflect.DeepEqual(*got, *tt.want) {
				g, _ := json.Marshal(got)
				w, _ := json.Marshal(tt.want)
				t.Errorf("NewLvmdConfigFromFile() = %s, want %s", string(g), string(w))
			}
		})
	}
}
