package hacket

import (
	"net"
	"sync"
	"time"
)

// PacketWriter interface used in handlers
type PacketWriter interface {
	WriteTo(msg PacketMessage, addr net.Addr) (int, error)
}

// hacketPacketWriter is a Wrapper around a packetConn. PacketWriter specifically
// wraps the Write side of a PacketConn only enforcing the PacketMessage param to
// writeTo. This is so we can force the user into using our PacketMessage and
// PacketMessageBuilder
type hacketPacketWriter struct {
	conn    net.PacketConn
	options *packetOptions
}

// WriteTo wraps the internal PacketConn WriteTo. PacketMessage is the payload of the network
// packet, Addr is the address of the remote peer we are sending to
func (hpw *hacketPacketWriter) WriteTo(msg PacketMessage, addr net.Addr) (int, error) {
	if hpw.options.WriteDeadline > 0 {
		deadline := time.Now().Add(hpw.options.WriteDeadline)
		if err := hpw.conn.SetWriteDeadline(deadline); err != nil {
			return 0, err
		}
	}
	return hpw.conn.WriteTo(msg, addr)
}

// PacketHandler defines a function to handle Packets
type PacketHandler interface {
	HandlePacket(packet Packet, pw PacketWriter)
}

// PacketHandlerFunc defines a function definition to handle packets
type PacketHandlerFunc func(packet Packet, pw PacketWriter)

// HandlePacket statifies the PacketHandler interface
func (s PacketHandlerFunc) HandlePacket(packet Packet, pw PacketWriter) {
	s(packet, pw)
}

type packetMuxEntry struct {
	packetHandler PacketHandler
	pktType       PacketType
}

//PacketMux allows PacketHandlers to be registered.
type PacketMux struct {
	mu sync.RWMutex
	m  map[PacketType]packetMuxEntry
}

// NewPacketMux initializes a PacketMux
func NewPacketMux() *PacketMux {
	return new(PacketMux)
}

// PacketHandler registers a PacketHandler with key PacketType
// PacketType acts as the route and PacketHandler is the function to be called
func (pmux *PacketMux) PacketHandler(pktType PacketType, packetHandler PacketHandler) error {
	pmux.mu.Lock()
	if packetHandler == nil {
		return ErrNilPacketHander
	}
	if pmux.m == nil {
		pmux.m = make(map[PacketType]packetMuxEntry)
	}
	if _, ok := pmux.m[pktType]; ok {
		return ErrPacketHandlerAlreadyExists
	}
	pmux.m[pktType] = packetMuxEntry{packetHandler: packetHandler, pktType: pktType}
	pmux.mu.Unlock()
	return nil
}

// PacketHandlerFunc registers a PacketHandlerFunc with key PacketType
// PacketType acts as the route and PacketHandler is the function to be called
func (pmux *PacketMux) PacketHandlerFunc(pktType PacketType, packetHandler func(packet Packet, pw PacketWriter)) error {
	pmux.mu.Lock()
	if packetHandler == nil {
		return ErrNilPacketHander
	}
	if pmux.m == nil {
		pmux.m = make(map[PacketType]packetMuxEntry)
	}
	if _, ok := pmux.m[pktType]; ok {
		return ErrPacketHandlerAlreadyExists
	}
	pmux.m[pktType] = packetMuxEntry{packetHandler: PacketHandlerFunc(packetHandler), pktType: pktType}
	pmux.mu.Unlock()
	return nil
}

// HandlePacket statifies the PacketHandler interface. A PacketType is expected to be
// set by the caller. The PacketType is used to find the PacketHandler to call
func (pmux *PacketMux) HandlePacket(packet Packet, pw PacketWriter) {
	pktType, msg, _ := decode(packet.Msg())
	// update packet msg with remove pktType
	packet.SetMsg(msg)

	handler := pmux.findPacketHandler(pktType)
	if handler != nil {
		handler.HandlePacket(packet, pw)
	}
}

// findPacketHandler returns a PacketHandler if the pktType is found
func (pmux *PacketMux) findPacketHandler(pktType PacketType) PacketHandler {
	pmux.mu.RLock()
	entry, ok := pmux.m[pktType]
	if !ok {
		return nil
	}
	pmux.mu.RUnlock()
	return entry.packetHandler
}
