package enet

import (
	"container/heap"
	"net"
	"os"
	"time"
)

type enet_host struct {
	total_rcvd_bytes int
	socket           *net.UDPConn
	laddr            *net.UDPAddr
	handlers         Handlers
	incoming         chan enet_command
	outgoing         chan enet_command
	tick             chan time.Time
	peers            map[string]*enet_peer
	timers           *timers
	wnd_size         uint32
	rcv_bandwidth    uint32
	snd_bandwidth    uint32
	throttle_i       uint32
	throttle_acce    uint32
	throttle_dece    uint32
	rcvd_bytes       int
	sent_bytes       int
	recv_bps         int
	send_bps         int
	bps_epoc         int64
	now              int64
}

func enet_host_new() *enet_host {
	return &enet_host{
		wnd_size:      enet_wnd_size_default,
		throttle_i:    enet_throttle_default,
		throttle_acce: enet_throttle_acce_default,
		throttle_dece: enet_throttle_dece_default,
	}
}

func host_peer_get(host *enet_host, addr net.Addr) Peer {
	id := addr.String()
	peer, ok := host.peers[id]
	if !ok {
		peer = enet_peer_new(addr)
		host.peers[id] = peer
	}
	return peer
}
func host_socket_send(host *enet_host, data []byte, addr *net.UDPAddr) error {
	_, err := host.socket.WriteToUDP(data, addr)
	return err
}

// false: break service
// use this function to do some cleanup
func host_handle_break(host *enet_host, sig os.Signal) bool {
	print("host cleanup...here")
	return false
}

func host_handle_tick(host *enet_host, now time.Time) bool {
	host.now = unixtime_now()
	for timer := host_first_timeo(host); timer != nil; {
		command_timeo_run(host, timer)
		timer = host_first_timeo(host)
	}
	return true
}

func host_first_timeo(host *enet_host) (t *enet_command) {
	if host.timers.Len() == 0 {
		return
	}

	top := (*host.timers)[0]
	if top.timeo < host.now {
		t = heap.Pop(host.timers).(*enet_command)
	}
	return
}
func host_timeo_remove(host *enet_host, idx int) {
	if idx >= host.timers.Len() {
		return
	}
	heap.Remove(host.timers, idx)
}
func host_timeo_push(host *enet_host, cmd *enet_command) {
	heap.Push(host.timers, cmd)
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

func command_timeo_run(host *enet_host, cmd *enet_command) {

}
