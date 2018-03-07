package main

// Message1_1 communicated between signaling server and client for ProtocolVersion1_1
type Message1_1 struct {
	LocalDescription string
	AuthCode         string
	IsServer         bool
}

// Message1_2 communicated between signaling server and client for ProtocolVersion1_2
type Message1_2 struct {
	LocalDescription string
	AuthCode         string
	IsReceiver       bool
}

// Message2_0 communicated between signaling server and client for ProtocolVersion2_0
type Message2_0 struct {
	LocalDescription string
	AuthCode         string
	IsReceiver       bool
}

// Handshake1_1 is the first message while communicating between signaling server and client for ProtocolVersion1_1
type Handshake1_1 struct {
	Version string
	ID      string
}

// Handshake1_2 is the first message while communicating between signaling server and client for ProtocolVersion1_2
type Handshake1_2 struct {
	Version string
	Pass    string
}

// Handshake2_0 is the first message while communicating between signaling server and client for ProtocolVersion2_0
type Handshake2_0 struct {
	Version string
	Pass    string
}
