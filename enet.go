package enet

import "net"

const (
	enet_cmd_none uint8 = iota
	enet_cmd_ack
	enet_cmd_syn      // with ack flag, connect
	enet_cmd_synack   //
	enet_cmd_fin      // ack when under connected or disconnect-later state, else use unseq
	enet_cmd_ping     // ack
	enet_cmd_reliable // ack
	enet_cmd_unreliable
	enet_cmd_fragment       // ack
	enet_cmd_unsequenced    // +unseq flag
	enet_cmd_bandwidthlimit // ack flag
	enet_cmd_throttle       // ack flag
	enet_cmd_unreliable_fragment
	enet_cmds_count
)

const (
	enet_cmd_mask       uint8 = 0x0f
	enet_cmd_flag_ack         = 1 << 7
	enet_cmd_flag_unseq       = 1 << 6
)

const (
	enet_peerid_flag_compressed uint16 = 1 << 14
	enet_peerid_flag_time              = 1 << 15
	enet_peerid_session_mask           = 3 << 12
	enet_peerid_session_shift          = 12
	enet_peerid_mask                   = 0x0fff
)

const (
	enet_wnd_size_min      uint32 = 4096
	enet_wnd_size_max             = 65536
	enet_mtu_min                  = 576
	enet_mtu_max                  = 4096
	enet_channel_count_min        = 1
	enet_channel_count_max        = 255
)

type enet_address *net.UDPAddr
