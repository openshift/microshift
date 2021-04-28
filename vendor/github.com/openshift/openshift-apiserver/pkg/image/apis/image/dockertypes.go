package image

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DockerImage is the type representing a container image and its various properties when
// retrieved from the Docker client API.
type DockerImage struct {
	metav1.TypeMeta

	ID              string
	Parent          string
	Comment         string
	Created         metav1.Time
	Container       string
	ContainerConfig DockerConfig
	DockerVersion   string
	Author          string
	Config          *DockerConfig
	Architecture    string
	Size            int64
}

// DockerConfig is the list of configuration options used when creating a container.
type DockerConfig struct {
	Hostname        string
	Domainname      string
	User            string
	Memory          int64
	MemorySwap      int64
	CPUShares       int64
	CPUSet          string
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	PortSpecs       []string
	ExposedPorts    map[string]struct{}
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             []string
	DNS             []string
	Image           string
	Volumes         map[string]struct{}
	VolumesFrom     string
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	SecurityOpts    []string
	OnBuild         []string
	Labels          map[string]string
}
