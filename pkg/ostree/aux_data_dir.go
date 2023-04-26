package ostree

import "os"

const (
	auxDir  = "/var/lib/microshift.aux"
	dirPerm = os.FileMode(0644)
)

func EnsureAuxDirExists() error {
	return os.MkdirAll(auxDir, dirPerm)
}
