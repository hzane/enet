package enet

import (
	"container/heap"
	"net"
	"os"
	"time"
)

type enet_host struct {
	total_rcvd_bytes int
	rcvd_bytes       int
	socket           *net.UDPConn
	laddr            *net.UDPAddr
	handlers         Handlers
	incoming         chan enet_packet
	outgoing         chan enet_packet
	tick             chan time.Time
	peers            map[string]*enet_peer
	timers           *timers
}

func enet_host_new() *enet_host {
	return &enet_host{}
}

func (self *enet_host) peer(addr *net.UDPAddr) Peer {
	peer, ok := self.peers[addr.String()]
	if !ok {
		peer = enet_peer_new()
	}
	return peer
}
func (host *enet_host) handle_send(packet enet_packet) bool {
	_, err := host.socket.WriteToUDP(packet.data, packet.addr)
	return err == nil
}
func (host *enet_host) handle_recv(packet enet_packet) bool {
	peer := host.peer(packet.addr).(*enet_peer)
	reader := enet_reader_new(packet.data)
	//hdraw, err :=enet_packet_header_raw_decode(reader)
	hdr, err := enet_packet_header_decode(reader)
	if err == nil || hdr.pid == peer.id {
		peer.handle_recv(host, hdr, reader)
	}
	host.rcvd_bytes += len(packet.data)
	return true
}

func (host *enet_host) send_to(data []byte, addr *net.UDPAddr) error {
	packet := enet_packet{addr, data}
	host.outgoing <- packet
	return nil
	//	_, err := host.socket.WriteToUDP(data, addr)
}

// false: break service
// use this function to do some cleanup
func (host *enet_host) handle_break(os.Signal) bool {
	print("host cleanup...here")
	return false
}

func (host *enet_host) handle_tick(time.Time) bool {
	for timer := host.timeout(); timer != nil; {
		timer.run(host)
		timer = host.timeout()
	}
	return true
}

func (host *enet_host) timeout() (t timer) {
	if host.timers.Len() == 0 {
		return
	}
	now := time.Now().UnixNano()
	top := (*host.timers)[0]
	if top.weight < now {
		t = heap.Pop(host.timers).(timer)
	}
	return
}
func enet_host_connect() (peer *enet_peer, err error) {
	return nil, enet_err_not_implemented
}

func enet_host_disconnect(peer *enet_peer) (err error) {
	return nil
}

func enet_peer_on_connect(peer *enet_peer) {
}

func enet_peer_on_disconnect(peer *enet_peer) {
}

func enet_peer_on_reliable(peer *enet_peer, data []uint8) {
}

func enet_peer_on_unreliable(peer *enet_peer, data []uint8) {
}

func enet_peer_on_unsequence(peer *enet_peer, data []uint8) {
}

func enet_peer_send_reliable() {
}
func enet_peer_send_unreliable() {
}
func enet_peer_send_unsequence() {
}
func enet_peer_disconnect() {
}

func enet_host_create() (host *enet_host, err error) {
	return nil, nil
}
func enet_host_destroy(host *enet_host) (err error) {
	return nil
}

var enet_cmd_handlers = make(map[uint8]func(*enet_host, *enet_peer, enet_packet_header, enet_reader))

func init() {

}

func enet_host_peer_via_address(host *enet_host, addr enet_address) (peer *enet_peer) {
	return &enet_peer{}
}

func enet_reader_create(data []uint8) enet_reader {
	return nil
}
