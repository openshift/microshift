package config

import "testing"

func Test_validateFeatureSets(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		c       *Config
		wantErr bool
	}{
		{
			name: "unrecognized featureSet should return an error",
			c: &Config{
				ApiServer: ApiServer{
					FeatureGates: FeatureGates{
						FeatureSet: "unrecognized",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "featureSet is CustomNoUpgrade and customNoUpgrade.[Enabled || Disabled] is empty should not return an error",
			c: &Config{
				ApiServer: ApiServer{
					FeatureGates: FeatureGates{
						FeatureSet: FeatureSetCustomNoUpgrade,
						CustomNoUpgrade: CustomNoUpgrade{
							Enabled:  []string{},
							Disabled: []string{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "customNoUpgrade.[Enabled || Disabled] can only be used when featureSet is CustomNoUpgrade",
			c: &Config{
				ApiServer: ApiServer{
					FeatureGates: FeatureGates{
						FeatureSet: "",
						CustomNoUpgrade: CustomNoUpgrade{
							Enabled:  []string{"feature1"},
							Disabled: []string{"feature2"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "customNoUpgrade.[Enabled && Disabled] cannot have the same feature gate",
			c: &Config{
				ApiServer: ApiServer{
					FeatureGates: FeatureGates{
						FeatureSet: FeatureSetCustomNoUpgrade,
						CustomNoUpgrade: CustomNoUpgrade{
							Enabled:  []string{"feature1"},
							Disabled: []string{"feature1"},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateFeatureGates(tt.c)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateFeatureSets() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateFeatureSets() succeeded unexpectedly")
			}
		})
	}
}

func Test_validateCustomNoUpgrade(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		c       *Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateFeatureGates(tt.c)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateCustomNoUpgrade() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateCustomNoUpgrade() succeeded unexpectedly")
			}
		})
	}
}
