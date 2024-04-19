package apiserver

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"testing"

	auditV1 "k8s.io/apiserver/pkg/apis/audit/v1"
)

func TestGetPolicy(t *testing.T) {
	type args struct {
		forProfile string
	}
	tests := []struct {
		name    string
		args    args
		want    *auditV1.Policy
		wantErr bool
	}{
		{
			name:    "providing a profile should return a policy",
			args:    args{forProfile: "Default"},
			want:    &auditV1.Policy{},
			wantErr: false,
		},
		{
			name:    "providing an unknown profile should return error",
			args:    args{forProfile: "NOEXIST"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPolicy(tt.args.forProfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			bs, err := yaml.Marshal(got)
			if err != nil {
				t.Errorf("Got err: %v", err)
			}
			if len(bs) == 0 {
				t.Errorf("Got empty bytes object")
			}
		})
	}
}
