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
	wg               sync.WaitGroup
}

// newUDPPacketServer creates a Packet Server that is configured for UDP
func newUDPPacketServer(conn *net.UDPConn, options *packetOptions) PacketServer {
	return &udpPacketServerImpl{
		conn:             conn,
		options:          options,
		concurrencyLimit: make(chan struct{}, options.ConcurrencyLimit),
	}
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

		// Try to incriment Wait Group, returns false if shutdown was already called
		if !ps.incrimentWaitGroup() {
			<-ps.concurrencyLimit
			return ErrPacketServiceShutdown
		}

		if ps.options.ReadDealine > 0 {
			deadline := time.Now().Add(ps.options.ReadDealine)
			if err := ps.conn.SetReadDeadline(deadline); err != nil {
				log.Println(err)
			}
		}
		buf := make([]byte, udpPacketBufSize)
		n, rAddr, err := ps.conn.ReadFrom(buf) // blocks until receive
		if err != nil {
			// log.Println("Error reading from udp socket", err)
			<-ps.concurrencyLimit
			ps.wg.Done()

			if ps.shutdown.isSet() {
				ps.conn.Close()
				return ErrPacketServiceShutdown
			}

			continue
		}

		ts := time.Now()
		// must be greater than zero to be considered a validate packet
		if n < 1 {
			// log.Println("Invalid packet received, packet size must be greater than zero")
			<-ps.concurrencyLimit
			ps.wg.Done()
			continue
		}
		go func() {
			defer ps.wg.Done()
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
	ps.mu.Lock()
	ps.shutdown.setTrue()
	ps.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ps.waitForHandlers():
		return nil
	}
}

// incrimentWaitGroup increments the waitgroup after checking that shutdown has not yet been called
// If shutdown is set return false without incrimenting
func (ps *udpPacketServerImpl) incrimentWaitGroup() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.shutdown.isSet() {
		return false
	}

	ps.wg.Add(1)
	return true
}

// waitForHandlers calls Wait on the waitgroup and closes returned channel when finished
func (ps *udpPacketServerImpl) waitForHandlers() <-chan struct{} {
	wgDone := make(chan struct{})

	go func() {
		ps.wg.Wait()
		close(wgDone)
	}()

	return wgDone
}
