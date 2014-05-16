package enet

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"time"
)

type enet_host struct {
	fail           int // socket
	socket         *net.UDPConn
	addr           *net.UDPAddr
	incoming       chan *enet_host_incoming_command
	outgoing       chan *enet_host_outgoing_command
	tick           <-chan time.Time
	peers          map[string]*enet_peer
	timers         enet_timer_queue
	next_clientid  uint32 // positive client id seed
	flags          int    // enet_host_flags_xxx
	rcvd_bytes     int
	sent_bytes     int
	rcvd_bps       int
	sent_bps       int
	update_epoc    int64
	now            int64 // ms
	last_recv_time int64
	last_send_time int64

	notify_connected    PeerEventHandler
	notify_disconnected PeerEventHandler
	notify_data         DataEventHandler
}

func NewHost(addr string) (Host, error) {
	// if failed, host will bind to a random address
	ep, err := net.ResolveUDPAddr("udp", addr)

	host := &enet_host{
		fail:     0,
		addr:     ep,
		incoming: make(chan *enet_host_incoming_command, 16),
		outgoing: make(chan *enet_host_outgoing_command, 16),
		tick:     time.Tick(time.Millisecond * enet_default_tick_ms),
		peers:    make(map[string]*enet_peer),
		timers:   new_enet_timer_queue(),
	}
	if err == nil {
		host.socket, err = net.ListenUDP("udp", ep)
	}
	if err != nil {
		host.flags |= enet_host_flags_stopped
	}
	if host.addr == nil && host.socket != nil {
		host.addr = host.socket.LocalAddr().(*net.UDPAddr)
	}
	debugf("host bind %v\n", ep)
	return host, err
}

// process:
// - incoming packets
// - outgoing data
// - exit signal
// - timer tick
func (host *enet_host) Run(sigs chan os.Signal) {
	host.flags |= enet_host_flags_running
	go host.run_socket()
	debugf("running...\n")
	for host.flags&enet_host_flags_stopped == 0 {
		select {
		case item := <-host.incoming:
			host.now = unixtime_now()
			host.when_incoming_host_command(item)
		case item := <-host.outgoing:
			host.now = unixtime_now()
			host.when_outgoing_host_command(item)
		case sig := <-sigs:
			host.now = unixtime_now()
			signal.Stop(sigs)
			host.when_signal(sig)
		case t := <-host.tick:
			host.now = unixtime_now()
			host.when_tick(t)
		}
	}
	debugf("%v run exits\n", host.addr)
	host.flags &= ^enet_host_flags_running
}

func (host *enet_host) Connect(ep string) {
	host.outgoing <- &enet_host_outgoing_command{ep, nil, enet_channel_id_all, true}
}
func (host *enet_host) Disconnect(ep string) {
	host.outgoing <- &enet_host_outgoing_command{ep, nil, enet_channel_id_none, true}
}

func (host *enet_host) Write(endp string, chanid uint8, dat []byte) {
	host.outgoing <- &enet_host_outgoing_command{endp, dat, chanid, true}
}
func (host *enet_host) Stop() {
	host.when_socket_incoming_packet(enet_protocol_header{}, enet_packet_header{}, nil, nil)
}

