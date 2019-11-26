package hacket

import (
	"net"
)

// New initalizes a packet server and a packet client
func New(network string, address string, options ...Options) (PacketServer, PacketClient, error) {
	// Setup the connection based on the network protocol
	switch network {
	case "udp":
		udpOptions := defaultPacketOption()
		// set all options if supplied
		for _, opt := range options {
			opt.apply(udpOptions)
		}
		udpAddr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, nil, err
		}
		conn, err := net.ListenUDP(network, udpAddr)
		if err != nil {
			return nil, nil, err
		}
		if udpOptions.ReadBufferSize > 0 {
			if err := conn.SetReadBuffer(udpOptions.ReadBufferSize); err != nil {
				return nil, nil, err
			}
		}
		if udpOptions.WriteBufferSize > 0 {
			if err := conn.SetWriteBuffer(udpOptions.WriteBufferSize); err != nil {
				return nil, nil, err
			}
		}
		udpServer := newUDPPacketServer(conn, udpOptions)
		udpClient := newUDPPacketClient(conn, udpOptions)
		return udpServer, udpClient, nil
	default:
		return nil, nil, ErrInvalidProtocol
	}
}
