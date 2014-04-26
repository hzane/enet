package enet

import (
	"fmt"
	"time"
)

func print(format string, a ...interface{}) {
	fmt.Printf(format, a)
}

func enet_on_connected(Host, Peer, uint) {
	print("on-connected")
}
func enet_on_disconnected(Host, Peer, uint) {
	print("on-disconnected")
}
func enet_on_reliable(Host, Peer, uint8, []uint8) {
	print("on-reliable")
}
func enet_on_unreliable(Host, Peer, uint8, []uint8) {
	print("on-unreliable")
}
func enet_on_unsequenced(Host, Peer, uint8, []uint8) {
	print("on-unsequenced")
}

func unixtime_nowui16() uint16 {
	return uint16(unixtime_now())
}

func unixtime_now() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
func unixtime_fromui16(t uint16, now int64) int64 {
	it := int64(t)
	v := it | now&0xffff0000
	if it&0x8000 > now&0x8000 {
		v -= 0x1000000
	}
	return v
}
func unixtimei64(t time.Time) int64 {
	n := t.UnixNano()
	return n / int64(time.Millisecond)
}
func unixtimeui32(t time.Time) uint32 {
	return uint32(unixtimei64(t))
}
func unixtimeui16(t time.Time) uint16 {
	return uint16(unixtimei64(t))
}
func dampui32(a, r, l uint32) uint32 {
	if a > l {
		a = l
	}
	if a < r {
		a = r
	}
	return a
}

func minui32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
func absi64(v int64) int64 {
	if v > 0 {
		return v
	}
	return -v
}
func mini64(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}
func maxi64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func maxui32(a, b uint32) uint32 {
	if a > b {
		return b
	}
	return a
}

func bool_int(v bool) int {
	if v {
		return 1
	}
	return 0
}
