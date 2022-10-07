package embedded

import "embed"

//go:embed components core crd scc version
var content embed.FS

func Asset(name string) ([]byte, error) {
	return content.ReadFile(name)
}

func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}
