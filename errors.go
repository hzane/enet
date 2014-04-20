package enet

import "fmt"

var (
	enet_err_unsupported_flags = enet_error("enet unsupport flag")
	enet_err_not_implemented   = enet_error("enet not implemented")
	enet_err_assert            = enet_error("assert false")
)

func assert(v bool, format string, a ...interface{}) {
	if !v {
		panic(enet_error(format, a))
	}
}

func enet_panic_error(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a))
}

func enet_error(format string, a ...interface{}) error {
	return fmt.Errorf(format, a)
}

func enet_err_invalid_packet_size(sz int) error {
	return fmt.Errorf("enet invalid packet size: %v", sz)
}
