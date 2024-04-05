package build

import "context"

type Build interface {
	Prepare(*Opts) error
	Execute(context.Context, *Opts) error
}

type build struct {
	Name string
	Path string
}
