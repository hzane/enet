package enet

import (
	"encoding/binary"
	"io"
)

type enet_packet_raw_header struct {
	pid      uint16 // pid, time-flag, compress-flag, sid
	snt_time uint16
	cmd      uint8 // packet-type + need_ack + is_unseq
	cid      uint8
	sn       uint16
}

type enet_packet_header struct {
	pid      uint16 //
	sid      uint16 // session id 2bits
	snt_time uint16 // millisecond, if has time-flag, unixtime casted to uint16
	crc      uint32 // not supported yet
	cmd      uint8  //
	need_ack uint8  // 1 bit enet_cmd_flag_ack
	is_unseq uint8  // 1 bit, enet_cmd_flag_unseq
	cid      uint8  // channel id
	sn       uint16 // reliable sequence number
}

func enet_packet_header_default() enet_packet_header {
	return enet_packet_header{
		pid:      0,
		sid:      0,
		snt_time: unixtime_nowui16(),
		crc:      0,
		cmd:      0,
		need_ack: 0,
		is_unseq: 0,
		cid:      0xff,
		sn:       0,
	}
}
func enet_packet_header_raw_decode(reader enet_reader) (packet enet_packet_raw_header, err error) {
	if reader.left() < binary.Size(packet) {
		err = enet_err_invalid_packet_size(reader.left())
		return
	}
	binary.Read(reader.(io.Reader), binary.BigEndian, &packet)
	return
}
func enet_packet_header_raw_encode(hdr enet_packet_header) (raw enet_packet_raw_header) {
	raw.pid = (hdr.pid & enet_peerid_mask) | enet_peerid_flag_time
	raw.cmd = hdr.cmd | hdr.is_unseq | hdr.need_ack
	raw.snt_time = unixtime_nowui16()
	raw.cid = hdr.cid
	raw.sn = hdr.sn
	return raw
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
