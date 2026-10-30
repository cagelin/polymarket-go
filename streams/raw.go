package streams

import "time"

// RawFrame retains a complete WebSocket market message in a freshly allocated buffer.
// Data must remain immutable: the typed decoder may read it after the callback.
// The SDK does not reuse the buffer; callers may retain it after the callback returns.
// Sequence starts at one and increases across market reconnections within a client.
// Generation identifies a connection across observed services within that client.
// Neither identifier is unique across client instances or process restarts.
// ReceivedAt is the local observation time after a complete read, not upstream time.
// RawOnly bypasses market typed decoding. Close waits for all callbacks to return.
type RawFrame struct {
	Service    string
	Generation uint64
	Sequence   uint64
	ReceivedAt time.Time
	Data       []byte
}

// Lifecycle describes an observed connection boundary or application heartbeat.
// Service is market, user, sports or rtds; price sessions are not observed.
// State is connected, disconnected, ping, pong or read_limit. Ping records an
// attempted application heartbeat write; pong records a market/user text PONG.
// The read_limit state reports an oversized message; partial bytes are not delivered.
// A disconnected observation also occurs on intentional shutdown and does not
// identify a cause, enumerate missing messages or guarantee lossless recovery.
// Generation is shared with RawFrame. At is the local observation time.
type Lifecycle struct {
	Service    string
	Generation uint64
	State      string
	At         time.Time
}

func (c *Client) observe(service string, generation uint64, state string) {
	if c.config.LifecycleHandler != nil {
		c.config.LifecycleHandler(Lifecycle{Service: service, Generation: generation, State: state, At: time.Now().UTC()})
	}
}
