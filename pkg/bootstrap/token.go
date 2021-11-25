package bootstrap

import (
	"encoding/hex"
	"math/rand"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

func CreateTokenFile(path string) {
	err := os.Remove(path)
	if err != nil {
		klog.ErrorS(err, "Token file does not exist")
	}

	token := randString(16)
	token = token + ",kubelet-bootstrap,10001,\"system:bootstrappers\""
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		klog.ErrorS(err, "Token file cannot be created")
	}
	defer f.Close()
	f.Write([]byte(token))

}

func GetToken(path string) string {
	f, err := os.ReadFile(path)
	if err != nil {
		klog.ErrorS(err, "Token file cannot be opened")
	}
	token := f[:strings.IndexByte(string(f), ',')]

	return string(token)
}

func randString(length int) string {
	b := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
