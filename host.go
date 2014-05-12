package enet

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"time"
)

type enet_host_incoming_command struct {
	protocol_header enet_protocol_header
	packet_header   enet_packet_header
	payload         []byte
	endpoint        *net.UDPAddr
}
type enet_host_outgoing_command struct {
	peer     Peer
	payload  []byte
	chanid   uint8
	reliable bool
}
type enet_host struct {
	fail              int // socket
	socket            *net.UDPConn
	local_addr        *net.UDPAddr
	incoming          chan *enet_host_incoming_command
	outgoing          chan *enet_host_outgoing_command
	tick              <-chan time.Time
	peers             map[string]*enet_peer
	timers            enet_timer_queue
	wnd_size          uint32 // enet_default_wndsize
	rcv_bandwidth     uint32
	snd_bandwidth     uint32
	throttle_interval uint32 // enet_default_throttle_interval
	throttle_acce     uint32 // enet_default_throttle_acce
	throttle_dece     uint32 // enet_default_throttle_dece
	next_clientid     uint32
	flags             int
	rcvd_bytes        int
	sent_bytes        int
	recv_bps          int
	send_bps          int
	bps_epoc          int64 // ms
	now               int64 // ms
}

func HostNew(addr string) (Host, error) {
	ep, err := net.ResolveUDPAddr("udp", addr)

	host := &enet_host{
		fail:              0,
		local_addr:        ep,
		incoming:          make(chan *enet_host_incoming_command),
		outgoing:          make(chan *enet_host_outgoing_command),
		tick:              time.Tick(time.Millisecond * enet_default_tick_ms),
		peers:             make(map[string]*enet_peer),
		wnd_size:          enet_default_wndsize,
		throttle_interval: enet_default_throttle_interval,
		throttle_acce:     enet_default_throttle_acce,
		throttle_dece:     enet_default_throttle_dece,
	}
	if err == nil {
		host.socket, err = net.ListenUDP("udp", ep)
	}
	if err != nil {
		host.flags |= enet_host_flags_stopped
	}

	return host, err
}

const (
	enet_host_flags_none = 1 << iota
	enet_host_flags_stopped
	enet_host_flags_sock_closed
	enet_host_flags_tick_closed
)

func (host *enet_host) Run(sigs chan os.Signal) {
	for host.flags&enet_host_flags_stopped == 0 {
		select {
		case item := <-host.incoming:
			host.now = unixtime_now()
			host.when_enet_host_incoming_command(item)
		case item := <-host.outgoing:
			host.now = unixtime_now()
			host.when_enet_outgoing_host_command(item)
		case sig := <-sigs:
			host.now = unixtime_now()
			signal.Stop(sigs)
			host.when_signal(sig)
		case t := <-host.tick:
			host.now = unixtime_now()
			host.when_tick(t)
		}
	}
}
func (host *enet_host) SetConnectionHandler(PeerEventHandler) {

}
func (host *enet_host) SetDisconnectionHandler(PeerEventHandler) {

}

func (host *enet_host) SetDataHandler(DataEventHandler) {

}

func (self *enet_host) Connect(ep string) {

}

func (host *enet_host) do_socket_receive() {
	buf := make([]byte, enet_udp_size) // large enough

	sock := host.socket
	for {
		n, addr, err := sock.ReadFromUDP(buf)
		if err != nil {
			break
		}
		dat := buf[:n]
		reader := bytes.NewReader(dat)
		var phdr enet_protocol_header
		binary.Read(reader, binary.BigEndian, &phdr)

		if phdr.flags&enet_protocol_flags_crc != 0 {
			var crc32 enet_crc32_header
			binary.Read(reader, binary.BigEndian, &crc32)
		}

		var pkhdr enet_packet_header
		for i := uint8(0); err == nil && i < phdr.packet_n; i++ {
			err = binary.Read(reader, binary.BigEndian, &pkhdr)
			pkhdr.size -= uint32(binary.Size(pkhdr))
			payload := make([]byte, pkhdr.size)
			n, err := reader.Read(payload)

			if err == nil {
				host.when_incoming_packet(phdr, pkhdr, payload, addr)
			}
		}

	}
	if host.flags&enet_host_flags_stopped == 0 {
		host.when_incoming_packet(enet_protocol_header{}, enet_packet_header{}, nil, nil)
	}
}

func (host *enet_host) when_signal(sig os.Signal) {
	host.close()
}

