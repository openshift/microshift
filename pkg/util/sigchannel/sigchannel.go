package sigchannel

// IsClosed tests whether a signalling channel has been closed.
// Note: Must only be used on broadcast signalling channels, i.e. channels
//       that only ever get closed, not sent any values.
func IsClosed(channel <-chan struct{}) bool {
	select {
	case <-channel:
		return true
	default:
		return false
	}
}

// AllClosed tests whether all signalling channels have been closed.
// Note: Must only be used on broadcast signalling channels, i.e. channels
//       that only ever get closed, not sent any values.
func AllClosed(channels []<-chan struct{}) bool {
	for _, channel := range channels {
		if !IsClosed(channel) {
			return false
		}
	}
	return true
}

// And returns a signalling channel that will be closed when all operand
// signalling channels have been closed.
func And(channels []<-chan struct{}) <-chan struct{} {
	andChannel := make(chan struct{})

	go func() {
		defer close(andChannel)
		for {
			if AllClosed(channels) {
				break
			}
		}
	}()

	return andChannel
}
