package psalabelsyncer

import (
	"testing"

	securityv1 "github.com/openshift/api/security/v1"
	corev1 "k8s.io/api/core/v1"
)

func Test_convert_hostDir(t *testing.T) {
	tests := []struct {
		name                     string
		allowHostDirVolumePlugin bool
		volumes                  []securityv1.FSType
		want                     uint8
	}{
		{
			name:                     "hostdir volume plugin means privileged",
			allowHostDirVolumePlugin: true,
			volumes:                  []securityv1.FSType{securityv1.FSProjected},
			want:                     privileged,
		},
		{
			name:                     "hostpath volumes means privileged",
			allowHostDirVolumePlugin: false,
			volumes: []securityv1.FSType{
				securityv1.FSTypeConfigMap,
				securityv1.FSTypeEmptyDir,
				securityv1.FSTypeHostPath,
			},
			want: privileged,
		},
		{
			name:                     "all volumes means privileged",
			allowHostDirVolumePlugin: false,
			volumes:                  []securityv1.FSType{securityv1.FSTypeAll},
			want:                     privileged,
		},
		{
			name:                     "anything else stays restricted",
			allowHostDirVolumePlugin: false,
			volumes: []securityv1.FSType{
				securityv1.FSPortworxVolume,
				securityv1.FSProjected,
				securityv1.FSTypeFlexVolume,
			},
			want: restricted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_hostDir(tt.allowHostDirVolumePlugin, tt.volumes); got != tt.want {
				t.Errorf("convert_hostDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_hostNamespace(t *testing.T) {
	tests := []struct {
		name             string
		allowHostIPC     bool
		allowHostNetwork bool
		allowHostPorts   bool
		want             uint8
	}{
		{
			name:         "IPC host namespace means privileged",
			allowHostIPC: true,
			want:         privileged,
		},
		{
			name:         "network host namespace means privileged",
			allowHostIPC: true,
			want:         privileged,
		},
		{
			name:         "ports host namespace means privileged",
			allowHostIPC: true,
			want:         privileged,
		},
		{
			name: "neither of the above is restricted",
			want: restricted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_hostNamespace(tt.allowHostIPC, tt.allowHostNetwork, tt.allowHostPorts); got != tt.want {
				t.Errorf("convert_hostNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_hostPorts(t *testing.T) {
	tests := []struct {
		name           string
		allowHostPorts bool
		want           uint8
	}{
		{
			name:           "allowing host ports means privileged",
			allowHostPorts: true,
			want:           privileged,
		},
		{
			name: "not allowing host ports means restricted",
			want: restricted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_hostPorts(tt.allowHostPorts); got != tt.want {
				t.Errorf("convert_hostPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_allowPrivilegeEscalation(t *testing.T) {
	pBool := func(b bool) *bool { return &b }
	tests := []struct {
		name                     string
		allowPrivilegeEscalation *bool
		want                     uint8
	}{
		{
			name:                     "allowing privilege escalation means baseline",
			allowPrivilegeEscalation: pBool(true),
			want:                     baseline,
		},
		{
			name: "no opinion on privilege escalation means baseline",
			want: baseline,
		},
		{
			name:                     "not allowing privilege escalation means restricted",
			allowPrivilegeEscalation: pBool(false),
			want:                     restricted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_allowPrivilegeEscalation(tt.allowPrivilegeEscalation); got != tt.want {
				t.Errorf("convert_allowPrivilegeEscalation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_allowPrivilegedContainer(t *testing.T) {
	tests := []struct {
		name                     string
		allowPrivilegedContainer bool
		want                     uint8
	}{
		{
			name:                     "allowing privileged containers means privileged",
			allowPrivilegedContainer: true,
			want:                     privileged,
		},
		{
			name:                     "not allowing privileged containers means restricted",
			allowPrivilegedContainer: false,
			want:                     restricted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_allowPrivilegedContainer(tt.allowPrivilegedContainer); got != tt.want {
				t.Errorf("convert_allowPrivilegedContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_allowedCapabilities(t *testing.T) {
	tests := []struct {
		name                     string
		sccAllowedCapabilities   []corev1.Capability
		requiredDropCapabilities []corev1.Capability
		want                     uint8
	}{
		{
			name:                     "drop all means restricted",
			requiredDropCapabilities: []corev1.Capability{"ALL"},
			want:                     restricted,
		},
		{
			name:                     "drop all with NET_BIND_SERVICE means restricted",
			sccAllowedCapabilities:   []corev1.Capability{"NET_BIND_SERVICE"},
			requiredDropCapabilities: []corev1.Capability{"ALL"},
			want:                     restricted,
		},
		{
			name:                     "drop all with any cap from baseline set means baseline",
			sccAllowedCapabilities:   []corev1.Capability{"NET_BIND_SERVICE", "MKNOD"},
			requiredDropCapabilities: []corev1.Capability{"ALL", "SYS_ADMIN"},
			want:                     baseline,
		},
		{
			name: "no drop caps required means baseline",
			want: baseline,
		},
		{
			name: "allowing caps from baseline set means baseline",
			sccAllowedCapabilities: []corev1.Capability{
				"KILL",
				"MKNOD",
				"NET_BIND_SERVICE",
				"SETFCAP",
				"SETGID",
			},
			want: baseline,
		},
		{
			name: "allowing any cap from outside the baseline set means privileged",
			sccAllowedCapabilities: []corev1.Capability{
				"KILL",
				"MKNOD",
				"NET_BIND_SERVICE",
				"SETFCAP",
				"SETGID",
				"NET_RAW", // <-- privileged
			},
			want: privileged,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_allowedCapabilities(tt.sccAllowedCapabilities, tt.requiredDropCapabilities); got != tt.want {
				t.Errorf("convert_allowedCapabilities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_unsafeSysctls(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name                 string
		allowedUnsafeSysctls []string
		want                 uint8
	}{
		{
			name: "no allowed unsafe sysctls means restricted",
			want: restricted,
		},
		{
			name: "allowed unsafe sysctls from the allowed set means restricted",
			allowedUnsafeSysctls: []string{
				"net.ipv4.ping_group_range",
				"net.ipv4.ip_unprivileged_port_start",
				"kernel.shm_rmid_forced",
			},
			want: restricted,
		},
		{
			name: "unsafe sysctls outside the allowed set means privileged",
			allowedUnsafeSysctls: []string{
				"net.ipv4.ping_group_range",
				"kernel.shm_rmid_forced",
				"net.ipv4.tcp_keepalive_time", // <-- privileged
			},
			want: privileged,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_unsafeSysctls(tt.allowedUnsafeSysctls); got != tt.want {
				t.Errorf("convert_unsafeSysctls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_runAsUser(t *testing.T) {
	pInt64 := func(i int64) *int64 {
		return &i
	}

	tests := []struct {
		name      string
		namespace *corev1.Namespace
		runAsUser *securityv1.RunAsUserStrategyOptions
		want      uint8
		wantErr   bool
	}{
		{
			name: "MustRunAs non-root user",
			runAsUser: &securityv1.RunAsUserStrategyOptions{
				Type: securityv1.RunAsUserStrategyMustRunAs,
				UID:  pInt64(1000),
			},
			want: restricted,
		},
		{
			name: "MustRunAs root user",
			runAsUser: &securityv1.RunAsUserStrategyOptions{
				Type: securityv1.RunAsUserStrategyMustRunAs,
				UID:  pInt64(0),
			},
			want: baseline,
		},
		{
			name: "MustRunAsRange not including root",
			runAsUser: &securityv1.RunAsUserStrategyOptions{
				Type:        securityv1.RunAsUserStrategyMustRunAsRange,
				UIDRangeMin: pInt64(1000),
				UIDRangeMax: pInt64(1010),
			},
			want: restricted,
		},
		{
			name: "MustRunAsRange including root",
			runAsUser: &securityv1.RunAsUserStrategyOptions{
				Type:        securityv1.RunAsUserStrategyMustRunAsRange,
				UIDRangeMin: pInt64(0),
				UIDRangeMax: pInt64(1010),
			},
			want: baseline,
		},
		{
			name: "RunAsNonroot",
			runAsUser: &securityv1.RunAsUserStrategyOptions{
				Type: securityv1.RunAsUserStrategyMustRunAsNonRoot,
			},
			want: restricted,
		},
		{
			name: "RunAsAny",
			runAsUser: &securityv1.RunAsUserStrategyOptions{
				Type: securityv1.RunAsUserStrategyRunAsAny,
			},
			want: baseline,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convert_runAsUser(tt.namespace, tt.runAsUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("convert_runAsUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convert_runAsUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_volumes(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name    string
		volumes []securityv1.FSType
		want    uint8
	}{
		{
			name: "hostpath means privileged",
			volumes: []securityv1.FSType{
				securityv1.FSProjected,
				securityv1.FSTypeHostPath,
				securityv1.FSTypeConfigMap,
				securityv1.FSTypeSecret,
			},
			want: privileged,
		},
		{
			name: "baseline volumes",
			volumes: []securityv1.FSType{
				securityv1.FSTypeSecret,
				securityv1.FSTypeCSI,
				securityv1.FSTypePersistentVolumeClaim,
				securityv1.FSTypeCephFS,
				securityv1.FSTypeFC,
			},
			want: baseline,
		},
		{
			name: "only restricted",
			volumes: []securityv1.FSType{
				securityv1.FSTypeSecret,
				securityv1.FSTypeCSI,
				securityv1.FSTypePersistentVolumeClaim,
				securityv1.FSTypeEphemeral,
			},
			want: restricted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_volumes(tt.volumes); got != tt.want {
				t.Errorf("convert_volumes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_seLinuxOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *securityv1.SELinuxContextStrategyOptions
		want uint8
	}{
		{
			name: "RunAsAny",
			opts: &securityv1.SELinuxContextStrategyOptions{
				Type: securityv1.SELinuxStrategyRunAsAny,
			},
			want: privileged,
		},
		{
			name: "MustRunAs with namespace-inherited opts",
			opts: &securityv1.SELinuxContextStrategyOptions{
				Type: securityv1.SELinuxStrategyMustRunAs,
			},
			want: restricted,
		},
		{
			name: "MustRunAs with allowed type",
			opts: &securityv1.SELinuxContextStrategyOptions{
				Type: securityv1.SELinuxStrategyMustRunAs,
				SELinuxOptions: &corev1.SELinuxOptions{
					Type: "container_t",
				},
			},
			want: restricted,
		},
		{
			name: "MustRunAs with custom type",
			opts: &securityv1.SELinuxContextStrategyOptions{
				Type: securityv1.SELinuxStrategyMustRunAs,
				SELinuxOptions: &corev1.SELinuxOptions{
					Type: "shell_exec_t",
				},
			},
			want: privileged,
		},
		{
			name: "MustRunAs with custom user",
			opts: &securityv1.SELinuxContextStrategyOptions{
				Type: securityv1.SELinuxStrategyMustRunAs,
				SELinuxOptions: &corev1.SELinuxOptions{
					Type: "container_t",
					User: "user_u",
				},
			},
			want: privileged,
		},
		{
			name: "MustRunAs with custom role",
			opts: &securityv1.SELinuxContextStrategyOptions{
				Type: securityv1.SELinuxStrategyMustRunAs,
				SELinuxOptions: &corev1.SELinuxOptions{
					Type: "container_t",
					Role: "role_r",
				},
			},
			want: privileged,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_seLinuxOptions(tt.opts); got != tt.want {
				t.Errorf("convert_seLinuxOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convert_seccompProfile(t *testing.T) {
	tests := []struct {
		name     string
		profiles []string
		want     uint8
	}{
		{
			name: "allow undefined",
			want: baseline,
		},
		{
			name:     "require runtime/default",
			profiles: []string{"runtime/default"},
			want:     restricted,
		},
		{
			name:     "allowed profiles",
			profiles: []string{"docker/default", "localhost/pepa", "localhost/zdenya"},
			want:     restricted,
		},
		{
			name:     "unconfined",
			profiles: []string{"docker/default", "localhost/pepa", "unconfined", "localhost/zdenya"},
			want:     privileged,
		},
		{
			name:     "unknown custom profile",
			profiles: []string{"super_custom_profile", "localhost/zdenya"},
			want:     privileged,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert_seccompProfile(tt.profiles); got != tt.want {
				t.Errorf("convert_seccompProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}
