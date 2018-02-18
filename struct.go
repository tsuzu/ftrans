package main

// Message communicated between signaling server and client
type Message struct {
	LocalDescription string
	AuthCode         string
	IsServer         bool
}

// Handshake is the first message while communicating between signaling server and client
type Handshake struct {
	Version string
	ID      string
}
