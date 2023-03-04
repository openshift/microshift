package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/util"
)

const (
	DefaultSystemdResolvedFile = "/run/systemd/resolve/resolv.conf"
)

type Config struct {
	DNS       DNS        `json:"dns"`
	Network   Network    `json:"network"`
	Node      Node       `json:"node"`
	ApiServer ApiServer  `json:"apiServer"`
	Etcd      EtcdConfig `json:"etcd"`
	Debugging Debugging  `json:"debugging"`

	// Internal-only fields
	Ingress IngressConfig `json:"-"`
}

var allHostnames []string

func getAllHostnames() ([]string, error) {
	if len(allHostnames) != 0 {
		return allHostnames, nil
	}
	cmd := exec.Command("/bin/hostname", "-A")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error when executing 'hostname -A': %v", err)
	}
	outString := out.String()
	outString = strings.Trim(outString[:len(outString)-1], " ")
	// Remove duplicates to avoid having them in the certificates.
	names := strings.Split(outString, " ")
	set := sets.NewString(names...)
	allHostnames = set.List()
	return allHostnames, nil
}

func NewMicroshiftConfig() *Config {
	nodeName, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Failed to get hostname %v", err)
	}
	nodeIP, err := util.GetHostIP()
	if err != nil {
		klog.Fatalf("failed to get host IP: %v", err)
	}
	subjectAltNames, err := getAllHostnames()
	if err != nil {
		klog.Fatalf("failed to get all hostnames: %v", err)
	}

	return &Config{
		Debugging: Debugging{
			LogLevel: "Normal",
		},
		ApiServer: ApiServer{
			SubjectAltNames: subjectAltNames,
			URL:             "https://localhost:6443",
		},
		Node: Node{
			HostnameOverride: strings.ToLower(nodeName),
			NodeIP:           nodeIP,
		},
		DNS: DNS{
			BaseDomain: "example.com",
		},
		Network: Network{
			ClusterNetwork: []ClusterNetworkEntry{
				{
					CIDR: "10.42.0.0/16",
				},
			},
			ServiceNetwork: []string{
				"10.43.0.0/16",
			},
			ServiceNodePortRange: "30000-32767",
			DNS:                  "10.43.0.10",
		},
		Etcd: EtcdConfig{
			MemoryLimitMB:           0,                 // No limit
			MinDefragBytes:          100 * 1024 * 1024, // 100MB
			MaxFragmentedPercentage: 45,                // percent
			DefragCheckFreq:         5 * time.Minute,
			DoStartupDefrag:         true,
			QuotaBackendBytes:       8 * 1024 * 1024 * 1024, // 8GB
		},
	}
}
