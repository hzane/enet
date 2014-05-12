package enet

import "encoding/binary"

func enet_packet_ack_default(chanid uint8, sn uint32) (hdr enet_packet_header, ack enet_packet_ack) {
	hdr.cmd = enet_packet_type_ack
	hdr.flags = 0
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr) + binary.Size(ack))
	hdr.sn = sn
	return
}

func enet_packet_syn_default(sn uint32) (hdr enet_packet_header, syn enet_packet_syn) {
	syn.peerid = 0
	syn.mtu = enet_default_mtu
	syn.wnd_size = enet_default_wndsize
	syn.channel_n = enet_default_channel_count
	syn.rcv_bandwidth = 0
	syn.snd_bandwidth = 0
	syn.throttle_interval = enet_default_throttle_interval
	syn.throttle_acce = enet_default_throttle_acce
	syn.throttle_dece = enet_default_throttle_dece

	hdr.cmd = enet_packet_type_syn
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = enet_channel_id_none
	hdr.size = uint32(binary.Size(hdr) + binary.Size(syn))
	hdr.sn = sn
	return
}
func enet_packet_synack_default(sn uint32) (hdr enet_packet_header, syn enet_packet_syn) {
	hdr, syn = enet_packet_syn_default(sn)
	hdr.cmd = enet_packet_type_synack
	return
}

func enet_packet_fin_default(sn uint32) (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_fin
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = enet_channel_id_none
	hdr.size = uint32(binary.Size(hdr))
	hdr.sn = sn
	return
}

func enet_packet_ping_default(chanid uint8, sn uint32) (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_ping
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr))
	hdr.sn = sn
	return
}

func enet_packet_reliable_default(chanid uint8, sn uint32) (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_reliable
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr))
	hdr.sn = sn
	return
}

func enet_packet_unreliable_default(chanid uint8, usn uint32) (hdr enet_packet_header, pkt enet_packet_unreliable) {
	hdr.cmd = enet_packet_type_unreliable
	hdr.flags = 0
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr) + binary.Size(pkt))
	hdr.sn = usn
	pkt.usn = usn
	return
}

func enet_packet_fragment_default(chanid uint8, sn uint32) (hdr enet_packet_header, pkt enet_packet_fragment) {
	hdr.cmd = enet_packet_type_fragment
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr) + binary.Size(pkt))
	hdr.sn = sn
	return
}

/*
func host_handle_send(host *enet_host, cmd enet_command) bool {
	peer := host_peer_get(host, cmd.addr).(*enet_peer)

	if peer.state != peer_state_established {
		return false
	}
	return peer_handle_send(peer, host, cmd)
}
*/
