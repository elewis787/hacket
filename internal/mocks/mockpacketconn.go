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

func NewMockPacketConn(reader *io.PipeReader, writer *io.PipeWriter) *MockPacketConn {
	return &MockPacketConn{
		reader: reader,
		writer: writer,
	}
}

func (m *MockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, err := m.reader.Read(b)
	return n, nil, err
}

func (m *MockPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := m.writer.Write(b)
	return n, err
}

func (m *MockPacketConn) Close() error {
	m.writer.Close()
	m.reader.Close()
	return nil
}

func (m *MockPacketConn) LocalAddr() net.Addr                { return nil }
func (m *MockPacketConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockPacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockPacketConn) SetWriteDeadline(t time.Time) error { return nil }
