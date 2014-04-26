package enet

import "io"

type enet_reader interface {
	io.Reader
	left() int
	byte() byte
	uint16() uint16
	uint32() uint32
	uint64() uint64
	bytes() []byte
}

func enet_reader_new(data []byte) enet_reader {
	return &enet_reader_{0, data}
}

type enet_reader_ struct {
	pointer int
	data    []byte
}

func (reader *enet_reader_) left() int {
	return len(reader.data) - reader.pointer
}

func (reader *enet_reader_) byte() byte {
	idx := reader.pointer
	reader.pointer++
	v := reader.data[idx]
	return v
}

func (reader *enet_reader_) uint16() uint16 {
	b0 := uint16(reader.byte())
	b1 := uint16(reader.byte())
	return b0<<8 + b1
}

func (reader *enet_reader_) uint32() uint32 {
	s0 := uint32(reader.uint16())
	s1 := uint32(reader.uint16())
	return s0<<16 + s1
}

func (reader *enet_reader_) uint64() uint64 {
	i0 := uint64(reader.uint32())
	i1 := uint64(reader.uint32())
	return i0<<32 + i1
}

func (reader *enet_reader_) bytes() []byte {
	v := reader.data[reader.pointer:]
	reader.pointer = len(reader.data)
	return v
}
func (reader *enet_reader_) Read(p []byte) (n int, err error) {
	n = copy(p, reader.bytes())
	return
}
