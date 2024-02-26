package config

type IngressConfig struct {
	ServingCertificate []byte `json:"-"`
	ServingKey         []byte `json:"-"`
}
