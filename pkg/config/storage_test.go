package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_IsEnabled(t *testing.T) {
	type fields struct {
		Driver        CSIStorageDriver
		CSIComponents []OptionalCsiComponent
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "is disabled when driver is none",
			fields: fields{
				Driver: CsiDriverNone,
			},
			want: false,
		},
		{
			name: "is enabled when driver is lvms",
			fields: fields{
				Driver: CsiDriverLVMS,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				Driver:                tt.fields.Driver,
				OptionalCSIComponents: tt.fields.CSIComponents,
			}
			assert.Equalf(t, tt.want, s.IsEnabled(), "IsEnabled()")
		})
	}
}

func TestStorage_IsValid(t *testing.T) {
	type fields struct {
		Driver        CSIStorageDriver
		CSIComponents []OptionalCsiComponent
	}
	tests := []struct {
		name   string
		fields fields
		want   []error
	}{
		{
			name: "is valid when a driver is set and csi-components are valid",
			fields: fields{
				Driver:        CsiDriverLVMS,
				CSIComponents: []OptionalCsiComponent{CsiComponentSnapshot, CsiComponentSnapshotWebhook},
			},
			want: []error{},
		},
		{
			name: "is valid when a driver unset and csi-components are valid",
			fields: fields{
				Driver:        CsiDriverUnset,
				CSIComponents: []OptionalCsiComponent{CsiComponentSnapshot},
			},
			want: []error{},
		},
		{
			name: "is valid when the driver is unset and csi-components are valid",
			fields: fields{
				Driver:        CsiDriverUnset,
				CSIComponents: []OptionalCsiComponent{CsiComponentSnapshot, CsiComponentSnapshotWebhook},
			},
			want: []error{},
		},
		{
			name: "is invalid when driver is valid, but csi-components are invalid",
			fields: fields{
				Driver:        CsiDriverLVMS,
				CSIComponents: []OptionalCsiComponent{"foobar", CsiComponentSnapshot, CsiComponentSnapshotWebhook},
			},
			want: []error{
				fmt.Errorf("invalid CSI components: [foobar]"),
			},
		},
		{
			name: "is invalid when driver is invalid, but CSI components are valid",
			fields: fields{
				Driver:        "foobar",
				CSIComponents: []OptionalCsiComponent{CsiComponentSnapshot, CsiComponentSnapshotWebhook},
			},
			want: []error{
				fmt.Errorf("invalid driver \"foobar\""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				Driver:                tt.fields.Driver,
				OptionalCSIComponents: tt.fields.CSIComponents,
			}
			got := s.IsValid()
			assert.Equalf(t, tt.want, got, "IsValid()")
		})
	}
}

func TestStorage_csiComponentIsValid(t *testing.T) {
	type fields struct {
		Driver        CSIStorageDriver
		CSIComponents []OptionalCsiComponent
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "passes for all csi-component values",
			fields: fields{
				CSIComponents: []OptionalCsiComponent{
					CsiComponentSnapshot,
					CsiComponentSnapshotWebhook,
				},
			},
			want: []string{},
		},
		{
			name: "passes when unsupported values are caught and returned",
			fields: fields{
				CSIComponents: []OptionalCsiComponent{"FOOBAR", CsiComponentSnapshot},
			},
			want: []string{"FOOBAR"},
		},
		{
			name: "passes for empty array",
			fields: fields{
				CSIComponents: []OptionalCsiComponent{},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				Driver:                tt.fields.Driver,
				OptionalCSIComponents: tt.fields.CSIComponents,
			}
			got := s.csiComponentsAreValid()
			assert.Equalf(t, tt.want, got, "csiComponentsAreValid()")
		})
	}
}

func TestStorage_driverIsValid(t *testing.T) {
	type fields struct {
		Driver        CSIStorageDriver
		CSIComponents []OptionalCsiComponent
	}
	tests := []struct {
		name            string
		fields          fields
		wantIsSupported bool
	}{
		{
			name: "is valid when value matches one of predefined drivers",
			fields: fields{
				Driver: CsiDriverLVMS,
			},
			wantIsSupported: true,
		},
		{
			name: "is valid when value is an empty string",
			fields: fields{
				Driver: "",
			},
			wantIsSupported: true,
		},
		{
			name: "is invalid when value does not match one of predefined drivers",
			fields: fields{
				Driver: "foobar",
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				Driver:                tt.fields.Driver,
				OptionalCSIComponents: tt.fields.CSIComponents,
			}
			assert.Equalf(t, tt.wantIsSupported, s.driverIsValid(), "driverIsValid()")
		})
	}
}

func TestStorage_noneIsMutuallyExclusive(t *testing.T) {
	type fields struct {
		Driver        CSIStorageDriver
		CSIComponents []OptionalCsiComponent
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
		want1  bool
	}{
		{
			name: "passes when none is the sole value",
			fields: fields{
				CSIComponents: []OptionalCsiComponent{CsiComponentNone},
			},
			want:  nil,
			want1: true,
		},
		{
			name: "fails when none is not the sole value",
			fields: fields{
				CSIComponents: []OptionalCsiComponent{CsiComponentNone, CsiComponentSnapshot},
			},
			want:  []string{string(CsiComponentSnapshot)},
			want1: false,
		},
		{
			name: "passes when none is not in a list of values",
			fields: fields{
				CSIComponents: []OptionalCsiComponent{CsiComponentSnapshot, CsiComponentSnapshotWebhook},
			},
			want:  nil,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				Driver:                tt.fields.Driver,
				OptionalCSIComponents: tt.fields.CSIComponents,
			}
			got, got1 := s.noneIsMutuallyExclusive()
			assert.Equalf(t, tt.want, got, "noneIsMutuallyExclusive()")
			assert.Equalf(t, tt.want1, got1, "noneIsMutuallyExclusive()")
		})
	}
}
