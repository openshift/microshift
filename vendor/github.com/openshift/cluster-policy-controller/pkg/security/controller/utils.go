package controller

import (
	"fmt"
	"strings"
)

const (
	namespaceNameKeyPrefix = "ns"
)

func asNamespaceNameKey(namespace string) (namespaceNameKey string) {
	return namespaceNameKeyPrefix + "/" + namespace
}

func parseNamespaceNameKey(key string) (namespace string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != 2 || parts[0] != namespaceNameKeyPrefix || parts[1] == "" {
		return "", fmt.Errorf("unexpected namespace name key format: %q", key)
	}

	return parts[1], nil
}
