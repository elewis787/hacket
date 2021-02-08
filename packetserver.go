package hacket

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type atomicBool int32

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }

// PacketServer interface used to describe a packet server.
type PacketServer interface {
	Port() (int, error)
	Serve(handler PacketHandler) error
	Shutdown(ctx context.Context) error
}

var _ PacketServer = &udpPacketServerImpl{}

// packetServerImpl a connection less server that wraps a net.PacketConn
type udpPacketServerImpl struct {
	conn             *net.UDPConn
	options          *packetOptions
	shutdown         atomicBool
	mu               sync.Mutex
	concurrencyLimit chan struct{}
}

// newUDPPacketServer creates a Packet Server that is configured for UDP
func newUDPPacketServer(conn *net.UDPConn, options *packetOptions) PacketServer {
	return &udpPacketServerImpl{
		conn:             conn,
		options:          options,
		concurrencyLimit: make(chan struct{}, options.ConcurrencyLimit),
	}
}

//Port returns the port
// Can be used to find out the real port if helve is started with
// port 0 to auto bind
func (ps *udpPacketServerImpl) Port() (int, error) {
	if ps.conn == nil {
		return 0, ErrNilConn
	}
	// We made sure there's at least one UDP listener, and that one's
	// port was applied to all the others for the dynamic bind case.
	return ps.conn.LocalAddr().(*net.UDPAddr).Port, nil
}

// Serve starts a Packet server
func (ps *udpPacketServerImpl) Serve(handler PacketHandler) error {
	if ps.shutdown.isSet() {
		return ErrPacketServiceShutdown
	} else if handler == nil {
		return ErrNilPacketHander
	} else if ps.conn == nil {
		return ErrMissingPacketConn
	}
	// Continuously listen/process packets
	for {
		// If at concurrency limit do not try to read from connection yet
		ps.concurrencyLimit <- struct{}{}

		if ps.shutdown.isSet() {
			<-ps.concurrencyLimit
			return ErrPacketServiceShutdown
		}

		if ps.options.ReadDeadline > 0 {
			deadline := time.Now().Add(ps.options.ReadDeadline)
			if err := ps.conn.SetReadDeadline(deadline); err != nil {
				log.Println(err)
			}
		}
		buf := make([]byte, udpPacketBufSize)
		n, rAddr, err := ps.conn.ReadFrom(buf) // blocks until receive
		if err != nil {
			// log.Println("Error reading from udp socket", err)
			<-ps.concurrencyLimit

			if ps.shutdown.isSet() {
				return ErrPacketServiceShutdown
			}
			continue
		}

		ts := time.Now()
		// must be greater than zero to be considered a validate packet
		if n < 1 {
			// log.Println("Invalid packet received, packet size must be greater than zero")
			<-ps.concurrencyLimit
			continue
		}
		go func() {
			// Form a packet and handle through a registered handler function
			handler.HandlePacket(NewPacket(buf[:n], rAddr, ts), &hacketPacketWriter{ps.conn, ps.options})
			<-ps.concurrencyLimit
		}()
	}
}

// Shutdown will wait for read messages to be finished processing and
// sets shutdown so that new messages will not be read
// Can end early by closing context.
func (ps *udpPacketServerImpl) Shutdown(ctx context.Context) error {
	// Mark server as shutdown
	ps.shutdown.setTrue()

	// Close connection to stop reading new messages
	ps.conn.Close()

	// Wait for handlers to finish or context to be done
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ps.waitForHandlers():
		return nil
	}
}

// waitForHandlers pushes to the concurrency limit channel until
// full to ensure that no more handlers are processing
func (ps *udpPacketServerImpl) waitForHandlers() <-chan struct{} {
	handlersDone := make(chan struct{})
	go func() {
		for i := uint32(0); i < ps.options.ConcurrencyLimit; i++ {
			ps.concurrencyLimit <- struct{}{}
		}
		close(handlersDone)
	}()
	return handlersDone
}
