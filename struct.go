package main

type Message struct {
	LocalDescription string
	AuthCode         string
	IsServer         bool
}

type Handshake struct {
	Version string
	ID      string
}

type Subscription struct {
	ID string

	Reciever       chan Message
	SenderReciever chan chan Message
}