// run in another routine
func (host *enet_host) run_socket() {
	buf := make([]byte, enet_udp_size) // large enough

	sock := host.socket
	for {
		n, addr, err := sock.ReadFromUDP(buf)
		// syscall.EINVAL
		if err != nil { // any err will make host stop run
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
			payload := make([]byte, int(pkhdr.size)-binary.Size(pkhdr))
			_, err := reader.Read(payload)
			debugf("socket recv %v\n", pkhdr)
			if err == nil {
				host.when_socket_incoming_packet(phdr, pkhdr, payload, addr)
			}
		}

	}
	// socket may be not closed yet
	if host.flags&enet_host_flags_stopped == 0 {
		host.when_socket_incoming_packet(enet_protocol_header{}, enet_packet_header{}, nil, nil)
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

	assert(host.socket != nil)
	if host.flags&enet_host_flags_sock_closed == 0 {
		host.flags |= enet_host_flags_sock_closed
		host.socket.Close()
	}

	// disable tick func
	host.flags |= enet_host_flags_tick_closed
}
func (host *enet_host) when_tick(t time.Time) {
	if host.flags&enet_host_flags_tick_closed != 0 {
		return
	}
	host.update_statis()
	for cb := host.timers.pop(host.now); cb != nil; cb = host.timers.pop(host.now) {
		cb()
	}
}

// push data to socket
func (host *enet_host) do_send(dat []byte, addr *net.UDPAddr) {
	assert(host.socket != nil)
	host.update_snt_statis(len(dat))
	n, err := host.socket.WriteToUDP(dat, addr)
	assert(n == len(dat) || err != nil)
	if err != nil {
		host.close()
	}
}

// move rcvd socket datagrams to run routine
// payload or addr is nil means socket recv breaks
func (host *enet_host) when_socket_incoming_packet(phdr enet_protocol_header,
	pkhdr enet_packet_header,
	payload []byte,
	addr *net.UDPAddr) (err error) {
	host.incoming <- &enet_host_incoming_command{
		phdr,
		pkhdr,
		payload,
		addr,
	}
	return
}
func (host *enet_host) connect_peer(ep string) {
	cid := host.next_clientid
	host.next_clientid++
	peer := host.peer_from_endpoint(ep, cid)
	if peer.clientid != cid { // connect a established peer?
		notify_peer_connected(peer, enet_peer_connect_result_duplicated)
		return
	}
	hdr, syn := enet_packet_syn_default()
	ch := peer.channel_from_id(enet_channel_id_none)
	//	ch.outgoing_pend(hdr, enet_packet_fragment{}, enet_packet_syn_encode(syn), nil)
	peer.outgoing_pend(ch, hdr, enet_packet_fragment{}, enet_packet_syn_encode(syn))
}
func (host *enet_host) disconnect_peer(ep string) {
	peer := host.peer_from_endpoint(ep, enet_peer_id_any)
	if peer.flags&enet_peer_flags_established == 0 {
		notify_peer_disconnected(peer, enet_peer_disconnect_result_invalid)
		return
	}
	if peer.flags&(enet_peer_flags_fin_rcvd|enet_peer_flags_fin_sending) != 0 {
		return
	}
}
func (host *enet_host) reset_peer(ep string) {
	peer := host.peer_from_endpoint(ep, enet_peer_id_any)
	host.destroy_peer(peer)
}
func (host *enet_host) when_outgoing_host_command(item *enet_host_outgoing_command) {
	if item.payload == nil {
		if item.chanid == enet_channel_id_all { // connect request
			host.connect_peer(item.peer)
		}
		if item.chanid == enet_channel_id_none { // disconnect
			host.disconnect_peer(item.peer)
		}
		return
	}
	peer := host.peer_from_endpoint(item.peer, enet_peer_id_any)
	if peer.flags&enet_peer_flags_established == 0 ||
		peer.flags&(enet_peer_flags_fin_sending|enet_peer_flags_synack_sending) != 0 {
		return
	}
	ch := peer.channel_from_id(item.chanid)
	l := uint32(len(item.payload))
	frags := (l + peer.mtu - 1) / peer.mtu
	firstsn := ch._next_sn
	if frags > 1 {
		for i := uint32(0); i < frags; i++ {
			dat := item.payload[i*peer.mtu : (i+1)*peer.mtu]
			pkhdr, frag := enet_packet_fragment_default(item.chanid, uint32(len(dat)))
			frag.count = frags
			frag.index = i
			frag.length = l
			frag.offset = i * peer.mtu
			frag.sn = firstsn
			peer.outgoing_pend(ch, pkhdr, frag, dat)
		}

	} else {
		pkhdr := enet_packet_reliable_default(item.chanid, l)
		peer.outgoing_pend(ch, pkhdr, enet_packet_fragment{}, item.payload)
	}
	ch.do_send(peer)
	return
}

func (host *enet_host) when_incoming_host_command(item *enet_host_incoming_command) {
	if item == nil || item.payload == nil {
		host.close()
		return
	}
	host.update_rcv_statis(int(item.packet_header.size))

	if item.packet_header.cmd > enet_packet_type_count {
		// invalid packet type, nothing should be done
		debugf("skipped packet: %v\n", item.packet_header.cmd)
		return
	}
	peer := host.peer_from_addr(item.endpoint, item.protocol_header.clientid)
	if peer.clientid != item.protocol_header.clientid {
		debugf("cid mismatch %v\n", peer.remote_addr)
		return
	}
	ch := peer.channel_from_id(item.packet_header.chanid)

	// ack if needed
	if item.packet_header.flags&enet_packet_header_flags_needack != 0 {
		hdr, ack := enet_packet_ack_default(item.packet_header.chanid)
		ack.sn = item.packet_header.sn
		ack.tm = item.protocol_header.time
		peer.outgoing_pend(ch, hdr, enet_packet_fragment{}, enet_packet_ack_encode(ack))
		//		i := &enet_channel_item{hdr, enet_packet_fragment{}, enet_packet_ack_encode(ack), 0, 0, nil}
		//		ch.outgoing_pend(i)
		//		ch.outgoing_ack(hdr.sn) // ack needn't ack, so we just mark it as acked
	}
	_when_enet_packet_incoming_disp[item.packet_header.cmd](peer, item.packet_header, item.payload)
	ch.do_send(peer)
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

func (host *enet_host) destroy_peer(peer *enet_peer) {
	id := peer.remote_addr.String()
	delete(host.peers, id)
	debugf("release peer %v\n", id)
}
func (host *enet_host) SetConnectionHandler(h PeerEventHandler) {
	host.notify_connected = h
}
func (host *enet_host) SetDisconnectionHandler(h PeerEventHandler) {
	host.notify_disconnected = h
}

func (host *enet_host) SetDataHandler(h DataEventHandler) {
	host.notify_data = h
}
func (host *enet_host) update_rcv_statis(rcvd int) {
	host.rcvd_bytes += rcvd
	host.last_recv_time = host.now
}
func (host *enet_host) update_snt_statis(snt int) {
	host.sent_bytes += snt
	host.last_send_time = host.now
}

func (host *enet_host) update_statis() {
	itv := int(host.now - host.update_epoc)
	host.rcvd_bps = host.rcvd_bytes * 1000 / itv
	host.sent_bps = host.sent_bytes * 1000 / itv
	host.rcvd_bytes = 0
	host.sent_bytes = 0
	for _, peer := range host.peers {
		peer.update_statis(itv)
	}
}

const (
	enet_host_flags_none = 1 << iota
	enet_host_flags_stopped
	enet_host_flags_running
	enet_host_flags_sock_closed
	enet_host_flags_tick_closed
)

type enet_host_incoming_command struct {
	protocol_header enet_protocol_header
	packet_header   enet_packet_header // .size == len(payload)
	payload         []byte
	endpoint        *net.UDPAddr
}
type enet_host_outgoing_command struct {
	peer     string
	payload  []byte
	chanid   uint8
	reliable bool
}

func (host *enet_host) peer_from_endpoint(ep string, clientid uint32) *enet_peer {
	addr, _ := net.ResolveUDPAddr("udp", ep)
	return host.peer_from_addr(addr, clientid)
}
func (host *enet_host) peer_from_addr(ep *net.UDPAddr, clientid uint32) *enet_peer {
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
