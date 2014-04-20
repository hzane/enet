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
