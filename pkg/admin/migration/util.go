package migration

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/net"
)

func getObjectMeta(item *unstructured.Unstructured) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
	}
}

func canRetry(err error) bool {
	err = interpret(err)
	if temp, ok := err.(TemporaryError); ok && !temp.Temporary() {
		return false
	}
	return true
}

func interpret(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.IsNotFound(err):
		// if the object is deleted, there is no need to migrate
		return nil
	case errors.IsMethodNotSupported(err):
		return ErrNotRetriable{err}
	case errors.IsConflict(err):
		return ErrRetriable{err}
	case errors.IsServerTimeout(err):
		return ErrRetriable{err}
	case errors.IsTooManyRequests(err):
		return ErrRetriable{err}
	case net.IsProbableEOF(err):
		return ErrRetriable{err}
	case net.IsConnectionReset(err):
		return ErrRetriable{err}
	case net.IsNoRoutesError(err):
		return ErrRetriable{err}
	case isConnectionRefusedError(err):
		return ErrRetriable{err}
	default:
		return err
	}
}

func isConnectionRefusedError(err error) bool {
	return strings.Contains(err.Error(), "connection refused")
}
