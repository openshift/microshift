package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func Test_listTestResultFor(t *testing.T) {
	type args struct {
		clusterProfile string
		featureGate    string
	}
	tests := []struct {
		name    string
		args    args
		want    map[JobVariant]*TestingResults
		wantErr bool
	}{
		{
			name: "test example",
			args: args{
				clusterProfile: "SelfManagedHA",
				featureGate:    "Example",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "platform example",
			args: args{
				clusterProfile: "VSphereGate",
				featureGate:    "Example",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "optional platform example",
			args: args{
				featureGate:    "NutanixGate",
				clusterProfile: "SelfManagedHA",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "install example",
			args: args{
				featureGate:    "FooBarInstall",
				clusterProfile: "Example",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("this is for ease of manual testing")

			got, err := listTestResultFor(tt.args.featureGate, sets.New[string](tt.args.clusterProfile))
			if (err != nil) != tt.wantErr {
				t.Errorf("listTestResultFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				for _, jobVariantResult := range got {
					result, serializationErr := json.MarshalIndent(jobVariantResult, "", "  ")
					if serializationErr != nil {
						t.Log(serializationErr.Error())
					}
					t.Log(string(result))
				}
				t.Errorf("listTestResultFor() got = %v, want %v", got, tt.want)
			}
		})
	}
}
