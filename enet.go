package enet

/*
udp_header | protocol_header | crc32_header? | ( packet_header | payload )+
*/
type EnetProtocolHeader struct {
	PeerID      uint16 // target peerid, not used
	Flags       uint8  // 0xcc : use crc32 header, default 0
	PacketCount uint8  // enet_packets in this datagram
	SntTime     uint32 // milli-second, sent-time
	ClientID    uint32 // client-id? , server would fill client's id, not his own
}
type EnetCrc32Header struct {
	CRC32 uint32
}
type EnetPacketHeader struct {
	Type      uint8  // enet_packet_type_xxx
	Flags     uint8  // _needack, _forcefin_, enet_packet_header_flags_xxx
	ChannelID uint8  // [0,n), or 0xff; oxff : control channel
	RSV       uint8  // not used
	Size      uint32 // including packet_header and payload, bytes
	SN        uint32 // used for any packet type which should be acked, not used for unreliable, ack
}

//cmd_type_ack = 1
// flags must be zero
type EnetPacketAck struct {
	SN      uint32 // rcvd-sn // not the next sn
	SntTime uint32 // rcvd sent time
}

//cmd_type_syn = 2
// flags = enet_packet_needack
type EnetPacketSyn struct { // ack by conack
	PeerID           uint16 // zero, whose id write the packet
	MTU              uint16 // default = 1200
	WndSize          uint32 // local recv window size, 0x8000
	ChannelCount     uint32 // channels count, default = 2
	RcvBandwidth     uint32 // local receiving bandwith bps, 0 means no limit
	SndBandwidth     uint32 // local sending bandwidth , 0 means no limit
	ThrottleInterval uint32 // = 0x1388 = 5000ms
	ThrottleAcce     uint32 // = 2
	ThrottleDece     uint32 // = 2
}

//cmd_type_synack = 3
// flags = enet_packet_needack
type EnetPacketSynAck EnetPacketSyn

//cmd_type_fin = 4
// flags = enet_packet_flags_forcefin_ if disconnect unconnected peer
type EnetPacketFin struct{}

// cmd_type = 5
type EnetPacketPing struct{}

//cmd_type_reliable = 6
// flags= enet_packet_header_flags_needack
type EnetPacketReliable struct{}

//cmd_type_unreliable = 7
// flags = enet_packet_header_flags_none
type EnetPacketUnreliable struct {
	SN uint32 // unreliable sequence number, filled with channel.next_usn
}

//cmd_type_fragment = 8
// [offset, length) of the packet sn
// packet was splitted into fragment_count parts
type EnetPacketFragment struct {
	SN     uint32 // start sequence number
	Count  uint32 // fragment counts
	Index  uint32 // index of current fragment
	Size   uint32 // total length of all fragments
	Offset uint32
}
