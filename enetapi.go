package enet

import (
	"os"
)

type PeerEventHandler func(host Host, endpoint string, ret int)

// chanid == 0xff unreliable data
type DataEventHandler func(host Host, endpoint string, chanid uint8, payload []byte)

// push a signal to chan os.Signal will make host quit run loop
type Host interface {
	SetConnectionHandler(PeerEventHandler)
	SetDisconnectionHandler(PeerEventHandler)
	SetDataHandler(DataEventHandler)
	Connect(endpoint string)
	Disconnect(endpoint string)
	Write(endpoint string, chanid uint8, dat []byte)
	Run(chan os.Signal)
	Stop()
}

// func NewHost(addr string) (Host, error)
/*
type Peer interface {
	Disconnect()                    // must be called in host-run routine
	Write(chanid uint8, dat []byte) // be carefull, must be called in host-run routine
	Addr() net.Addr
}*/
