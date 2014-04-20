package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/heartszhang/enet"
)

func main() {
	endpoint := "127.0.0.1:19091"
	c := signal_create()

	handlers := enet.HandlersCreate()
	handlers.OnConnected = ping_on_connect

	var err error
	var client enet.Host

	client, err = enet.ClientCreate(handlers)
	if err == nil {
		_, err = enet.ClientConnect(client, endpoint)
	}
	if err == nil {
		err = enet.HostRun(client, c)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func ping_on_connect(host enet.Host, peer enet.Peer, reason uint) {
	enet.SendReliable(host, peer, 0, []byte("hello enet"))
}
func ping_on_reliable(host enet.Host, peer enet.Peer, chid uint8, data []uint8) {
	s := string([]byte(data))

	fmt.Println(s)
}
func signal_create() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}
