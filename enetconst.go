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
	enet_protocol_flags_crc        = 0xcc // use enet_crc32_header
)

const (
	enet_packet_header_flags_none      uint8 = iota
	enet_packet_header_flags_needack         // for syn, syncak, fin, reliable, ping, fragment
	enet_packet_header_flags_forcefin_       // i don't know how to use this flag
)

const (
	enet_default_mtu               = 1400
	enet_default_wndsize           = 0x8000 // bytes
	enet_default_channel_count     = 2
	enet_default_throttle_interval = 5000 // ms
	enet_default_throttle_acce     = 2
	enet_default_throttle_dece     = 2
	enet_default_tick_ms           = 20  // ms
	enet_default_rtt               = 500 //ms
	enet_udp_size                  = 65536
	enet_default_throttle          = 32
	enet_throttle_scale            = 32
	enet_timeout_limit             = 32   // 30 seconds
	enet_timeout_min               = 5000 // 5 second
	enet_timeout_max               = 30000
	enet_ping_interval             = 1000 // 1 second
	enet_bps_interval              = 1000 // 1 second
)
const (
	enet_channel_id_none uint8 = 0xff
	enet_channel_id_all        = 0xfe
)

const (
	enet_peer_connect_result_duplicated = 1
	enet_peer_disconnect_result_invalid
)
const enet_peer_id_any uint32 = 0xffffffff

/* uncompatiable with enet origin protocol
enet_cmd_unsequenced    // +unseq flag
enet_cmd_bandwidthlimit // ack flag
enet_cmd_throttle       // ack flag
enet_cmd_unreliable_fragment
*/
