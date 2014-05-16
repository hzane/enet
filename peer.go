package enet

import (
	"bytes"
	"encoding/binary"
	"net"
)

type enet_peer struct {
	clientid          uint32 // local peer id
	mtu               uint32 // remote mtu
	snd_bandwidth     uint32
	rcv_bandwidth     uint32
	wnd_bytes         uint32
	chan_count        uint8
	throttle          uint32
	throttle_interval int64
	throttle_acce     uint32
	throttle_dece     uint32
	rdata             uint
	ldata             uint // should use uint
	flags             int
	wnd_size          uint32 // bytes, calculated by throttle and wnd_bytes
	intrans_bytes     int
	channel           [enet_default_channel_count + 1]enet_channel
	remote_addr       *net.UDPAddr
	rcvd_bytes        int
	sent_bytes        int
	rcvd_bps          int
	sent_bps          int
	rtt_timeo         int64
	rtt               int64 // ms
	rttv              int64
	last_rtt          int64
	last_rttv         int64
	lowest_rtt        int64
	highest_rttv      int64
	rtt_epoc          int64
	throttle_epoc     int64
	timeout_limit     int64
	timeout_min       int64
	timeout_max       int64
	host              *enet_host
}

func new_enet_peer(addr *net.UDPAddr, host *enet_host) *enet_peer {
	cid := host.next_clientid
	host.next_clientid++
	return &enet_peer{
		clientid:          cid,
		flags:             0,
		mtu:               enet_default_mtu,
		wnd_size:          enet_default_wndsize,
		chan_count:        enet_default_channel_count,
		throttle:          enet_default_throttle,
		throttle_interval: enet_default_throttle_interval,
		throttle_acce:     enet_default_throttle_acce,
		throttle_dece:     enet_default_throttle_dece,
		rtt:               enet_default_rtt,
		last_rtt:          enet_default_rtt,
		lowest_rtt:        enet_default_rtt,
		rtt_epoc:          0, // may expire as soon as fast
		throttle_epoc:     0, // may expire immediately
		timeout_limit:     enet_timeout_limit,
		timeout_min:       enet_timeout_min,
		timeout_max:       enet_timeout_max,
		remote_addr:       addr,
		host:              host,
	}
}
func (peer *enet_peer) do_send(hdr EnetPacketHeader, frag EnetPacketFragment, dat []byte) {
	writer := bytes.NewBuffer(nil)
	phdr := EnetProtocolHeader{0, 0, 1, uint32(peer.host.now), peer.clientid}
	binary.Write(writer, binary.BigEndian, phdr)
	binary.Write(writer, binary.BigEndian, &hdr)
	if hdr.Type == enet_packet_type_fragment {
		binary.Write(writer, binary.BigEndian, &frag)
	}
	binary.Write(writer, binary.BigEndian, dat)
	peer.host.do_send(writer.Bytes(), peer.remote_addr)
}
func (peer *enet_peer) channel_from_id(cid uint8) *enet_channel {
	if cid >= peer.chan_count {
		return &peer.channel[enet_default_channel_count]
	}
	v := &peer.channel[cid]
	return v
}

func peer_window_is_full(peer *enet_peer) bool {
	return peer.intrans_bytes >= int(peer.wnd_size)
}

func peer_window_is_empty(peer *enet_peer) bool {
	return peer.intrans_bytes == 0
}

