package sigchannel

import (
	"testing"
	"time"
)

func TestIsClosed(t *testing.T) {
	ch := make(chan struct{})
	if IsClosed(ch) != false {
		t.Errorf("IsClosed() on new channel returned 'true' expected 'false'")
	}
	close(ch)
	if IsClosed(ch) != true {
		t.Errorf("IsClosed() on closed channel returned 'false' expected 'true'")
	}
}

func TestAllClosed(t *testing.T) {
	var newChannel = func() chan struct{} { return make(chan struct{}) }
	var newClosedChannel = func() chan struct{} { ch := make(chan struct{}); close(ch); return ch }

	var tests = []struct {
		channels []<-chan struct{}
		out      bool
	}{
		{[]<-chan struct{}{}, true},
		{[]<-chan struct{}{newChannel()}, false},
		{[]<-chan struct{}{newClosedChannel()}, true},
		{[]<-chan struct{}{newChannel(), newChannel()}, false},
		{[]<-chan struct{}{newChannel(), newClosedChannel()}, false},
		{[]<-chan struct{}{newClosedChannel(), newChannel()}, false},
		{[]<-chan struct{}{newClosedChannel(), newClosedChannel()}, true},
	}

	for i, test := range tests {
		want := test.out
		got := AllClosed(test.channels)
		if want != got {
			t.Errorf("for channel set %v got %v, want %v", i, got, want)
		}
	}
}

func toSendChannels(channels []chan struct{}) []<-chan struct{} {
	sendch := make([]<-chan struct{}, len(channels))
	for i, ch := range channels {
		sendch[i] = ch
	}
	return sendch
}

func TestAnd(t *testing.T) {
	channels := make([]chan struct{}, 2)
	channels[0] = make(chan struct{})
	channels[1] = make(chan struct{})
	ch := And(toSendChannels(channels))
	if IsClosed(ch) != false {
		t.Errorf("And() on new channels returned 'true' expected 'false'")
	}
	close(channels[0])
	for !IsClosed(channels[0]) {
		// as close is async, wait for close to complete
	}
	if IsClosed(ch) != false {
		t.Errorf("And() on with non-closed channels returned 'true' expected 'false'")
	}
	close(channels[1])
	for !IsClosed(channels[1]) {
		// as close is async, wait for close to complete
	}
	for {
		// wait until And() closes the channel or time out
		select {
		case <-ch:
			if IsClosed(ch) != true {
				t.Errorf("And() on closed channels returned 'true' expected 'false'")
			}
			return
		case <-time.After(time.Duration(time.Second * 5)):
			t.Errorf("timed out waiting for And() to close channels")
			return
		}
	}
}
