package enet

import (
	"bytes"
	"encoding/binary"
)

func host_handle_send(host *enet_host, cmd enet_command) bool {
	peer := host_peer_get(host, cmd.addr).(*enet_peer)

	if peer.state != peer_state_established {
		return false
	}
	return peer_handle_send(peer, host, cmd)
}

func peer_handle_send(peer *enet_peer, host *enet_host, cmd enet_command) bool {
	cmd.pid = peer.rid
	cmd.snt_time = uint16(host.now)
	if cmd.is_unseq != 0 {
		pkt := enet_packet_unsequence{uint16(cmd.cid), uint16(len(cmd.data))}
		writer := new(bytes.Buffer)
		rawhdr := enet_packet_header_raw_encode(cmd.enet_packet_header)
		binary.Write(writer, binary.BigEndian, &rawhdr)
		binary.Write(writer, binary.BigEndian, &pkt)
		writer.Write(cmd.data)
		host_socket_send(host, writer.Bytes(), cmd.addr)
		return true
	}
	if cmd.need_ack == 0 {
		pkt := enet_packet_unreliable{cmd.sn, uint16(len(cmd.data))}
		writer := new(bytes.Buffer)
		rawhdr := enet_packet_header_raw_encode(cmd.enet_packet_header)
		binary.Write(writer, binary.BigEndian, &rawhdr)
		binary.Write(writer, binary.BigEndian, &pkt)
		writer.Write(cmd.data)
		host_socket_send(host, writer.Bytes(), cmd.addr)
		return true
	}
	if peer.wnd_used >= int(enet_wnd_count) {
		return false
	}
	idx := cmd.sn % enet_wnd_count
	assure(peer.wnd[idx] == nil, "peer wnd occupied %v", idx)
	peer.wnd[idx] = &cmd
	peer.wnd_used++
	host_timeo_push(host, &cmd)
	peer_window_send(peer, host)
	return true
}

func peer_window_send(peer *enet_peer, host *enet_host) {
	if peer_window_is_full(peer) || peer_window_is_empty(peer) {
		return
	}
	peer.last_intrans++
	idx := peer.last_intrans % enet_wnd_count
	cmd := peer.wnd[idx]
	assure(cmd != nil, "invalid send-cmd")
	cmd.timeo = unixtime_now()
	writer := new(bytes.Buffer)
	rawhdr := enet_packet_header_raw_encode(cmd.enet_packet_header)
	binary.Write(writer, binary.BigEndian, &rawhdr)
	binary.Write(writer, binary.BigEndian, &enet_packet_reliable{uint16(len(cmd.data))})
	writer.Write(cmd.data)
	data := writer.Bytes()
	peer.intrans_bytes += len(data)
	host_socket_send(host, data, peer.raddr)
}
func peer_do_ack(peer *enet_peer, host *enet_host, chid uint8, sn, snttime uint16) {
	if peer.ack_used >= int(enet_wnd_count) {
		return // ack queue overflow
	}
	hdr := enet_packet_header{
		pid:      peer.rid,
		sid:      peer.outgoing_sessid,
		snt_time: unixtime_nowui16(),
		cmd:      enet_cmd_ack,
		cid:      chid,
	}
	ack := enet_packet_ack{sn, snttime}
	writer := new(bytes.Buffer)
	binary.Write(writer, binary.BigEndian, &ack)

	idx := sn % enet_wnd_count
	peer.acks[idx] = &enet_command{hdr, writer.Bytes(), peer.raddr, -1, 0}
	peer.ack_used++
	peer_ack_send(peer, host)
}

func peer_ack_send(peer *enet_peer, host *enet_host) {
	if peer.ack_used <= 0 || peer_snd_bandwidth_is_full(peer) {
		return
	}
	idx := peer.first_ack % enet_wnd_count
	peer.first_ack++
	cmd := peer.acks[idx]
	peer.acks[idx] = nil

	writer := new(bytes.Buffer)
	hdrraw := enet_packet_header_raw_encode(cmd.enet_packet_header)
	binary.Write(writer, binary.BigEndian, &hdrraw)
	writer.Write(cmd.data)
	host_socket_send(host, writer.Bytes(), cmd.addr)
}
