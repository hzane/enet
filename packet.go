package enet

import "encoding/binary"

type enet_packet_syn struct { // ack by conack
	pid           uint16 // positive peer id, who start the connection
	in_sid        uint8  // 0xff
	out_sid       uint8  // 0xff
	mtu           uint32
	wnd_size      uint32 // local recv window size
	ccount        uint32 // channels count, not zero?
	rcv_bandwidth uint32 // local receiving bandwith bps, 0 means no limit
	snd_bandwidth uint32 // local sending bandwidth , 0 means no limit
	throttle_i    uint32 // packet throttle refresh interval, defined
	throttle_acce uint32 // packet throttle acceleration / throttle_scale defined by programm
	throttle_dece uint32 // packet throttle deceleration, defined by programm
	conn_id       uint32 // connection id = random
	data          uint32 // user data
}

type enet_packet_ack struct {
	sn   uint16 // rcvd-sn // not the next sn
	time uint16 // rcvd sent time
}
type enet_packet_synack enet_packet_syn

type enet_packet_fin struct {
	data uint32
}

type enet_packet_ping struct{} // should ack	//	enet_packet_header

type enet_packet_bandwidth_limit struct {
	rcv_bandwidth uint32
	snd_bandwidth uint32
}

type enet_packet_throttle_configure struct {
	interval     uint32
	acceleration uint32
	deceleration uint32
}
type enet_packet_reliable struct { // ack is needed
	length uint16 // datalength
}

type enet_packet_unreliable struct {
	u_sn   uint16 // unreliable sequence number
	length uint16
}

type enet_packet_unsequence struct {
	group  uint16
	length uint16
}

// [offset, length) of the packet sn
// packet was splitted into fragment_count parts
type enet_packet_fragment struct {
	sn              uint16
	length          uint16 // fragment length
	fragment_count  uint32
	fragment_n      uint32
	packet_length   uint32
	fragment_offset uint32
}

func enet_packet_syn_decode(reader enet_reader) (packet enet_packet_syn, err error) {
	// pid, insid, outsid, mtu, wnd, ccount, bw, bw, thro, throa, throd, cid, data
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}

func enet_packet_ack_decode(reader enet_reader) (packet enet_packet_ack, err error) {
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}
func enet_packet_bandwidth_limit_decode(reader enet_reader) (packet enet_packet_bandwidth_limit, err error) {
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}
func enet_packet_throttle_decode(reader enet_reader) (packet enet_packet_throttle_configure, err error) {
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}
func enet_packet_synack_decode(reader enet_reader) (packet enet_packet_synack, err error) {
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}

	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}

func enet_packet_fin_decode(reader enet_reader) (packet enet_packet_fin, err error) {
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}
