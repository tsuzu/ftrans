package main

import "github.com/gorilla/websocket"

const (
	Version1_0 = "1.0"
	Version1_1 = "1.1" // Update following go-easyp2p

	VersionLatest = Version1_1
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
