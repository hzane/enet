package enet

import (
	"bytes"
	"encoding/binary"
)

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_ack_default(chanid uint8) (hdr EnetPacketHeader, ack EnetPacketAck) {
	hdr.Type = enet_packet_type_ack
	hdr.Flags = 0
	hdr.ChannelID = chanid
	hdr.Size = uint32(binary.Size(hdr) + binary.Size(ack))
	return
}
func enet_packet_ack_encode(ack EnetPacketAck) []byte {
	writer := bytes.NewBuffer(nil)
	binary.Write(writer, binary.BigEndian, &ack)
	return writer.Bytes()
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_syn_default() (hdr EnetPacketHeader, syn EnetPacketSyn) {
	syn.PeerID = 0
	syn.MTU = enet_default_mtu
	syn.WndSize = enet_default_wndsize
	syn.ChannelCount = enet_default_channel_count
	syn.RcvBandwidth = 0
	syn.SndBandwidth = 0
	syn.ThrottleInterval = enet_default_throttle_interval
	syn.ThrottleAcce = enet_default_throttle_acce
	syn.ThrottleDece = enet_default_throttle_dece

	hdr.Type = enet_packet_type_syn
	hdr.Flags = enet_packet_header_flags_needack
	hdr.ChannelID = enet_channel_id_none
	hdr.Size = uint32(binary.Size(hdr) + binary.Size(syn))
	return
}
func enet_packet_syn_encode(syn EnetPacketSyn) []byte {
	writer := bytes.NewBuffer(nil)
	binary.Write(writer, binary.BigEndian, &syn)
	return writer.Bytes()
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_synack_default() (hdr EnetPacketHeader, sak EnetPacketSynAck) {
	hdr, syn := enet_packet_syn_default()
	hdr.Type = enet_packet_type_synack
	sak = EnetPacketSynAck(syn)
	return
}
func enet_packet_synack_encode(sak EnetPacketSynAck) []byte {
	writer := bytes.NewBuffer(nil)
	binary.Write(writer, binary.BigEndian, &sak)
	return writer.Bytes()
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_fin_default() (hdr EnetPacketHeader) {
	hdr.Type = enet_packet_type_fin
	hdr.Flags = enet_packet_header_flags_needack
	hdr.ChannelID = enet_channel_id_none
	hdr.Size = uint32(binary.Size(hdr))
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_ping_default(chanid uint8) (hdr EnetPacketHeader) {
	hdr.Type = enet_packet_type_ping
	hdr.Flags = enet_packet_header_flags_needack
	hdr.ChannelID = chanid
	hdr.Size = uint32(binary.Size(hdr))
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_reliable_default(chanid uint8, payloadlen uint32) (hdr EnetPacketHeader) {
	hdr.Type = enet_packet_type_reliable
	hdr.Flags = enet_packet_header_flags_needack
	hdr.ChannelID = chanid
	hdr.Size = uint32(binary.Size(hdr)) + payloadlen
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_unreliable_default(chanid uint8, payloadlen, usn uint32) (hdr EnetPacketHeader, pkt EnetPacketUnreliable) {
	hdr.Type = enet_packet_type_unreliable
	hdr.Flags = 0
	hdr.ChannelID = chanid
	hdr.Size = uint32(binary.Size(hdr)+binary.Size(pkt)) + payloadlen
	pkt.SN = usn
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_fragment_default(chanid uint8, fraglen uint32) (hdr EnetPacketHeader, pkt EnetPacketFragment) {
	hdr.Type = enet_packet_type_fragment
	hdr.Flags = enet_packet_header_flags_needack
	hdr.ChannelID = chanid
	hdr.Size = uint32(binary.Size(hdr)+binary.Size(pkt)) + fraglen
	return
}

// 完成 enet_packet_header的填充，没有具体的packetheader填充
func enet_packet_eg_default() (hdr EnetPacketHeader) {
	hdr.Type = enet_packet_type_fragment
	hdr.Flags = enet_packet_header_flags_needack // should be acked
	hdr.ChannelID = enet_channel_id_none
	hdr.Size = uint32(binary.Size(hdr))
	return
}
