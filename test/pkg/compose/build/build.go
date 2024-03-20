package build

// TODO: Signal handling: INT -> cancel build/download

type Build interface {
	Prepare(*Opts) error
	Execute(*Opts) error
}

type build struct {
	Name string
	Path string
}
