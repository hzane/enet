package enet

import "net"

type enet_peer struct {
	id            uint16
	rid           uint16
	cid           uint32
	in_sid        uint16
	out_sid       uint16
	state         int
	out_sn        uint16
	mtu           uint32
	snd_bandwidth uint32
	rcv_bandwidth uint32
	wnd_size      uint32
	chcount       uint32
	throttle_i    uint32
	throttle_acce uint32
	throttle_dece uint32
	data          uint32
	raddr         *net.UDPAddr
}

func enet_peer_new() *enet_peer {
	return &enet_peer{}
}

func (self *enet_peer) connect(host Host) error {
	return nil
}

func (self *enet_peer) disconnect(host Host) error {
	return nil
}

func (self *enet_peer) handle_recv(host Host, hdr enet_packet_header, reader enet_reader) {
	handler := enet_peer_state_handler_get(self.state)
	handler(self, host.(*enet_host), hdr, reader)
	if hdr.need_ack != 0 {
		self.send_ack(host, hdr.sn)
	}
}

func (self *enet_peer) send_ack(Host, uint16) {

}

func (self *enet_peer) out_sn_inc() uint16 {
	v := self.out_sn
	self.out_sn++
	return v
}
