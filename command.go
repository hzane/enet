package enet

import "net"

// channel_id : 0xff , unsequenced
// reliable : false, unreliable data
// func EnetSendTo(data []byte, channelid uint8, reliable bool, addr net.Addr)
type enet_command struct {
	enet_packet_header
	data     []byte
	addr     *net.UDPAddr
	heap_idx int
	timeo    int64
}

//send-reliable command
//send-unreliable command
//send-unsequence command
// disconnect command?

// rcvd command
func enet_command_reliable_new(cid uint8, data []byte, peer *enet_peer) enet_command {
	cmd := enet_command{}
	cmd.cmd = enet_cmd_reliable
	cmd.need_ack = enet_cmd_flag_ack

	cmd.snt_time = unixtime_nowui16()
	cmd.cid = cid
	cmd.addr = peer.raddr
	cmd.data = data
	cmd.pid = peer.rid
	cmd.sid = peer.outgoing_sessid
	if cid != enet_channel_id_none {
		cmd.sn = peer_channel_get(peer, cid).out_sn()
	}
	return cmd
}
