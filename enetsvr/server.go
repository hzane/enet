package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/heartszhang/enet"
)

func main() {
	endpoint := "127.0.0.1:19091"
	c := signal_create()
	var host enet.Host
	var err error
	var addr *net.UDPAddr
	handlers := enet.HandlersCreate()
	handlers.OnReliable = pong_on_reliable

	addr, err = net.ResolveUDPAddr("udp", endpoint)
	if err == nil {
		host, err = enet.HostNew(handlers, addr)
	}
	if err == nil {
		enet.HostRun(host, c)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func signal_create() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}

func pong_on_reliable(host enet.Host, peer enet.Peer, chid uint8, data []uint8) {
	enet.SendReliable(host, peer, chid, data)
}
