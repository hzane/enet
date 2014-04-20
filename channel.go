package enet

import "net"

type enet_channel struct {
}

type enet_packet struct {
	addr *net.UDPAddr
	data []byte
}
