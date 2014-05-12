package enet

import "time"

func unixtime_now() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func time2i64(t time.Time) int64 {
	n := t.UnixNano()
	return n / int64(time.Millisecond)
}
func time2ui32(t time.Time) uint32 {
	return uint32(time2i64(t))
}

func betweenui32(a, r, l uint32) uint32 {
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

func bool2int(v bool) int {
	if v {
		return 1
	}
	return 0
}
