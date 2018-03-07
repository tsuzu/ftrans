package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	ProtocolVersion1_0 = "1.0"
	ProtocolVersion1_1 = "1.1" // Update following go-easyp2p
	ProtocolVersion1_2 = "1.2"
	ProtocolVersion2_0 = "2.0"

	ProtocolVersionLatest = ProtocolVersion2_0
)

var ProtocolVersionArray = []string{
	ProtocolVersion1_0,
	ProtocolVersion1_1,
	ProtocolVersion1_2,
}

var (
	binaryVersion  = "unknown"
	binaryRevision = "unknown"
)

var dialer = websocket.Dialer{
	HandshakeTimeout: 5 * time.Second,
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5 * time.Second,
}

const ProtocolVersionHeaderKey = "X-Ftrans-Protocol-Version"

const defaultSignalingServer = "wss://ftrans.cs3238.com/ws"
const defaultSTUNServer = "stun.l.google.com:19302"
