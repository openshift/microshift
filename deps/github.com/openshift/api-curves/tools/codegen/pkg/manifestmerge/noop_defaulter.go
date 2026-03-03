package manifestmerge

import "k8s.io/apimachinery/pkg/runtime"

type noopDefaulter struct {
}

func (d noopDefaulter) Default(in runtime.Object) {
}
