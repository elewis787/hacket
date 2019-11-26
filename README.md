# Hacket
Golang Packet Handlers for net.PacketConn 

### Basic Exmaple 

```go
package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/elewis787/hacket"
)

func main() {
	_, client1, err := hacket.New("udp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}

	option2 := []hacket.Options{hacket.WithConcurrencyLimit(1)}
	server2, _, err := hacket.New("udp", "127.0.0.1:1235", option2...)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		h := &basicHandler{}
		if err := server2.Serve(h); err != nil {
			log.Println(err)
		}
	}()

	server2Addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1235")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		// write to node b
		msg, err := hacket.NewPacketMessageBuilder([]byte(fmt.Sprintf("message-%d", i))).Build()
		if err != nil {
			log.Fatal(err)
		}
		if _, err := client1.WriteTo(msg, server2Addr); err != nil {
			log.Fatal(err)
		}
	}
	wg.Wait()
}

type basicHandler struct {
	counter int32
}

func (h *basicHandler) HandlePacket(packet hacket.Packet, pw hacket.PacketWriter) {
	x := atomic.AddInt32(&h.counter, 1)
	log.Printf("got : %s with count %d\n", string(packet.Msg()), x)
}
```

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