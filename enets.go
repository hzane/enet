package enet

import (
	"net"
	"os"
	"time"
)

type Host interface {
	//	peer(*net.UDPAddr) Peer
	//	handle_recv(enet_command) bool
	//	handle_send(enet_command) bool
	//	handle_tick(time.Time) bool
	//	handle_break(os.Signal) bool

	//	socket_send([]byte, *net.UDPAddr) error
}

type Peer interface {
	connect(Host) error
	disconnect(Host) error
	//	handle_recv(Host, enet_packet_header, enet_reader)
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
	if laddr == nil {
		host.laddr = host.socket.LocalAddr().(*net.UDPAddr)
	}
	return host, err
}
func ClientConnect(host Host, endpoint string) (Peer, error) {
	var err error
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr("udp", endpoint)
	peer := host_peer_get(host.(*enet_host), addr)
	if err == nil {
		err = peer.connect(host)
	}
	return peer, err
}

func PeerDisconnect(host Host, peer Peer) error {
	return peer.disconnect(host)
}
func SendReliable(host Host, peer Peer, chid uint8, data []byte) error {
	cmd := enet_command_reliable_new(chid, data, peer.(*enet_peer))
	h := host.(*enet_host)
	h.outgoing <- cmd
	return nil
}

/*
func SendReliable(host Host, peer Peer, chid uint8, data []byte) {
	peer := p.(*enet_peer)
	if peer.state != peer_state_established {
		return
	}
	pktm := net_reliable_cmd_new(host, peer, chid, data)
	peer.append_reliable(pktm)
	peer.try_send(host)
}

func SendUnreliable(host Host, p Peer, chid uint8, data []byte) {
	peer := p.(*enet_peer)
	if peer.state != peer_state_established {
		return
	}
	pktm := peer.unreliable_packet_new(sn, chid, data)
	peer.append_unreliable(pktm)
	peer.try_send(host)
}

func SendUnsequenced(host Host, p Peer, groupid uint8, data []byte) {
	peer := p.(*enet_peer)
	if peer.state != peer_state_established {
		return
	}
	pktm := peer.unsequenced_cmd_new(groupid, data)
	peer.append_unsequenced(pktm)
	peer.try_send(host)
}
*/

func HostRun(host Host, c chan os.Signal) error {
	var (
		ok   = true
		h    = host.(*enet_host)
		in   = h.incoming
		out  = h.outgoing
		tick = h.tick
		sig  os.Signal
		cmd  enet_command
		now  time.Time
	)
	for ok {
		select {
		case cmd, ok = <-in:
			if ok {
				ok = host_handle_recv(h, cmd)
			}
		case cmd, ok = <-out:
			if ok {
				ok = host_handle_send(h, cmd)
			}
		case now, ok = <-tick:
			if ok {
				ok = host_handle_tick(h, now)
			}
		case sig, ok = <-c:
			host_handle_break(h, sig)
			ok = false
		}
	}
	return nil
}
