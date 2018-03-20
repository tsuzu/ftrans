package main

// Message2_0 communicated between signaling server and client for ProtocolVersion2_0
type Message2_0 struct {
	LocalDescription string
	AuthCode         string
	IsReceiver       bool
}

// Handshake2_0 is the first message while communicating between signaling server and client for ProtocolVersion2_0
type Handshake2_0 struct {
	Version string
	Pass    string
}
