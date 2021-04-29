package util

import (
	"k8s.io/apimachinery/pkg/util/net"
)

func GetHostIP() (string, error) {
	ip, err := net.ChooseHostInterface()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}
