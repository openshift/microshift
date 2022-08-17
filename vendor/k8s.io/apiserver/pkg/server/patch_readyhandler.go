package server

import (
	"net"
	"sync"
	goatomic "sync/atomic"
)

// readyOnlyLoggingListener wraps the given listener to mark late connections
// as such, identified by the remote address. In parallel, we have a filter that
// logs bad requests through these connections. We need this filter to get
// access to the http path in order to filter out healthz or readyz probes that
// are allowed at any point during readyOnly.
//
// Connections are late after the lateStopCh has been closed.
type readyOnlyLoggingListener struct {
	net.Listener
	lateStopCh <-chan struct{}
}

type eventfFunc func(eventType, reason, messageFmt string, args ...interface{})

var (
	lateConnectionRemoteAddrsLock sync.RWMutex
	lateConnectionRemoteAddrs     = map[string]bool{}

	unexpectedRequestsEventf goatomic.Value
)

func (l *readyOnlyLoggingListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	select {
	case <-l.lateStopCh:
		lateConnectionRemoteAddrsLock.Lock()
		defer lateConnectionRemoteAddrsLock.Unlock()
		lateConnectionRemoteAddrs[c.RemoteAddr().String()] = true
	default:
	}

	return c, nil
}