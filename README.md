# Hacket
Golang Packet Handlers for net.PacketConn 

[![GoDoc](https://godoc.org/github.com/elewis787/hacket?status.svg)](https://godoc.org/github.com/elewis787/hacket)
[![Go Report Card](https://goreportcard.com/badge/github.com/elewis787/hacket)](https://goreportcard.com/report/github.com/elewis787/hacket)

### TODO 
- background 
- cleanup of shutdown 
- cleanup docs 
- extend examples 
- add additional implementations

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
	options := []hacket.Options{}
	// initalize udp server, no need for the client 
	hacketServer, _, err := hacket.New("udp", "127.0.0.1:1234", options...)
	if err != nil {
		log.Fatal(err)
	}
	
	// mux used for registering packet handlers 
	mux := hacket.NewPacketMux()
	
	// inline packet handler 
	mux.PacketHandlerFunc(ping, func(packet hacket.Packet, pw hacket.PacketWriter) {
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
	
	// start hacket server 
	go func() {
		if err := hacketServer.Serve(mux); err != nil {
			log.Fatal(err)
		}
	}()
	
	// Ping service and expect pong back ( using standard udp connection)

	// get address of to hacketServer
	remoteUDPAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}

	// "conncet" to hacketServer
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
	// check to see if the contents match what we expect from the packet handler 
	if string(buf) != "pong" {
		log.Fatal("did not receive pong message")
	} else {
		log.Println("Success !!!")
	}
	// shut down the server 
	hacketServer.Shutdown(context.Background())
}

```
