package main

import "github.com/gorilla/websocket"

const (
	ProtocolVersion1_0 = "1.0"
	ProtocolVersion1_1 = "1.1" // Update following go-easyp2p
	ProtocolVersion1_2 = "1.2" // Update following go-easyp2p

	ProtocolVersionLatest = ProtocolVersion1_2
)

var (
	binaryVersion  = "unknown"
	binaryRevision = "unknown"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