func (host *enet_host) close() {
	if host.flags&enet_host_flags_stopped != 0 {
		return
	}
	host.flags |= enet_host_flags_stopped
	// force close socket
	assert(host.socket != nil)
	host.flags |= enet_host_flags_sock_closed
	host.socket.Close()
	//		host.socket = nil

	// disable tick func
	host.flags |= enet_host_flags_tick_closed
}
func (host *enet_host) when_tick(t time.Time) {
	if host.flags&enet_host_flags_tick_closed != 0 {
		return
	}
	for cb := host.timers.pop(host.now); cb != nil; cb = host.timers.pop(host.now) {
		cb()
	}
}

func (host *enet_host) peer_from_endpoint_clientid(ep *net.UDPAddr, clientid uint32) *enet_peer {
	assert(ep != nil)
	id := ep.String()
	peer, ok := host.peers[id]
	if !ok {
		peer = new_enet_peer(ep, host)
		peer.clientid = clientid
		host.peers[id] = peer
	}
	return peer
}
func (host *enet_host) socket_send(data []byte, addr *net.UDPAddr) {
	assert(host.socket != nil)
	n, err := host.socket.WriteToUDP(data, addr)
	assert(n == len(data) || err != nil)
	if err != nil {
		host.close()
	}
}

func (host *enet_host) when_incoming_packet(phdr enet_protocol_header,
	pkhdr enet_packet_header, payload []byte, addr *net.UDPAddr) (err error) {
	host.incoming <- &enet_host_incoming_command{
		phdr,
		pkhdr,
		payload,
		addr,
	}
	return
}
func (host *enet_host) when_enet_outgoing_host_command(item *enet_host_outgoing_command) {
	peer := item.peer.(*enet_peer)
	ch := peer.channel_from_id(item.chanid)
	l := len(item.payload)
	frags := (uint32(l) + peer.mtu - 1) / peer.mtu
	firstsn := ch.next_sn
	if frags > 1 {
		for i := uint32(0); i < frags; i++ {
			sn := ch.next_sn
			ch.next_sn++
			dat := item.payload[i*peer.mtu : (i+1)*peer.mtu]
			pkhdr, frag := enet_packet_fragment_default(item.chanid, sn)
			frag.count = frags
			frag.index = i
			frag.length = uint32(l)
			frag.offset = i * peer.mtu
			frag.sn = firstsn
			to := host.timers.push(host.now+peer.rtt_timeo, func() {})
			wi := &enet_channel_item{pkhdr, frag, dat, 0, 0, to}
			ch.outgoing_trans(wi)
		}

	} else {
		pkhdr := enet_packet_header{}
		to := host.timers.push(host.now+peer.rtt_timeo, func() {})
		wi := &enet_channel_item{pkhdr, enet_packet_fragment{}, item.payload, 0, 0, to}
		ch.outgoing_trans(wi)
	}
	ch.do_send(peer)
	return
}

type when_enet_packet_incoming_disp func(peer *enet_peer, hdr enet_packet_header, payload []byte)

var _when_enet_packet_incoming_disp = []when_enet_packet_incoming_disp{
	(*enet_peer).when_enet_incoming_ack,
	(*enet_peer).when_enet_incoming_syn,
	(*enet_peer).when_enet_incoming_synack,
	(*enet_peer).when_enet_incoming_fin,
	(*enet_peer).when_enet_incoming_ping,
	(*enet_peer).when_enet_incoming_reliable,
	(*enet_peer).when_enet_incoming_unrelialbe,
	(*enet_peer).when_enet_incoming_fragment,
	(*enet_peer).when_unknown,
	(*enet_peer).when_unknown,
	(*enet_peer).when_unknown,
	(*enet_peer).when_enet_incoming_eg,
	(*enet_peer).when_unknown,
}

func (host *enet_host) when_enet_host_incoming_command(item *enet_host_incoming_command) {
	if item == nil || item.payload == nil {
		host.close()
		return
	}
	if item.packet_header.cmd > enet_packet_type_count {

		return
	}
	peer := host.peer_from_endpoint_clientid(item.endpoint, item.protocol_header.clientid)
	ch := peer.channel_from_id(item.packet_header.chanid)

	// ack if needed
	if item.packet_header.flags&enet_packet_header_flags_needack != 0 {
		sn := ch.next_sn
		ch.next_sn++
		hdr, ack := enet_packet_ack_default(item.packet_header.chanid, sn)
		ack.sn = item.packet_header.sn
		ack.tm = item.protocol_header.time
		item := &enet_channel_item{hdr, enet_packet_fragment{}, nil, 0, 0, nil}
		ch.outgoing_trans(item)
		ch.outgoing_ack(hdr.sn)
	}
	_when_enet_packet_incoming_disp[item.packet_header.cmd](peer, item.packet_header, item.payload)
	ch.do_send(peer)
}
func (host *enet_host) destroy_peer(peer *enet_peer) {

}
