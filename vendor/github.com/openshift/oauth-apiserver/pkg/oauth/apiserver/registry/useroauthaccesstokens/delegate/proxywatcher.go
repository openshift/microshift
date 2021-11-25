package delegate

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"

	oauthapi "github.com/openshift/oauth-apiserver/pkg/oauth/apis/oauth"
)

// OAuthAccessTokenWatcher implements a Watcher interface to watch OAuthAccessTokens
// and filter them to match only those for the current user
type OAuthAccessTokenWatcher struct {
	wrappedWatcher watch.Interface
	outgoing       chan watch.Event
	incoming       <-chan watch.Event
	stopCh         chan struct{}

	username string

	mutex   sync.Mutex
	stopped bool
}

var _ watch.Interface = &OAuthAccessTokenWatcher{}

// NewOAuthAccessTokenWatcher creates new OAuthAccessTokenWatcher by wrapping a channel of watched
// oauthaccesstokens. The username parameter allows filtering the oauthaccesstokens from the incoming
// channel for that specific user.
func newOAuthAccessTokenWatcher(wrappedWatcher watch.Interface, username string) *OAuthAccessTokenWatcher {
	return &OAuthAccessTokenWatcher{
		wrappedWatcher: wrappedWatcher,
		incoming:       wrappedWatcher.ResultChan(),
		outgoing:       make(chan watch.Event),
		stopCh:         make(chan struct{}),
		stopped:        false,

		username: username,
	}
}

// Stop implements Interface
func (w *OAuthAccessTokenWatcher) Stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.stopped {
		w.stopped = true
		w.wrappedWatcher.Stop()
		close(w.stopCh)
	}
}

// ResultChan returns channel for filtered results
func (w *OAuthAccessTokenWatcher) ResultChan() <-chan watch.Event {
	return w.outgoing
}

func (w *OAuthAccessTokenWatcher) Watch(ctx context.Context) {
	defer close(w.outgoing)
	defer utilruntime.HandleCrash()

	for {
		select {
		case <-w.stopCh:
			return
		case event := <-w.incoming:
			switch event.Type {
			case watch.Error:
				w.outgoing <- event
			default:
				tokenOrig, ok := event.Object.(*oauthapi.OAuthAccessToken)
				if !ok {
					w.outgoing <- createErrorEvent(errors.NewInternalError(fmt.Errorf("failed to convert incoming object to an OAuthAccessToken type")))
					continue
				}

				if !isValidUserToken(tokenOrig, w.username) {
					continue
				}
				event.Object = (*oauthapi.UserOAuthAccessToken)(tokenOrig)
				w.outgoing <- event
			}
		case <-ctx.Done(): // user cancel
			w.Stop()
			return
		}
	}
}

func createErrorEvent(err errors.APIStatus) watch.Event {
	status := err.Status()
	return watch.Event{
		Type:   watch.Error,
		Object: &status,
	}
}
