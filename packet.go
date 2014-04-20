package enet

import (
	"encoding/binary"
	"io"
	"time"
)

type enet_packet_header struct {
	pid      uint16 // peer-id, must be target's id, time(1bit)+compress(1bit)+sess(2bit)+id, < 0xfff
	sid      uint16 // session id 2bits
	snt_time uint16 // millisecond, if has time-flag, unixtime casted to uint16
	crc      uint32 // not supported yet
	cmd      uint8  // packet-type + need_ack + is_unseq
	need_ack uint8  // 1 bit
	is_unseq uint8  // 1 bit
	cid      uint8  // channel id
	sn       uint16 // reliable sequence number
}

type enet_packet_raw_header struct {
	pid      uint16
	snt_time uint16
	cmd      uint8
	cid      uint8
	sn       uint16
}

func enet_packet_header_raw_decode(reader enet_reader) (packet enet_packet_raw_header, err error) {
	if reader.bytes() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}
	binary.Read(reader.(io.Reader), binary.BigEndian, &packet)
	return
}
func enet_packet_header_raw_encode(hdr enet_packet_header) (raw enet_packet_raw_header) {
	raw.pid = (hdr.pid & enet_peerid_mask) | enet_peerid_flag_time
	raw.cmd = hdr.cmd | hdr.is_unseq | hdr.need_ack
	raw.snt_time = unixtimeui16(time.Now())
	raw.cid = hdr.cid
	raw.sn = hdr.sn
	return raw
}

type enet_packet_syn struct { // ack by conack
	//	enet_packet_header
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

type enet_reader interface {
	io.Reader
	bytes() int
	uint8() uint8
	uint16() uint16
	uint32() uint32
	uint64() uint64
	left() []uint8
}

func enet_reader_new([]uint8) enet_reader {
	return nil
}
func enet_packet_syn_decode(reader enet_reader) (packet enet_packet_syn, err error) {
	// pid, insid, outsid, mtu, wnd, ccount, bw, bw, thro, throa, throd, cid, data
	if reader.bytes() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}

type enet_packet_ack struct {
	//	enet_packet_header
	sn   uint16 // rcvd-sn // not the next sn
	time uint16 // rcvd sent time
}

func enet_packet_ack_decode(reader enet_reader) (packet enet_packet_ack, err error) {
	if reader.bytes() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}

/*
type enet_packet_connack struct {
	//	enet_packet_header
	pid            uint16 // negative peer-id
	in_sid         uint8  // zero
	out_sid        uint8  // zero
	mtu            uint32
	wnd_size       uint32
	ccount         uint32 // channel count
	rcv_bandwidth  uint32 // rcv bandwidth
	snd_bandwidth  uint32 // snd bandwidth
	pthrottle_i    uint32
	pthrottle_acce uint32
	pthrottle_dece uint32
	conn_id        uint32
}*/
type enet_packet_synack enet_packet_syn

func enet_packet_synack_decode(reader enet_reader) (packet enet_packet_synack, err error) {
	if reader.bytes() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}

	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}

type enet_packet_fin struct {
	//	enet_packet_header
	data uint32
}

func enet_packet_fin_decode(reader enet_reader) (packet enet_packet_fin, err error) {
	if reader.bytes() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}
	err = binary.Read(reader, binary.BigEndian, &packet)
	return
}

type enet_packet_ping struct{} // should ack	//	enet_packet_header

type enet_packet_bandwidth_limit struct {
	//	enet_packet_header
	rcv_bandwidth uint32
	snd_bandwidth uint32
}

type enet_packet_throttle_configure struct {
	//	enet_packet_header
	interval     uint32
	acceleration uint32
	deceleration uint32
}
type enet_packet_reliable struct { // ack is needed
	//	enet_packet_header
	length uint16 // datalength
	// payload [length]uint8
}

type enet_packet_unreliable struct {
	//	enet_packet_header

	u_sn   uint16 // unreliable sequence number
	length uint16
	// playload [length]uint8
}

type enet_packet_unsequence struct {
	//	enet_packet_header
	group   uint16
	length  uint16
	payload []uint8
}

// [offset, length) of the packet sn
// packet was splitted into fragment_count parts
type enet_packet_fragment struct {
	enet_packet_header
	sn              uint16
	length          uint16 // fragment length
	fragment_count  uint32
	fragment_n      uint32
	packet_length   uint32
	fragment_offset uint32
	payload         []uint8 // length
}

func enet_packet_header_decode(reader enet_reader) (hdr enet_packet_header, err error) {
	packet, err := enet_packet_header_raw_decode(reader)
	if err != nil {
		return
	}
	hdr.pid = packet.pid & enet_peerid_mask
	hdr.sid = packet.pid & enet_peerid_session_mask >> enet_peerid_session_shift
	if packet.pid&enet_peerid_flag_time == 0 || packet.pid&enet_peerid_flag_compressed != 0 {
		err = enet_err_unsupported_flags
		return
	}
	hdr.snt_time = packet.snt_time
	hdr.cmd = packet.cmd & enet_cmd_mask
	hdr.cid = packet.cid
	hdr.sn = packet.sn
	hdr.need_ack = packet.cmd & enet_cmd_flag_ack
	hdr.is_unseq = packet.cmd & enet_cmd_flag_unseq
	return
}

func enet_packet_header_decode2(reader enet_reader) (hdr enet_packet_header, err error) {
	if reader.bytes() < 2 {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}
	pidflag := reader.uint16()
	hdr.pid = pidflag & enet_peerid_mask
	hdr.sid = (pidflag & enet_peerid_session_mask) >> enet_peerid_session_shift

	hast := pidflag & enet_peerid_flag_time

	//	compressed := pidflag & enet_peerid_flag_compressed

	// we dont support compressor
	hascrc := false // host defined

	// sent-time, crc, cmd, cid, sn
	r := bool_int(hast != 0)*2 + bool_int(hascrc)*4 + 1 + 1 + 2
	if reader.bytes() < r {
		err = enet_err_invalid_packet_size(reader.bytes())
		return
	}

	if hast != 0 {
		hdr.snt_time = reader.uint16()
	}
	if hascrc {
		reader.uint32() // ignore crc
	}

	cmdflag := reader.uint8()
	hdr.cmd = cmdflag & enet_cmd_mask
	hdr.need_ack = cmdflag & enet_cmd_flag_ack
	hdr.is_unseq = cmdflag & enet_cmd_flag_unseq

	hdr.cid = reader.uint8() // channel id
	hdr.sn = reader.uint16()
	return
}

func bool_int(v bool) int {
	if v {
		return 1
	}
	return 0
}
