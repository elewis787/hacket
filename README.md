# Hacket
Golang Packet Handlers for net.PacketConn 


#### Ping Example 
```go
package main

import (
	"context"
	"log"
	"net"

	"github.com/elewis787/hacket"
)

const (
	ping hacket.PacketType = iota
)

func main() {
	// Setup Ping Service
	option1 := []hacket.Options{}
	pingServer, _, err := hacket.New("udp", "127.0.0.1:1234", option1...)
	if err != nil {
		log.Fatal(err)
	}

	mux1 := hacket.NewPacketMux()
	mux1.PacketHandlerFunc(ping, func(packet hacket.Packet, pw hacket.PacketWriter) {
		msgBytes := packet.Msg()
		if string(msgBytes) != "ping" {
			return
		}
		log.Println("Service recieved ping from:", packet.FromAddr().String())
		// write pong back
		pongMsg, err := hacket.NewPacketMessageBuilder([]byte("pong")).Build()
		if err != nil {
			return
		}
		pw.WriteTo(pongMsg, packet.FromAddr())
	})
	go func() {
		if err := pingServer.Serve(mux1); err != nil {
			log.Fatal(err)
		}
	}()

	// Ping service and expect pong back ( using standard udp connection)

	// get address of to pingService
	remoteUDPAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}

	// "conncet" to pingService
	conn, err := net.DialUDP("udp", nil, remoteUDPAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Build a Packet message containing the "route" or PacketType  for ping handler
	pingMsg, err := hacket.NewPacketMessageBuilder([]byte("ping")).WithPacketType(ping).Build()
	if err != nil {
		log.Fatal(err)
	}

	// Send a packet
	if _, err := conn.Write(pingMsg); err != nil {
		log.Fatal(err)
	}

	// read from connnect
	buf := make([]byte, 4)
	if _, _, err := conn.ReadFromUDP(buf); err != nil {
		log.Fatal(err)
	}

	if string(buf) != "pong" {
		log.Fatal("did not receive pong message")
	} else {
		log.Println("Success !!!")
	}

	pingServer.Shutdown(context.Background())
}

```