package enet

import (
	"bytes"
	"encoding/binary"
	"net"
)

func enet_socket_recv(in chan<- enet_command, sock *net.UDPConn) {
	var err error
	var n int
	var addr *net.UDPAddr
	var hdr enet_packet_header
	buf := make([]byte, 65536)
	for err == nil {
		n, addr, err = sock.ReadFromUDP(buf)
		var reader enet_reader
		if n > 0 {
			reader = enet_reader_new(buf[:n])
			hdr, err = enet_packet_header_decode(reader)
		}
		if err == nil {
			cmd := enet_command{hdr, reader.bytes(), addr, -1, 0}
			in <- cmd
		}
	}
}

func host_handle_recv(host *enet_host, cmd enet_command) bool {
	host.now = unixtime_now()
	peer := host_peer_get(host, cmd.addr).(*enet_peer)
	peer_handle_recv(peer, host, cmd)

	host.rcvd_bytes += len(cmd.data)
	return true
}

type enet_command_handler func(peer *enet_peer, host *enet_host, cmd enet_command)

var enet_command_handlers = [...]enet_command_handler{
	peer_handle_recv_none,
	peer_handle_recv_ack,
	peer_handle_recv_syn,
	peer_handle_recv_synack,
	peer_handle_recv_fin,
	peer_handle_recv_ping,
	peer_handle_recv_reliable,
	peer_handle_recv_unreliable,
	peer_handle_recv_fragment,
	peer_handle_recv_unsequenced,
	peer_handle_recv_bandwidth_limit,
	peer_handle_recv_throttle,
	peer_handle_recv_unreliable_fragment,
	peer_handle_recv_count,
	peer_handle_recv_none,
	peer_handle_recv_none,
}

func peer_handle_recv(peer *enet_peer, host *enet_host, cmd enet_command) {
	handler := enet_command_handlers[cmd.cmd]
	handler(peer, host, cmd)
}
func peer_handle_recv_none(peer *enet_peer, host *enet_host, cmd enet_command) {
	print("invalid command rcvd %v from %v", cmd.cmd, cmd.addr)
}
func peer_handle_recv_ack(peer *enet_peer, host *enet_host, cmd enet_command) {
	reader := enet_reader_new(cmd.data)
	ack, err := enet_packet_ack_decode(reader)
	if err != nil {
		return
	}
	v0 := peer.state == peer_state_established
	v1 := peer.state == peer_state_syn_rcvd  // send syn-ack's ack
	v2 := peer.state == peer_state_time_wait // wait nothing
	v3 := peer.state == peer_state_fin_sent  // wait fin_ack
	if !v0 || !v1 || !v2 || !v3 {
		return
	}
	idx := ack.sn % enet_wnd_count
	cmdx := peer.wnd[idx]
	if cmdx == nil {
		return
	}
	if cmdx.cmd == enet_cmd_synack && v1 { // synack's ack
		peer.state = peer_state_established
		peer_notify_connected(peer, host)
	}
	peer.wnd_used--
	peer.wnd[idx] = nil
	host_timeo_remove(host, cmdx.heap_idx)

	sntt := unixtime_fromui16(ack.time, host.now)
	rtt := host.now - sntt
	peer_throttle_update(peer, host, rtt)
	peer_rtt_update(peer, host, rtt)
}
func peer_handle_recv_fin(peer *enet_peer, host *enet_host, cmd enet_command) {
	s0 := peer.state == peer_state_established
	s1 := peer.state == peer_state_fin_rcvd
	//	s3 := peer.state == peer_state_time_wait
	//	s2 := peer.state == peer_state_fin_sent
	if !s0 || !s1 {
		return
	}
	if cmd.need_ack != 0 {
		peer.state = peer_state_fin_rcvd
		peer_do_ack(peer, host, cmd.cid, cmd.sn, cmd.snt_time)
	} else {
		peer.state = peer_state_time_wait
	}
	peer_notify_disconnected(peer, host)
}
func peer_handle_recv_ping(peer *enet_peer, host *enet_host, cmd enet_command) {
	s0 := peer.state == peer_state_established
	if !s0 {
		return
	}
	if cmd.need_ack != 0 {
		peer_do_ack(peer, host, cmd.cid, cmd.sn, cmd.snt_time)
	}
}
func peer_handle_recv_reliable(peer *enet_peer, host *enet_host, cmd enet_command) {
	s0 := peer.state == peer_state_established
	if !s0 {
		return
	}
	ch := peer_channel_get(peer, cmd.cid)

	if ch.expected_sn != cmd.sn { // some packets were lost
		return // discard this packet
	}
	reader := enet_reader_new(cmd.data)
	length := reader.uint16()
	data := reader.bytes()
	assure(int(length) == len(data), "corrupted packet")

	ch.expected_sn++
	peer_notify_data(peer, host, cmd.cid, data)
	if cmd.need_ack != 0 {
		peer_do_ack(peer, host, cmd.cid, cmd.sn, cmd.snt_time)
	}
}
func peer_handle_recv_fragment(peer *enet_peer, host *enet_host, cmd enet_command) {
	print("not implemented %v from %v", cmd.cmd, cmd.addr)
}
func peer_handle_recv_bandwidth_limit(peer *enet_peer, host *enet_host, cmd enet_command) {
	reader := enet_reader_new(cmd.data)
	bl, err := enet_packet_bandwidth_limit_decode(reader)
	if err != nil {
		return
	}
	peer.rcv_bandwidth = bl.rcv_bandwidth
	peer.snd_bandwidth = bl.snd_bandwidth
	if cmd.need_ack != 0 {
		peer_do_ack(peer, host, cmd.cid, cmd.sn, cmd.snt_time)
	}
}

