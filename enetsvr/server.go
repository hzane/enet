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
	c := install_signal()
	host, err := enet.NewHost(endpoint)
	panic_if_error(err)

	host.OnReliable = pong
	host.OnUnreliable = pong
	host.OnConnected = on_peer_connected
	host.OnDisconnected = on_peer_disconnected
	host.Run(c)
}

func on_peer_connected(peer enet.Peer, ret int) {

}
func on_peer_disconnected(peer enet.Peer) {

}
func install_signal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}

func pong(peer enet.Peer, chid uint8, data []uint8) {
	enet.SendReliable(host, peer, chid, data)
}

func panic_if_error(err error) {
	if err != nil {
		panic(err)
	}
}
