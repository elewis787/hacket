package hacket

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

// Run this test with --race to check for data race during shutdown
func TestShutdownRace(t *testing.T) {
	addr := "127.0.0.1:12341"
	//Default Concurrency limit is 1
	server, client, err := New("udp", addr)
	if err != nil {
		t.Fatal("Error creating new server and client:", err)
	}

	delayType := PacketType(1)

	mux := NewPacketMux()
	mux.PacketHandlerFunc(delayType, delayHandler)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Serve(mux)
		if err != nil {
			t.Log("Server ended with error:", err)
		}
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatal("Error resolving udp addr:", err)
	}

	// Write 2 messages to be handled in delayHandler
	// Due to concurrency limit of 1 the first message will be processing while the second waits before starting
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 2; i++ {
			msg, err := NewPacketMessageBuilder([]byte{5}).WithPacketType(delayType).Build()
			if err != nil {
				t.Fatal("Error building message:", err)
			}
			if _, err := client.WriteTo(msg, udpAddr); err != nil {
				t.Fatalf("Error during message %d WriteTo: $%v", i, err)
			}
		}
	}()

	// Sleep a bit to allow client to send the messages
	time.Sleep(time.Millisecond * 10)

	// Try to shutdown server
	err = server.Shutdown(context.TODO())
	if err != nil {
		t.Fatal("Error during shutdown:", err)
	}

	wg.Wait()
}

func TestShutdownContext(t *testing.T) {
	addr := "127.0.0.1:12342"
	//Default Concurrency limit is 1
	server, client, err := New("udp", addr)
	if err != nil {
		t.Fatal("Error creating new server and client:", err)
	}

	delayType := PacketType(1)

	mux := NewPacketMux()
	mux.PacketHandlerFunc(delayType, delayHandler)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Serve(mux)
		if err != nil {
			t.Log("Server ended with error:", err)
		}
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatal("Error resolving udp addr:", err)
	}

	// Write 2 messages to be handled in delayHandler
	// Due to concurrency limit of 1 the first message will be processing while the second waits before starting
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 2; i++ {
			msg, err := NewPacketMessageBuilder([]byte{5}).WithPacketType(delayType).Build()
			if err != nil {
				t.Fatal("Error building message:", err)
			}
			if _, err := client.WriteTo(msg, udpAddr); err != nil {
				t.Fatalf("Error during message %d WriteTo: $%v", i, err)
			}
		}
	}()

	// Sleep a bit to allow client to send the messages
	time.Sleep(time.Millisecond * 10)

	// Try to shutdown server
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context in 100 milliseconds
	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	// Shutdown
	err = server.Shutdown(ctx)
	if err != nil {
		if err != context.Canceled {
			t.Fatal("Expected context canceled error, receieved:", err)
		}
	} else {
		t.Fatal("Expected context cancel error, received none")
	}

	wg.Wait()
}

// Handler that takes 1 second to process message
func delayHandler(packet Packet, pw PacketWriter) {
	time.Sleep(time.Second)
}
