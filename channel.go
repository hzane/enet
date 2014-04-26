package enet

const channel_wnd_size = 256
const enet_channel_id_none uint8 = 0xff

type enet_channel struct {
	_out_sn      uint16
	_out_usn     uint16
	expected_sn  uint16
	expected_usn uint16
	wnd          [channel_wnd_size]uint16
	wnd_used     int
	last_acked   uint16
	last_intrans uint16
}

func (ch enet_channel) wnd_is_full() bool {
	return ch.wnd_used >= channel_wnd_size
}

func (ch *enet_channel) wnd_put_in(sn uint16) {
	idx := sn % channel_wnd_size
	v := ch.wnd[idx]
	if v == 0 {
		ch.wnd_used++
		ch.last_intrans = sn
	}
	ch.wnd[idx]++
}

func (ch *enet_channel) wnd_pop_out(sn uint16) {
	idx := sn % channel_wnd_size
	v := ch.wnd[idx]
	if v <= 0 {
		return
	}
	ch.wnd_used--
	ch.wnd[idx] = 0
	ch.last_acked = sn
}
func (ch *enet_channel) out_sn() uint16 {
	v := ch._out_sn
	ch._out_sn++
	return v
}

func (ch *enet_channel) out_usn() uint16 {
	v := ch._out_usn
	ch._out_usn++
	return v
}
