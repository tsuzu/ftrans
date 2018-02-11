package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	mode      = flag.String("mode", "client", "Server(signaling server) or client(sender or reciever)")
	stun      = flag.String("stun", "stun.l.google.com:19302", "STUN server addresses(split with ',')")
	signaling = flag.String("sig", "wss://ftrans.cs3238.com/ws", "Signaling server address")
)

func main() {
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of ftrans:")
		fmt.Fprintln(os.Stderr, "  ftrans [options] password [file paths...]")
		fmt.Fprintln(os.Stderr, "  If no path is passed, this runs as a reciever.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "  To launch a signaling server: ftrans --mode server")
		fmt.Fprintln(os.Stderr)

		flag.PrintDefaults()
	}

	if *mode == "server" {
		runServer()
	} else {
		runClient()
	}
}
