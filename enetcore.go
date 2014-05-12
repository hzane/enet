package enet

const (
	enet_packet_type_unspec     uint8 = iota
	enet_packet_type_ack              = 1
	enet_packet_type_syn              = 2
	enet_packet_type_synack           = 3
	enet_packet_type_fin              = 4
	enet_packet_type_ping             = 5
	enet_packet_type_reliable         = 6
	enet_packet_type_unreliable       = 7
	enet_packet_type_fragment         = 8
	enet_packet_type_eg               = 12
	enet_packet_type_count            = 12
)

const (
	enet_protocol_flags_none uint8 = iota
	enet_protocol_flags_crc        = 0xcc
)

const (
	enet_packet_header_flags_none uint8 = iota
	enet_packet_header_flags_needack
	enet_packet_header_flags_forcefin_
)

/*
const (
	enet_protocol_header_bytes     = 12
	enet_protocol_crc_bytes        = 4
	enet_protocol_ack_bytes        = enet_protocol_header_bytes + 8
	enet_protocol_syn_bytes        = enet_protocol_header_bytes + 32
	enet_protocol_synack_bytes     = enet_protocol_syn_bytes
	enet_protocol_unreliable_bytes = enet_protocol_header_bytes + 4
	enet_protocol_fragment_bytes   = enet_protocol_header_byte + 20
	enet_protocol_fin_bytes        = enet_protocol_header_bytes
)
*/
const (
	enet_default_mtu               = 1200
	enet_default_wndsize           = 0x8000 // bytes
	enet_default_channel_count     = 2
	enet_default_throttle_interval = 5000 // ms
	enet_default_throttle_acce     = 2
	enet_default_throttle_dece     = 2
	enet_default_tick_ms           = 20 // ms
	enet_udp_size                  = 65536
	enet_default_throttle          = 32
	enet_throttle_scale            = 32
	enet_default_rtt               = 50 //ms
	enet_timeout_limit             = enet_default_rtt * 15
	enet_timeout_min               = enet_default_rtt
	enet_timeout_max               = enet_timeout_limit
)
const enet_channel_id_none uint8 = 0xff

/* uncompatiable with enet origin protocol
enet_cmd_unsequenced    // +unseq flag
enet_cmd_bandwidthlimit // ack flag
enet_cmd_throttle       // ack flag
enet_cmd_unreliable_fragment
*/
