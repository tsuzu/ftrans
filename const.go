package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	protocolVersion1_0 = "1.0"
	protocolVersion1_1 = "1.1" // Update following go-easyp2p
	protocolVersion1_2 = "1.2"
	protocolVersion2_0 = "2.0"

	protocolVersionLatest = protocolVersion2_0
)

var protocolVersionArray = []string{
	protocolVersion1_0,
	protocolVersion1_1,
	protocolVersion1_2,
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

const protocolVersionHeaderKey = "X-Ftrans-Protocol-Version"

const defaultSignalingServer = "wss://ftrans.cs3238.com/ws"
const defaultSTUNServer = "stun.l.google.com:19302"
