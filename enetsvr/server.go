package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/heartszhang/enet"
)

func main() {
	endpoint := "127.0.0.1:19091"
	c := install_signal()
	host, err := enet.NewHost(endpoint)
	panic_if_error(err)

	host.SetDataHandler(pong)
	host.SetConnectionHandler(on_peer_connected)
	host.SetDisconnectionHandler(on_peer_disconnected)
	host.Run(c)
}

func on_peer_connected(host enet.Host, peer string, ret int) {
	fmt.Printf("%v conn\n", peer)
}
func on_peer_disconnected(host enet.Host, peer string, ret int) {
	fmt.Printf("%v disconn\n", peer)
}
func install_signal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}

func pong(host enet.Host, ep string, chanid uint8, payload []byte) {
	host.Write(ep, chanid, payload)

	fmt.Printf("dat pong %v\n", ep)
}

func panic_if_error(err error) {
	if err != nil {
		panic(err)
	}
}
