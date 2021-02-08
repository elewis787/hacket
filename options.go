package hacket

import "time"

// packetOptions config options for packets
type packetOptions struct {
	WriteBufferSize  int
	ReadBufferSize   int
	WriteDeadline    time.Duration
	ReadDeadline     time.Duration
	ConcurrencyLimit uint32
}

// Options interface for applying service options
type Options interface {
	apply(*packetOptions)
}

// funcPacketServiceOption wraps a function that modifies packetOptions into an
// implementation of the PacketOption interface.
type funcPacketServiceOption struct {
	f func(*packetOptions)
}

func (fpso *funcPacketServiceOption) apply(opt *packetOptions) {
	fpso.f(opt)
}

func newFuncPacketOption(f func(*packetOptions)) *funcPacketServiceOption {
	return &funcPacketServiceOption{
		f: f,
	}
}

// WithWriteBufferSize used to set the size of the operating system's
// transmit buffer associated with the connection. A zero value indicates
// usage of the system default socket size
func WithWriteBufferSize(s int) Options {
	return newFuncPacketOption(func(o *packetOptions) {
		o.WriteBufferSize = s
	})
}

// WithReadBufferSize used to set the size of the operating system's
// receive buffer associated with the connection. A zero value indicates
// usage of the system default socket size
func WithReadBufferSize(s int) Options {
	return newFuncPacketOption(func(o *packetOptions) {
		o.ReadBufferSize = s
	})
}

// WithWriteDeadline duration until conn's write deadline is triggered
func WithWriteDeadline(t time.Duration) Options {
	return newFuncPacketOption(func(o *packetOptions) {
		o.WriteDeadline = t
	})
}

// WithReadDeadline duration until conn's read deadline is triggered
func WithReadDeadline(t time.Duration) Options {
	return newFuncPacketOption(func(o *packetOptions) {
		o.ReadDeadline = t
	})
}

// WithConcurrencyLimit number of concurrent rountines created while process packets
func WithConcurrencyLimit(l uint32) Options {
	return newFuncPacketOption(func(o *packetOptions) {
		o.ConcurrencyLimit = l
	})
}

func defaultPacketOption() *packetOptions {
	return &packetOptions{
		ReadBufferSize:   0, // use go upd socket size default
		WriteBufferSize:  0, // use go udp socket size default
		WriteDeadline:    0, // no deadline by default
		ReadDeadline:     0, // no deadline by default
		ConcurrencyLimit: 1, // process one packet at a time
	}
}