func (peer *enet_peer) when_enet_incoming_ack(header EnetPacketHeader, payload []byte) {
	if peer.flags&enet_peer_flags_stopped != 0 {
		return
	}
	reader := bytes.NewReader(payload)
	var ack EnetPacketAck
	err := binary.Read(reader, binary.BigEndian, &ack)

	if err != nil {
		return
	}
	rtt := peer.host.now - int64(ack.SntTime)
	peer.update_rtt(rtt)
	peer.update_throttle(rtt)

	ch := peer.channel_from_id(header.ChannelID)
	ch.outgoing_ack(ack.SN)
	for i := ch.outgoing_slide(); i != nil; i = ch.outgoing_slide() {
		if i.retrans != nil {
			peer.host.timers.remove(i.retrans.index)
			i.retrans = nil
		}
		if i.header.Type == enet_packet_type_syn {
			peer.flags |= enet_peer_flags_syn_sent
			if peer.flags&enet_peer_flags_synack_rcvd != 0 {
				peer.flags |= enet_peer_flags_established
				notify_peer_connected(peer, 0)
			}
		}
		if i.header.Type == enet_packet_type_fin {
			peer.flags |= enet_peer_flags_stopped
			notify_peer_disconnected(peer, 0)
			peer.host.destroy_peer(peer)
		}
	}
}
func notify_data(peer *enet_peer, dat []byte) {

}
func notify_peer_connected(peer *enet_peer, ret int) {

}
func notify_peer_disconnected(peer *enet_peer, ret int) {

}
func (peer *enet_peer) reset() {

}
func (peer *enet_peer) handshake(syn EnetPacketSyn) {

}
func (peer *enet_peer) when_enet_incoming_syn(header EnetPacketHeader, payload []byte) {
	reader := bytes.NewReader(payload)
	var syn EnetPacketSyn
	err := binary.Read(reader, binary.BigEndian, &syn)

	if err != nil || peer.flags&enet_peer_flags_synack_sending != 0 {
		return
	}
	if peer.flags&(enet_peer_flags_syn_sent|enet_peer_flags_syn_rcvd) != 0 {
		peer.reset()
	}
	peer.handshake(syn)
	// send synack
	peer.flags |= enet_peer_flags_synack_sending | enet_peer_flags_syn_rcvd
	ch := peer.channel_from_id(enet_channel_id_none)
	phdr, synack := enet_packet_synack_default()

	// todo add retrans timer
	//	ch.outgoing_pend(phdr, enet_packet_fragment{}, enet_packet_synack_encode(synack), nil)
	peer.outgoing_pend(ch, phdr, EnetPacketFragment{}, enet_packet_synack_encode(synack))
}

func (peer *enet_peer) when_enet_incoming_synack(header EnetPacketHeader, payload []byte) {
	reader := bytes.NewReader(payload)
	var syn EnetPacketSyn
	err := binary.Read(reader, binary.BigEndian, &syn)

	if err != nil || peer.flags&enet_peer_flags_syn_sending == 0 {
		peer.reset()
		return
	}
	peer.handshake(syn)
	peer.flags |= enet_peer_flags_synack_rcvd
	if peer.flags&enet_peer_flags_syn_sent != 0 {
		peer.flags |= enet_peer_flags_established
		notify_peer_connected(peer, 0)
	}
}

func (peer *enet_peer) when_enet_incoming_fin(header EnetPacketHeader, payload []byte) {
	if peer.flags&enet_peer_flags_fin_sending != 0 {
		// needn't do anything, just wait for self fin's ack
		return
	}
	peer.flags |= enet_peer_flags_fin_rcvd | enet_peer_flags_lastack // enter time-wait state
	notify_peer_disconnected(peer, 0)

	peer.host.timers.push(peer.host.now+peer.rtt_timeo*2, func() { peer.host.destroy_peer(peer) })
}

func (peer *enet_peer) when_enet_incoming_ping(header EnetPacketHeader, payload []byte) {

}

func (peer *enet_peer) when_enet_incoming_reliable(header EnetPacketHeader, payload []byte) {
	if peer.flags&enet_peer_flags_established == 0 {
		return
	}
	ch := peer.channel_from_id(header.ChannelID)
	if ch == nil {
		return
	}
	ch.incoming_trans(&enet_channel_item{header, EnetPacketFragment{}, payload, 0, 0, nil})
	ch.incoming_ack(header.SN)
	for i := ch.incoming_slide(); i != nil; i = ch.incoming_slide() {
		notify_data(peer, i.payload)
	}
}

func (peer *enet_peer) when_enet_incoming_fragment(header EnetPacketHeader, payload []byte) {
	reader := bytes.NewReader(payload)
	var frag EnetPacketFragment
	binary.Read(reader, binary.BigEndian, &frag)

	dat := make([]byte, int(header.Size)-binary.Size(frag))
	reader.Read(dat)
	ch := peer.channel_from_id(header.ChannelID)
	if ch == nil {
		return
	}
	ch.incoming_trans(&enet_channel_item{header, frag, dat, 0, 0, nil})
	ch.incoming_ack(header.SN)
	for i := ch.incoming_slide(); i != nil; i = ch.incoming_slide() {
		notify_data(peer, i.payload)
	}
}

