package embedded

import (
	"embed"
	"io/fs"
)

//go:embed components controllers core crd version release
var content embed.FS

func Asset(name string) ([]byte, error) {
	return content.ReadFile(name)
}

func AssetStreamed(name string) (fs.File, error) {
	return content.Open(name)
}

func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}
