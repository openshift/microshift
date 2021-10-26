package oauthclient

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	oauthapi "github.com/openshift/api/oauth/v1"
)

// Getter exposes a way to get a specific client.  This is useful for other registries to get scope limitations
// on particular clients.   This interface will make its easier to write a future cache on it
type Getter interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*oauthapi.OAuthClient, error)
}
