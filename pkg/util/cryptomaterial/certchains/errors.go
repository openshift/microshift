package certchains

import (
	"errors"
	"fmt"
)

var _ error = &SignerNotFound{}

type SignerNotFound struct {
	name string
}

func NewSignerNotFound(signerName string) *SignerNotFound {
	return &SignerNotFound{
		name: signerName,
	}
}

func (e *SignerNotFound) Error() string {
	return fmt.Sprintf("signer %q was not found", e.name)
}

func IsSignerNotFoundError(err error) bool {
	var t *SignerNotFound
	return errors.As(err, &t)
}