func (peer *enet_peer) when_enet_incoming_unrelialbe(header EnetPacketHeader, payload []byte) {
	reader := bytes.NewReader(payload)
	var ur EnetPacketUnreliable
	binary.Read(reader, binary.BigEndian, &ur)

	dat := make([]byte, int(header.Size)-binary.Size(ur))
	reader.Read(dat)
	notify_data(peer, dat)
}
func (peer *enet_peer) when_unknown(header EnetPacketHeader, payload []byte) {
	debugf("peer skipped packet : %v\n", header.Type)
}
func (peer *enet_peer) when_enet_incoming_eg(header EnetPacketHeader, payload []byte) {

}

const (
	enet_peer_flags_none        = 1 << iota // never used
	enet_peer_flags_sock_closed             // sock is closed
	enet_peer_flags_stopped                 // closed, rcvd fin, and sent fin+ack and then rcvd fin+ack's ack
	enet_peer_flags_lastack                 // send fin's ack, and waiting retransed fin in rtttimeout
	enet_peer_flags_syn_sending             // connecting            sync-sent
	enet_peer_flags_syn_sent                // syn acked
	enet_peer_flags_syn_rcvd                // acking-connect        sync-rcvd
	enet_peer_flags_listening               // negative peer
	enet_peer_flags_established             // established
	enet_peer_flags_fin_sending             // sent fin, waiting the ack
	enet_peer_flags_fin_sent                // rcvd fin's ack
	enet_peer_flags_fin_rcvd                //
	enet_peer_flags_nothing
	enet_peer_flags_synack_sending
	enet_peer_flags_synack_rcvd
	enet_peer_flags_synack_sent
)

func (peer *enet_peer) update_rtt(rtt int64) {
	v := rtt - peer.rtt
	peer.rtt += v / 8
	peer.rttv = peer.rttv - peer.rttv/4 + absi64(v/4)

	peer.lowest_rtt = mini64(peer.lowest_rtt, peer.rtt)
	peer.highest_rttv = maxi64(peer.highest_rttv, peer.rttv)

	if peer.host.now > peer.throttle_interval+peer.throttle_epoc {
		peer.throttle_epoc = peer.host.now
		peer.last_rtt = peer.lowest_rtt
		peer.last_rttv = peer.highest_rttv
		peer.lowest_rtt = peer.rtt
		peer.highest_rttv = peer.rttv
	}
	peer.rtt_timeo = rtt + peer.rttv<<1
}

func (peer *enet_peer) update_throttle(rtt int64) {
	// unstable network
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

func (peer *enet_peer) update_window_size() {
	peer.wnd_size = peer.wnd_bytes * peer.throttle / enet_throttle_scale
}

func (peer *enet_peer) update_statis(itv int) {
	peer.rcvd_bps = peer.rcvd_bytes * 1000 / itv
	peer.rcvd_bytes = 0
	peer.sent_bps = peer.sent_bytes * 1000 / itv
	peer.sent_bytes = 0
}
func (peer *enet_peer) Addr() net.Addr {
	return peer.remote_addr
}

// do with 2 times retry
func (peer *enet_peer) outgoing_pend(ch *enet_channel,
	hdr EnetPacketHeader,
	frag EnetPacketFragment,
	dat []byte) {
	item := &enet_channel_item{hdr, frag, dat, 0, 0, nil}
	reset := func() {

	}
	reretrans := func() {
		item.retrans = peer.host.timers.push(peer.host.now+peer.rtt_timeo, reset)
		item.retries++
		peer.do_send(item.header, item.fragment, item.payload)
	}
	retrans := func() {
		item.retrans = peer.host.timers.push(peer.host.now+peer.rtt_timeo, reretrans)
		item.retries++
		peer.do_send(item.header, item.fragment, item.payload)
	}
	if hdr.Type != enet_packet_type_ack {
		item.retrans = peer.host.timers.push(peer.host.now+peer.rtt_timeo, retrans)
	}
	ch.outgoing_pend(item)
}
