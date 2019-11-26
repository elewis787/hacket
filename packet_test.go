package hacket

import (
	"bytes"
	"testing"
)

// testaPacketHandler - dummy function used for lookup tests
func testPacketHandler(packet *Packet, pw *PacketWriter) error {
	// dummy
	return nil
}

func TestMessageBuilderWithPacketType(t *testing.T) {
	testcases := []struct {
		name    string
		payload []byte
		msgType PacketType
		want    PacketMessage
	}{
		{
			name:    "test-messagebuilder-1",
			payload: []byte("test1"),
			msgType: PacketType(uint8(1)),
			want:    PacketMessage(append([]byte{uint8(1)}, []byte("test1")...)),
		},
		{
			name:    "test-messagebuilder-2",
			payload: []byte("test2"),
			msgType: PacketType(uint8(2)),
			want:    PacketMessage(append([]byte{uint8(2)}, []byte("test2")...)),
		},
		{
			name:    "test-messagebuilder-3",
			payload: []byte("!@#$%^&*()1234567890qwertyuiopasdfghjklzxcvbnm{}[]:;?"),
			msgType: PacketType(uint8(3)),
			want:    PacketMessage(append([]byte{uint8(3)}, []byte("!@#$%^&*()1234567890qwertyuiopasdfghjklzxcvbnm{}[]:;?")...)),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewPacketMessageBuilder(tt.payload).WithPacketType(tt.msgType).Build()
			if err != nil {
				t.Error(err)
			}
			if value := bytes.Compare(msg, tt.want); value != 0 {
				t.Error("Byte slices do not match")
			}
		})
	}
}

func TestMessageBuilderNoPacketType(t *testing.T) {
	testcases := []struct {
		name    string
		payload []byte
		want    PacketMessage
	}{
		{
			name:    "test-messagebuilder-1",
			payload: []byte("test1"),
			want:    PacketMessage([]byte("test1")),
		},
		{
			name:    "test-messagebuilder-2",
			payload: []byte("test2"),
			want:    PacketMessage([]byte("test2")),
		},
		{
			name:    "test-messagebuilder-3",
			payload: []byte("!@#$%^&*()1234567890qwertyuiopasdfghjklzxcvbnm{}[]:;?"),
			want:    PacketMessage([]byte("!@#$%^&*()1234567890qwertyuiopasdfghjklzxcvbnm{}[]:;?")),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewPacketMessageBuilder(tt.payload).Build()
			if err != nil {
				t.Error(err)
			}
			if value := bytes.Compare(msg, tt.want); value != 0 {
				t.Error("Byte slices do not match")
			}
		})
	}
}

func TestMessageBuilderNil(t *testing.T) {
	_, err := NewPacketMessageBuilder(nil).Build()
	if err == nil {
		t.Error(err)
	}
}
