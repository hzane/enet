package enet

import "fmt"

var (
	enet_err_unsupported_flags   = enet_error("unsupport flag")
	enet_err_not_implemented     = enet_error("not implemented")
	enet_err_invalid_status      = enet_error("invalid status")
	enet_err_invalid_packet_size = enet_error("invalid packet size: %v")
	enet_err_assert              = enet_error("assert false")
)

type enet_error string

func assert(v bool) {
	if !v {
		panic(enet_err_assert.f())
	}
}
func assure(v bool, format string, a ...interface{}) {
	if !v {
		fmt.Printf(format, a)
	}
}
func enet_panic_error(format string, a ...interface{}) {
	panic(fmt.Errorf(format, a))
}

func (err enet_error) f(a ...interface{}) error {
	return fmt.Errorf(string(err), a...)
}

var enable_debug bool = true

func debugf(format string, a ...interface{}) {
	if enable_debug {
		fmt.Printf(format, a...)
	}
}
