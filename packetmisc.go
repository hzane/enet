package enet

import (
	"bytes"
	"encoding/binary"
)

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_ack_default(chanid uint8) (hdr enet_packet_header, ack enet_packet_ack) {
	hdr.cmd = enet_packet_type_ack
	hdr.flags = 0
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr) + binary.Size(ack))
	return
}
func enet_packet_ack_encode(ack enet_packet_ack) []byte {
	writer := bytes.NewBuffer(nil)
	binary.Write(writer, binary.BigEndian, &ack)
	return writer.Bytes()
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_syn_default() (hdr enet_packet_header, syn enet_packet_syn) {
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
	return
}
func enet_packet_syn_encode(syn enet_packet_syn) []byte {
	writer := bytes.NewBuffer(nil)
	binary.Write(writer, binary.BigEndian, &syn)
	return writer.Bytes()
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_synack_default() (hdr enet_packet_header, sak enet_packet_synack) {
	hdr, syn := enet_packet_syn_default()
	hdr.cmd = enet_packet_type_synack
	sak = enet_packet_synack(syn)
	return
}
func enet_packet_synack_encode(sak enet_packet_synack) []byte {
	writer := bytes.NewBuffer(nil)
	binary.Write(writer, binary.BigEndian, &sak)
	return writer.Bytes()
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_fin_default() (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_fin
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = enet_channel_id_none
	hdr.size = uint32(binary.Size(hdr))
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_ping_default(chanid uint8) (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_ping
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr))
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_reliable_default(chanid uint8, payloadlen uint32) (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_reliable
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr)) + payloadlen
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_unreliable_default(chanid uint8, payloadlen, usn uint32) (hdr enet_packet_header, pkt enet_packet_unreliable) {
	hdr.cmd = enet_packet_type_unreliable
	hdr.flags = 0
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr)+binary.Size(pkt)) + payloadlen
	pkt.usn = usn
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_fragment_default(chanid uint8, fraglen uint32) (hdr enet_packet_header, pkt enet_packet_fragment) {
	hdr.cmd = enet_packet_type_fragment
	hdr.flags = enet_packet_header_flags_needack
	hdr.chanid = chanid
	hdr.size = uint32(binary.Size(hdr)+binary.Size(pkt)) + fraglen
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_eg_default() (hdr enet_packet_header) {
	hdr.cmd = enet_packet_type_fragment
	hdr.flags = enet_packet_header_flags_needack // should be acked
	hdr.chanid = enet_channel_id_none
	hdr.size = uint32(binary.Size(hdr))
	return
}
