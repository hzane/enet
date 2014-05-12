package enet

import (
	"os"
)

type PeerEventHandler func(Peer, ret int)
type DataEventHandler func(Peer, chanid uint8, payload []uint8)

type Host interface {
	SetConnectionHandler(PeerEventHandler)
	SetDisconnectionHandler(PeerEventHandler)
	SetDataHandler(DataEventHandler)
	Connect(dest string)
	Run(chan os.Signal)
}

type Peer interface {
	Disconnect()
	Write(chanid uint8, dat []byte)
}
