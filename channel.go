package enet

const channel_wnd_size = 256

type enet_channel_item struct {
	header   enet_packet_header
	fragment enet_packet_fragment // if header.cmd == enet_packet_fragment
	payload  []uint8              // not include packet-header
	retries  int                  // sent times for outgoing packet, acked times for incoming packet
	acked    int                  // acked times
	retrans  *enet_timer_item     // retrans timer
}

// outgoing: ->end ..untransfered.. next ..transfered.. begin ->
// incoming: <-begin ..acked.. next ..unacked.. end<-
type enet_channel struct {
	next_sn        uint32 // next reliable packet number for sent
	next_usn       uint32 // next unsequenced packet number for sent
	outgoing       [channel_wnd_size]*enet_channel_item
	outgoing_used  uint32 // in trans packets not acked
	outgoing_begin uint32 // the first one is not acked yet
	outgoing_end   uint32 // the last one is not acked yet
	outgoing_next  uint32 // the next one would be transfered
	incoming       [channel_wnd_size]*enet_channel_item
	incoming_used  uint32 // rcvd packet count in incoming window
	incoming_begin uint32 // the first one has be received
	incoming_end   uint32 // the last one has been received
	//	incoming_next  uint32 // the next one will be acked
}

func (ch *enet_channel) outgoing_trans(item *enet_channel_item) {
	idx := item.header.sn % channel_wnd_size
	v := ch.outgoing[idx]
	assert(v == nil && item.header.sn == ch.outgoing_end)
	ch.outgoing[idx] = item
	if ch.outgoing_end <= item.header.sn {
		ch.outgoing_end = item.header.sn + 1
	}
	item.retries++
	ch.outgoing_used++
}

func (ch *enet_channel) outgoing_ack(sn uint32) {
	if sn < ch.outgoing_begin || sn >= ch.outgoing_end { // already acked or error
		return
	}
	idx := sn % channel_wnd_size
	v := ch.outgoing[idx]
	assert(v != nil && v.header.sn == sn)

	v.acked++
}

func (ch *enet_channel) outgoing_do_trans() (item *enet_channel_item) {
	assert(ch.outgoing_next <= ch.outgoing_end)
	if ch.outgoing_next >= ch.outgoing_end {
		return
	}
	idx := ch.outgoing_next % channel_wnd_size
	item = ch.outgoing[idx]
	assert(item != nil)
	ch.outgoing_next++
	return
}

// may be retransed packet
func (ch *enet_channel) incoming_trans(item *enet_channel_item) {
	if item.header.sn < ch.incoming_begin {
		return
	}
	idx := item.header.sn % channel_wnd_size
	v := ch.incoming[idx]
	// duplicated packet
	if v != nil {
		return
	}
	assert(v == nil || v.header.sn == item.header.sn)

	ch.incoming[idx] = item
	ch.incoming_used++
	if ch.incoming_end <= item.header.sn {
		ch.incoming_end = item.header.sn + 1
	}
}

// when do ack incoming packets
func (ch *enet_channel) incoming_ack(sn uint32) {
	if sn < ch.incoming_begin || sn >= ch.incoming_end { // reack packet not in wnd
		return
	}
	idx := sn % channel_wnd_size
	v := ch.incoming[idx]
	assert(v != nil && v.header.sn == sn)
	v.acked++
}

// called after incoming-ack
func (ch *enet_channel) incoming_slide() (item *enet_channel_item) { // return value may be ignored
	if ch.incoming_begin >= ch.incoming_end {
		return
	}
	idx := ch.incoming_begin % channel_wnd_size
	v := ch.incoming[idx]
	if v == nil || v.acked <= 0 { // not received yet
		return
	}
	assert(v.header.sn == ch.incoming_begin)

	if v.header.cmd == enet_packet_type_fragment {
		all := true
		for i := uint32(1); i < v.fragment.count; i++ {
			n := ch.incoming[idx+i]
			if n == nil || n.header.sn != v.header.sn+i || n.fragment.sn != v.header.sn {
				all = false
				break
			}
		}
		if !all {
			return
		}

		item = v
		ch.incoming_begin += v.fragment.count
		ch.incoming_used -= v.fragment.count
		for i := uint32(1); i < v.fragment.count; i++ {
			item.payload = append(item.payload, ch.incoming[idx+1].payload...)
			ch.incoming[idx+i] = nil
		}
		ch.incoming[idx] = nil

		return
	}
	item = v
	ch.incoming_begin++
	ch.incoming_used--
	ch.incoming[idx] = nil
	return
}

func (ch enet_channel) is_outgoing_full() bool {
	return ch.outgoing_end-ch.outgoing_begin >= channel_wnd_size
}
func (ch enet_channel) is_incoming_full() bool {
	return ch.incoming_end-ch.incoming_begin >= channel_wnd_size
}
func (ch enet_channel) is_outgoing_null() bool {
	return ch.outgoing_end <= ch.outgoing_begin
}
func (ch enet_channel) is_incoming_null() bool {
	return ch.incoming_end <= ch.incoming_begin
}
func (ch enet_channel) outgoing_intrans() int {
	return int(ch.outgoing_next - ch.outgoing_begin)
}
