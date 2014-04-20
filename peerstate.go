package enet

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	peer_state_none             int = iota
	peer_state_closed               //  disconnected          closed-tcp
	peer_state_syn_sent             // connecting            sync-sent
	peer_state_syn_rcvd             // acking-connect        sync-rcvd
	peer_state_listening            // connection-pending    listening //connection-succeeded
	peer_state_established          //             established
	peer_state_disconnect_later     //      closing(fin_sent+fin_rcvd+fin_ack_sent)
	peer_state_fin_sent             // disconnecting         fin_wait_1(fin_sent)
	peer_state_time_wait            // zombie                time_wait
	//	peer_state_fin_ack_rcvd          // acking-disconnected   fint_wait_2(fin_ack-rcvd)
	peer_state_nothing
)

type enet_peer_state_handler func(*enet_peer, *enet_host, enet_packet_header, enet_reader)

var enet_peer_state_handlers = []enet_peer_state_handler{
	enet_peer_state_none_handle,
	enet_peer_state_closed_handle,
	enet_peer_state_syn_sent_handle,
	enet_peer_state_syn_rcvd_handle,
	enet_peer_state_listening_handle,
	enet_peer_state_established_handle,
	enet_peer_state_fin_sent_handle,
	enet_peer_state_time_wait_handle,
	enet_peer_state_nothing_handle,
}

//var enet_peer_state_handlers = make(map[uint]enet_peer_state_handler)

func init() {
}
func enet_peer_state_handler_get(state int) enet_peer_state_handler {
	cnt := len(enet_peer_state_handlers)
	if state < 0 || state >= cnt {
		return enet_peer_state_nothing_handle
	}
	return enet_peer_state_handlers[state]
}

func enet_peer_state_none_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	if hdr.cmd != enet_cmd_syn {
		return
	}
	isreliable := hdr.is_unseq == 0
	ndack := hdr.need_ack != 0
	nocid := hdr.cid == 0xff
	if !isreliable || !ndack || !nocid {
		return
	}
	peer.handle_syn(host, hdr, reader)
}

func (peer *enet_peer) handle_syn(host *enet_host, hdr enet_packet_header, reader enet_reader) {
	syn, err := enet_packet_syn_decode(reader)
	if err != nil {
		return
	}
	//	assert(peer.pid == hdr.pid, "peer-id mismatch in conn")
	peer.in_sid = 0  // ignore syn's out_sid
	peer.out_sid = 0 // ignore syn's in_sid
	peer.rid = syn.pid
	peer.mtu = dampui32(syn.mtu, enet_mtu_min, enet_mtu_max)
	peer.rcv_bandwidth = syn.rcv_bandwidth
	peer.snd_bandwidth = syn.snd_bandwidth
	peer.wnd_size = dampui32(syn.wnd_size, enet_wnd_size_min, enet_wnd_size_max)
	peer.chcount = dampui32(minui32(syn.ccount, peer.chcount), enet_channel_count_min, enet_channel_count_min)
	peer.throttle_i = syn.throttle_i
	peer.throttle_acce = syn.throttle_acce
	peer.throttle_dece = syn.throttle_dece
	peer.cid = syn.conn_id
	peer.data = syn.data
	peer.state = peer_state_syn_rcvd

	hdr.pid = peer.rid
	hdr.cmd = enet_cmd_synack
	hdr.sid = peer.out_sid
	hdr.snt_time = unixtimeui16(time.Now())
	hdr.sn = peer.out_sn_inc()
	hdrr := enet_packet_header_raw_encode(hdr)

	syn.ccount = peer.chcount
	syn.in_sid = uint8(peer.out_sid)
	syn.out_sid = uint8(peer.in_sid)
	syn.mtu = peer.mtu
	syn.wnd_size = peer.wnd_size
	syn.throttle_i = peer.throttle_i
	syn.throttle_acce = peer.throttle_acce
	syn.throttle_dece = peer.throttle_dece
	syn.conn_id = peer.cid
	syn.pid = peer.id

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, &hdrr)
	binary.Write(buf, binary.BigEndian, &syn)
	host.send_to(buf.Bytes(), peer.raddr)
	// ack isn't needed
}
func enet_peer_state_closed_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	enet_peer_state_none_handle(peer, host, hdr, reader)
}

func (peer *enet_peer) handle_synack(host *enet_host, hdr enet_packet_header, reader enet_reader) {
	synack, err := enet_packet_synack_decode(reader)
	if err != nil {
		return
	}
	// makesure
	peer.chcount = dampui32(minui32(peer.chcount, synack.ccount), enet_channel_count_min, enet_channel_count_max)
	td := (peer.throttle_dece == synack.throttle_dece)
	ti := (peer.throttle_i == synack.throttle_i)
	ta := (peer.throttle_acce == synack.throttle_acce)
	cia := (peer.cid == synack.conn_id)
	assert(td && ti && ta && cia, "invalid handshake param")
	if !cia {
		peer.state = peer_state_closed
	}
	peer.out_sid = uint16(synack.out_sid)
	peer.in_sid = uint16(synack.in_sid)
	peer.rid = synack.pid
	peer.mtu = dampui32(synack.mtu, enet_mtu_min, enet_mtu_max)
	peer.wnd_size = dampui32(synack.wnd_size, enet_wnd_size_min, enet_wnd_size_max)
	peer.rcv_bandwidth = synack.rcv_bandwidth
	peer.snd_bandwidth = synack.snd_bandwidth

	peer.send_ack(host, hdr.sn)
	peer.state = peer_state_established
	// notify connected
}

// syn+ack=>established
func enet_peer_state_syn_sent_handle(peer *enet_peer,
	host *enet_host,
	hdr enet_packet_header,
	reader enet_reader) {
	if hdr.cmd != enet_cmd_synack {
		return
	}
	peer.handle_synack(host, hdr, reader)
}

// syn's ack => established
func enet_peer_state_syn_rcvd_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	if hdr.cmd != enet_cmd_synack {
		return
	}
	peer.handle_synack(host, hdr, reader)
}
func enet_peer_state_listening_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	enet_peer_state_none_handle(peer, host, hdr, reader)
}
func enet_peer_state_established_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	switch hdr.cmd {
	case enet_cmd_fin:
		peer.handle_fin(host, hdr, reader)
	case enet_cmd_ping:
	case enet_cmd_reliable:
	case enet_cmd_unreliable:
	case enet_cmd_unsequenced:
	case enet_cmd_bandwidthlimit:
	case enet_cmd_throttle:
	case enet_cmd_unreliable_fragment:
	case enet_cmd_syn, enet_cmd_synack:
		// do fail
	default:
		// do skip
	}
}
func (peer *enet_peer) handle_fin(host *enet_host, hdr enet_packet_header, reader enet_reader) {
	peer.state = peer_state_time_wait
	peer.send_ack(host, hdr.sn)
}

func enet_peer_state_fin_sent_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	switch hdr.cmd {
	case enet_cmd_ack:
		peer.state = peer_state_closed
	case enet_cmd_fin:
		peer.handle_fin(host, hdr, reader)
	}
}

// wait for fin's ack
func enet_peer_state_time_wait_handle(peer *enet_peer, host *enet_host, hdr enet_packet_header, reader enet_reader) {
	if hdr.cmd != enet_cmd_ack {
		return
	}
	peer.state = peer_state_closed
}
func enet_peer_state_nothing_handle(*enet_peer, *enet_host, enet_packet_header, enet_reader) {
}
