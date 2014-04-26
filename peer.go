package enet

import (
	"net"
)

const enet_channel_count = 16

type enet_peer struct {
	lid             uint16 // local peer id
	rid             uint16 // remote peer id
	conn_id         uint32 // connection id
	incoming_sessid uint16 // must be 0
	outgoing_sessid uint16 // outgoing session id, must be 0
	mtu             uint32 // remote mtu
	snd_bandwidth   uint32
	rcv_bandwidth   uint32
	wnd_size        uint32 // bytes
	chan_count      uint32
	throttle        uint32
	throttle_i      int64
	throttle_acce   uint32
	throttle_dece   uint32
	rdata           uint
	ldata           uint // should use uint
	state           int
	wnd_used        int
	ack_used        int
	channel         [enet_channel_count]enet_channel
	wnd             [enet_wnd_count]*enet_command
	acks            [enet_wnd_count]*enet_command
	last_intrans    uint16 // last packet id sent but not acked
	first_ack       uint16 // first ack should be sent in acks
	raddr           *net.UDPAddr
	intrans_bytes   int
	rcvd_bytes      int
	sent_bytes      int
	recv_bps        int
	sent_bps        int
	bps_epoc        int64
	rtt             int64
	rttv            int64
	last_rtt        int64
	last_rttv       int64
	lowest_rtt      int64
	highest_rttv    int64
	rtt_epoc        int64
	throttle_epoc   int64
	timeout_limit   int64
	timeout_min     int64
	timeout_max     int64
}

var (
	enet_peer_id_seed      uint16 = 0
	enet_peer_conn_id_seed uint32 = 0
)

func enet_peer_new(addr net.Addr) *enet_peer {
	enet_peer_id_seed++
	enet_peer_conn_id_seed++
	return &enet_peer{
		lid:           enet_peer_id_seed,
		conn_id:       enet_peer_conn_id_seed,
		state:         peer_state_closed,
		mtu:           enet_mtu_default,
		wnd_size:      enet_wnd_size_default,
		chan_count:    enet_channel_count_default,
		throttle:      enet_throttle_default,
		throttle_i:    enet_throttle_interval_default,
		throttle_acce: enet_throttle_acce_default,
		throttle_dece: enet_throttle_dece_default,
		rtt:           enet_rtt_default,
		last_rtt:      enet_rtt_default,
		lowest_rtt:    enet_rtt_default,
		rtt_epoc:      unixtime_now() - int64(enet_throttle_interval_default<<1),
		throttle_epoc: unixtime_now() - int64(enet_throttle_interval_default<<1),
		timeout_limit: enet_timeout_limit,
		timeout_min:   enet_timeout_min,
		timeout_max:   enet_timeout_max,
		raddr:         addr.(*net.UDPAddr),
	}
}

func (self *enet_peer) connect(host Host) error {
	return nil
}

func (self *enet_peer) disconnect(host Host) error {
	return nil
}

func peer_do_send(peer *enet_peer, host *enet_host) {
	peer_ack_send(peer, host)
	peer_window_send(peer, host)
}

func peer_channel_get(peer *enet_peer, cid uint8) *enet_channel {
	if cid >= enet_channel_count {
		return &enet_channel{}
	}
	v := &peer.channel[cid]
	return v
}

func peer_window_is_full(peer *enet_peer) bool {
	wndsz := int(peer.wnd_size * peer.throttle / enet_throttle_scale)
	wf := wndsz < peer.intrans_bytes
	return wf || peer.wnd_used >= int(enet_wnd_count)
}

func peer_window_is_empty(peer *enet_peer) bool {
	return peer.wnd_used == 0
}

func peer_snd_bandwidth_is_full(peer *enet_peer) bool {
	return false
}
