package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/elewis787/hacket"
)

const (
	ping hacket.PacketType = iota
	pong
)

func main() {
	// Setup Ping Service
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	pongServer, pongClient, err := hacket.New("udp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}

	pongMux := hacket.NewPacketMux()

	pongMux.PacketHandlerFunc(pong, func(p hacket.Packet, pw hacket.PacketWriter) {
		body := p.Msg()
		switch string(body) {
		case "pong":
			log.Println("pong")
			response, err := hacket.NewPacketMessageBuilder([]byte("ping")).
				WithPacketType(ping).
				Build()
			if err != nil {
				log.Println(err)
			}
			pw.WriteTo(response, p.FromAddr())
		default:
			log.Println("no in pong")
			return
		}
	})

	go func() {
		log.Println("starting pong server")
		if err := pongServer.Serve(pongMux); err != nil {
			log.Fatal(err)
		}
	}()

	remoteUDPAddr, err := net.ResolveUDPAddr("udp", "localhost:9001")
	if err != nil {
		log.Fatal(err)
	}

	pingMsg, err := hacket.NewPacketMessageBuilder([]byte("ping")).
		WithPacketType(ping).
		Build()

	pingServer, _, err := hacket.New("udp", "localhost:9001")
	if err != nil {
		log.Fatal(err)
	}

	pingMux := hacket.NewPacketMux()

	pingMux.PacketHandlerFunc(ping, func(p hacket.Packet, pw hacket.PacketWriter) {
		body := p.Msg()
		switch string(body) {
		case "ping":
			log.Println("ping")
			response, err := hacket.NewPacketMessageBuilder([]byte("pong")).
				WithPacketType(pong).
				Build()
			if err != nil {
				log.Println(err)
			}
			pw.WriteTo(response, p.FromAddr())
		default:
			log.Println("no in ping")
			return
		}
	})

	go func() {
		log.Println("starting ping server")
		if err := pingServer.Serve(pingMux); err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := pongClient.WriteTo(pingMsg, remoteUDPAddr); err != nil {
		log.Println(err)
	}

	<-ctx.Done()
}
