package hacket

import (
	"bytes"
	"net"
	"time"
)

const (
	// udpPacketBufSize is used to buffer incoming packets during read
	// operations.
	udpPacketBufSize = 65535 // (8 byte UDP Header - 20 Byte IP Header = 65507)
)

// PacketMessage is a custom byte slice used to create a network packet
type PacketMessage []byte

// PacketType defines the type of message to expect as the network packets payload
// Additonally, PacketType is used by the PacketMux during handler selection
type PacketType uint8

// PacketMessageBuilder provides a function chain to create either a raw
// PacketMessage without a PacketType or to create a PacketMessage with a
// prepended PacketType. This is used my the PacketMux when selecting
// a packet Handler.
type PacketMessageBuilder struct {
	m       PacketMessage
	pktType *PacketType
}

// NewPacketMessageBuilder initalizes a new PacketMessageBuilder with b
// If b is nil the PacketMessage Builder will throw an error during the build call.
func NewPacketMessageBuilder(b []byte) *PacketMessageBuilder {
	return &PacketMessageBuilder{m: b}
}

// Build creates a PacketMessage with the bytes supplied when the PacketMessageBuilder was created
// If a PacketType has been set, it will be prepended to the PacketMessage
func (mb *PacketMessageBuilder) Build() (PacketMessage, error) {
	if mb.m == nil {
		return nil, ErrNilByteSlice
	}
	if len(mb.m) > udpPacketBufSize {
		return nil, ErrMaxMessageSize
	}
	if mb.pktType != nil {
		encodedMsg, err := encode(*mb.pktType, mb.m)
		if err != nil {
			return nil, err
		}
		return encodedMsg, nil
	}
	return mb.m, nil
}

// WithPacketType sets a PacketType to be prepended to a PacketMessage
func (mb *PacketMessageBuilder) WithPacketType(pktType PacketType) *PacketMessageBuilder {
	mb.pktType = &pktType
	return mb
}

// encode prepends a PacketType to a PacketMessage
func encode(pktType PacketType, b []byte) (PacketMessage, error) {
	if b == nil {
		return nil, ErrNilByteSlice
	}
	buf := bytes.NewBuffer(nil)
	if err := buf.WriteByte(uint8(pktType)); err != nil {
		return nil, err
	}
	n, err := buf.Write(b)
	if n != len(b) {
		return nil, ErrByteWrite
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decode will return the first byte of the PacketMessage. It is the responsibility of the caller
// to ensure that the first byte is a valid packet type.PacketMessage will contain the remainder
// of the byte slice
func decode(b []byte) (PacketType, PacketMessage, error) {
	pktType := PacketType(b[0])
	return pktType, b[1:], nil
}

// Packet is used to wrap incoming network packets from peers
// over a packet connection. Addtionally it provides Meta data about
// the packet. Message is the network packets payload, FromAddr is the
// address of the remote peer, and Timestamp is local timestamp of when the
// packet was created/recieved
type Packet struct {
	msg       PacketMessage
	fromAddr  net.Addr
	timestamp time.Time
}

// NewPacket returns a new packet
func NewPacket(msg PacketMessage, fromAddr net.Addr, timestamp time.Time) Packet {
	return Packet{
		msg:       msg,
		fromAddr:  fromAddr,
		timestamp: timestamp,
	}
}

// Msg returns the message in the packet
func (p *Packet) Msg() PacketMessage {
	return p.msg
}

// SetMsg sets the message of the packet
func (p *Packet) SetMsg(msg PacketMessage) {
	p.msg = msg
}

// FromAddr returns the address of the remote peer
func (p *Packet) FromAddr() net.Addr {
	return p.fromAddr
}

// Timestamp is the local timestamp when packet was created
func (p *Packet) Timestamp() time.Time {
	return p.timestamp
}
