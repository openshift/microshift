package build

type Build interface {
	Execute(*Opts) error
}

type build struct {
	Name string
	Path string
}
