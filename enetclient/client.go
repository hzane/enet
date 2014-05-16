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

	host, err := enet.NewHost("")
	panic_if_error(err)
	host.SetDataHandler(ping_on_reliable)
	host.SetConnectionHandler(ping_on_connect)
	host.SetDisconnectionHandler(func(enet.Host, string, int) {
		host.Stop()
	})
	host.Connect(endpoint)
	host.Run(c)
}

func ping_on_connect(host enet.Host, ep string, reason int) {
	if reason == 0 {
		host.Write(ep, 0, []byte("hello enet"))
	}
}
func ping_on_reliable(host enet.Host, ep string, chid uint8, data []byte) {
	s := string([]byte(data))

	fmt.Println(s)
	host.Disconnect(ep)
}
func install_signal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	return c
}

func panic_if_error(err error) {
	if err != nil {
		panic(err)
	}
}