func peer_handle_recv_throttle(peer *enet_peer, host *enet_host, cmd enet_command) {
	reader := enet_reader_new(cmd.data)
	tc, err := enet_packet_throttle_decode(reader)
	if err != nil {
		return
	}
	peer.throttle_i = int64(tc.interval)
	peer.throttle_acce = tc.acceleration
	peer.throttle_dece = tc.deceleration
	assure(cmd.need_ack != 0, "throttle-configure must be acked")
	if cmd.need_ack != 0 {
		peer_do_ack(peer, host, cmd.cid, cmd.sn, cmd.snt_time)
	}
}
func peer_handle_recv_unreliable_fragment(peer *enet_peer, host *enet_host, cmd enet_command) {
	print("not implemented %v from %v", cmd.cmd, cmd.addr)
}
func peer_handle_recv_unreliable(peer *enet_peer, host *enet_host, cmd enet_command) {
	print("not implemented %v from %v", cmd.cmd, cmd.addr)
}
func peer_handle_recv_unsequenced(peer *enet_peer, host *enet_host, cmd enet_command) {
	print("not implemented %v from %v", cmd.cmd, cmd.addr)
}
func peer_handle_recv_count(peer *enet_peer, host *enet_host, cmd enet_command) {
	print("invalid command rcvd %v from %v", cmd.cmd, cmd.addr)
}

func peer_handle_recv_syn(peer *enet_peer, host *enet_host, cmd enet_command) {
	s0 := peer.state == peer_state_closed
	s1 := peer.state == peer_state_none
	s2 := peer.state == peer_state_nothing
	if !s0 && !s1 && !s2 {
		print("syn abandoned %v", peer.state)
		return
	}
	reader := enet_reader_new(cmd.data)
	syn, err := enet_packet_syn_decode(reader)
	if err != nil {
		return
	}
	peer.rid = syn.pid
	peer.mtu = dampui32(syn.mtu, enet_mtu_min, enet_mtu_max)
	peer.rcv_bandwidth = syn.rcv_bandwidth
	peer.snd_bandwidth = syn.snd_bandwidth
	peer.wnd_size = dampui32(syn.wnd_size, enet_wnd_size_min, enet_wnd_size_max)
	peer.chan_count = dampui32(minui32(syn.ccount, peer.chan_count), enet_channel_count_min, enet_channel_count_min)
	peer.throttle_i = int64(syn.throttle_i)
	peer.throttle_acce = syn.throttle_acce
	peer.throttle_dece = syn.throttle_dece
	peer.conn_id = syn.conn_id
	peer.rdata = uint(syn.data)
	peer.state = peer_state_syn_rcvd

	// send -ack
	cmd.pid = syn.pid
	cmd.sid = peer.outgoing_sessid
	cmd.snt_time = unixtime_nowui16()
	cmd.cmd = enet_cmd_synack
	cmd.need_ack = enet_cmd_flag_ack

	syn.pid = peer.lid
	syn.in_sid = uint8(peer.outgoing_sessid)
	syn.out_sid = uint8(peer.incoming_sessid)
	syn.mtu = peer.mtu
	syn.wnd_size = host.wnd_size
	syn.ccount = peer.chan_count
	syn.rcv_bandwidth = host.rcv_bandwidth
	syn.snd_bandwidth = host.snd_bandwidth
	syn.throttle_i = host.throttle_i
	syn.throttle_acce = host.throttle_acce
	syn.throttle_dece = host.throttle_dece
	syn.conn_id = peer.conn_id
	syn.data = uint32(peer.ldata)

	writer := new(bytes.Buffer)

	binary.Write(writer, binary.BigEndian, &syn)
	cmd.data = writer.Bytes()
	cmd.addr = peer.raddr
	peer_handle_send(peer, host, cmd)
}
func peer_handle_recv_synack(peer *enet_peer, host *enet_host, cmd enet_command) {
	s0 := peer.state == peer_state_syn_sent
	if !s0 {
		return
	}
	reader := enet_reader_new(cmd.data)
	syn, err := enet_packet_syn_decode(reader)
	if err != nil {
		return
	}
	idx := peer.last_intrans % enet_wnd_count // must be syn
	cmdx := peer.wnd[idx]
	assure(cmdx != nil && cmdx.cmd == enet_cmd_syn, "invalid state transition")
	peer.wnd_used--
	peer.wnd[idx] = nil
	//host_pop_timeo(cmdx)
	host_timeo_remove(host, cmdx.heap_idx)

	peer.rid = syn.pid
	peer.mtu = dampui32(syn.mtu, enet_mtu_min, enet_mtu_max)
	peer.rcv_bandwidth = syn.rcv_bandwidth
	peer.snd_bandwidth = syn.snd_bandwidth
	peer.wnd_size = dampui32(syn.wnd_size, enet_wnd_size_min, enet_wnd_size_max)
	peer.chan_count = dampui32(minui32(syn.ccount, peer.chan_count), enet_channel_count_min, enet_channel_count_min)
	peer.throttle_i = int64(syn.throttle_i)
	peer.throttle_acce = syn.throttle_acce
	peer.throttle_dece = syn.throttle_dece
	peer.conn_id = syn.conn_id
	peer.rdata = uint(syn.data)
	peer.state = peer_state_established

	if cmd.need_ack != 0 {
		peer_do_ack(peer, host, 0xff, cmd.sn, cmd.snt_time)
	}
	peer_notify_connected(peer, host)
}

func peer_notify_connected(peer *enet_peer, host *enet_host) {

}

func peer_notify_disconnected(peer *enet_peer, host *enet_host) {

}
func peer_notify_data(peer *enet_peer, host *enet_host, cid uint8, data []byte) {

}
