package mocks

import (
	"io"
	"net"
	"time"
)

// MockPacketConn implements the PacketConn interface
// using a io Pipe to read and write
// the to and from address are ignored
// this is a WIP and only used for testing
type MockPacketConn struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

// NewMockPacketConn creates a mock packet connection
func NewMockPacketConn(reader *io.PipeReader, writer *io.PipeWriter) *MockPacketConn {
	return &MockPacketConn{
		reader: reader,
		writer: writer,
	}
}

// ReadFrom mocks the readfrom function on packet conn interface
func (m *MockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, err := m.reader.Read(b)
	return n, nil, err
}

// WriteTo mocks the writeto function on the packet conn interface
func (m *MockPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := m.writer.Write(b)
	return n, err
}

// Close mocks the close function on the packet conn interface
func (m *MockPacketConn) Close() error {
	m.writer.Close()
	m.reader.Close()
	return nil
}

// LocalAddr noop
func (m *MockPacketConn) LocalAddr() net.Addr { return nil }

// SetDeadline noop
func (m *MockPacketConn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline noop
func (m *MockPacketConn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline noop
func (m *MockPacketConn) SetWriteDeadline(t time.Time) error { return nil }
