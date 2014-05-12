package enet

/*
udp_header | protocol_header | crc32_header? | ( packet_header | payload )+
*/
type enet_protocol_header struct {
	peerid   uint16 // target peerid, not used
	flags    uint8  // 0xcc if crc or 0
	packet_n uint8
	time     uint32 // milli-second, sent-time
	clientid uint32 // client-id?
}
type enet_crc32_header struct {
	crc32 uint32
}
type enet_packet_header struct {
	cmd    uint8
	flags  uint8 // _needack, _forcefin_
	chanid uint8
	rsv    uint8
	size   uint32 // including packet_header
	sn     uint32
}

//cmd_type_ack = 1
// flags must be zero
type enet_packet_ack struct {
	sn uint32 // rcvd-sn // not the next sn
	tm uint32 // rcvd sent time
}

//cmd_type_syn = 2
// flags = enet_packet_needack
type enet_packet_syn struct { // ack by conack
	peerid            uint16 // zero, whose id write the packet
	mtu               uint16 // default = 1200
	wnd_size          uint32 // local recv window size, 0x8000
	channel_n         uint32 // channels count, default = 2
	rcv_bandwidth     uint32 // local receiving bandwith bps, 0 means no limit
	snd_bandwidth     uint32 // local sending bandwidth , 0 means no limit
	throttle_interval uint32 // = 0x1388 = 5000ms
	throttle_acce     uint32 // = 2
	throttle_dece     uint32 // = 2
}

//cmd_type_synack = 3
// flags = enet_packet_needack
type enet_packet_synack enet_packet_syn

//cmd_type_fin = 4
// flags = enet_packet_flags_forcefin_ if disconnect unconnected peer
type enet_packet_fin struct{}

// cmd_type = 5
type enet_packet_ping struct{}

//cmd_type_reliable = 6
// flags= enet_packet_header_flags_needack
type enet_packet_reliable struct{}

//cmd_type_unreliable = 7
// flags = enet_packet_header_flags_none
type enet_packet_unreliable struct {
	usn uint32 // unreliable sequence number
}

//cmd_type_fragment = 8
// [offset, length) of the packet sn
// packet was splitted into fragment_count parts
type enet_packet_fragment struct {
	sn     uint32 // start sequence number
	count  uint32 // fragment counts
	index  uint32 // index of current fragment
	length uint32 // total length of all fragments
	offset uint32
}
