package hacket

import (
	"net"
	"time"
)

// PacketClient defines a packet client interface
type PacketClient interface {
	WriteTo(msg PacketMessage, addr net.Addr) (int, error)
}

var _ PacketClient = &udpPacketClientImpl{}

// udpPacketClientImpl implements the packet client interface
type udpPacketClientImpl struct {
	conn    net.PacketConn
	options *packetOptions
}

// NewUDPPacketClient creates a new UDP packet client
func newUDPPacketClient(conn *net.UDPConn, options *packetOptions) *udpPacketClientImpl {
	return &udpPacketClientImpl{
		conn:    conn,
		options: options,
	}
}

// WriteTo writes a packet to the target destination
func (pc *udpPacketClientImpl) WriteTo(msg PacketMessage, address net.Addr) (int, error) {
	if pc.options.WriteDeadline > 0 {
		deadline := time.Now().Add(pc.options.WriteDeadline)
		if err := pc.conn.SetWriteDeadline(deadline); err != nil {
			return 0, err
		}
	}
	// Interal packet write function
	return pc.conn.WriteTo(msg, address)
}
