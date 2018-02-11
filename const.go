package main

import "github.com/gorilla/websocket"

const (
	Version1_0 = "1.0"

	VersionLatest = Version1_0
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
