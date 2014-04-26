package enet

import ()

const (
	peer_state_none             int = iota
	peer_state_closed               //  disconnected          closed-tcp
	peer_state_syn_sent             // connecting            sync-sent
	peer_state_syn_rcvd             // acking-connect        sync-rcvd
	peer_state_listening            // connection-pending    listening //connection-succeeded
	peer_state_established          //             established
	peer_state_disconnect_later     //      closing(fin_sent+fin_rcvd+fin_ack_sent)
	peer_state_fin_sent             // disconnecting         fin_wait_1(fin_sent)
	peer_state_time_wait            // zombie                time_wait
	peer_state_fin_rcvd             // acking-disconnected   fint_wait_2(fin_ack-rcvd)
	peer_state_nothing
)

func peer_rtt_update(peer *enet_peer, host *enet_host, rtt int64) {
	v := rtt - peer.rtt
	peer.rtt += v / 8
	peer.rttv = peer.rttv - peer.rttv/4 + absi64(v/4)

	peer.lowest_rtt = mini64(peer.lowest_rtt, peer.rtt)
	peer.highest_rttv = maxi64(peer.highest_rttv, peer.rttv)

	if host.now-peer.throttle_epoc > peer.throttle_i {
		peer.last_rtt = peer.lowest_rtt
		peer.last_rttv = peer.highest_rttv
		peer.lowest_rtt = peer.rtt
		peer.highest_rttv = peer.rttv
		peer.throttle_epoc = host.now
	}
}

func peer_throttle_update(peer *enet_peer, host *enet_host, rtt int64) {
	if peer.last_rtt <= peer.last_rttv {
		peer.throttle = enet_throttle_scale
		return
	}
	if rtt < peer.last_rtt {
		peer.throttle = minui32(peer.throttle+peer.throttle_acce, enet_throttle_scale)
		return
	}
	if rtt > peer.last_rtt+peer.last_rttv<<1 {
		peer.throttle = maxui32(peer.throttle-peer.throttle_dece, 0)
	}
}
