package enet

import (
	"net"
	"os"
	"time"
)

type Host interface {
	peer(*net.UDPAddr) Peer
	handle_recv(enet_packet) bool
	handle_send(enet_packet) bool
	handle_tick(time.Time) bool
	handle_break(os.Signal) bool
	send_to([]byte, *net.UDPAddr) error
}
type Peer interface {
	connect(Host) error
	disconnect(Host) error
	handle_recv(Host, enet_packet_header, enet_reader)
}

type EnetPeerHandler func(Host, Peer, uint)
type EnetReceivedHandler func(Host, Peer, uint8, []uint8)

type Handlers struct {
	OnConnected    EnetPeerHandler
	OnDisconnected EnetPeerHandler
	OnReliable     EnetReceivedHandler
	OnUnreliable   EnetReceivedHandler
	OnUnrequenced  EnetReceivedHandler
}

func HandlersCreate() Handlers {
	return Handlers{
		enet_on_connected,
		enet_on_disconnected,
		enet_on_reliable,
		enet_on_unreliable,
		enet_on_unsequenced,
	}
}

func HostNew(handlers Handlers, laddr *net.UDPAddr) (Host, error) {
	var err error
	host := enet_host_new()
	host.handlers = handlers
	host.laddr = laddr
	host.socket, err = net.ListenUDP("udp", laddr)
	return host, err
}
func ClientConnect(host Host, endpoint string) (Peer, error) {
	var err error
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr("udp", endpoint)
	peer := host.peer(addr)
	if err == nil {
		err = peer.connect(host)
	}
	return peer, err
}

func PeerDisconnect(host Host, peer Peer) error {
	return peer.disconnect(host)
}

func SendReliable(Host, Peer, uint8, []byte) {

}

func HostRun(host Host, c chan os.Signal) error {
	var (
		ok     = true
		sig    os.Signal
		h      = host.(*enet_host)
		in     = h.incoming
		out    = h.outgoing
		tick   = h.tick
		packet enet_packet
		now    time.Time
	)
	for ok {
		select {
		case packet, ok = <-in:
			if ok {
				ok = host.handle_recv(packet)
			}
		case packet, ok = <-out:
			if ok {
				ok = host.handle_send(packet)
			}
		case now, ok = <-tick:
			if ok {
				ok = host.handle_tick(now)
			}
		case sig, ok = <-c:
			host.handle_break(sig)
			ok = false
		}
	}
	return nil
}
