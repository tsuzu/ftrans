package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	easyp2p "git.mox.si/tsuzu/go-easyp2p"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/xtaci/smux"
	pb "gopkg.in/cheggaaa/pb.v1"
)

func runClient() {
	if len(os.Args) < 2 {
		flag.Usage()

		return
	}
	id := os.Args[1]
	paths := os.Args[2:]

	stuns := strings.Split(*stun, ",")
	for i := range stuns {
		stuns[i] = strings.TrimPrefix(stuns[i], " ")
	}

	func() {
		m := make(map[string]struct{})
		for i := range paths {
			if _, ok := m[filepath.Base(paths[i])]; ok {
				panic("Duplicated filename")
			}
			m[filepath.Base(paths[i])] = struct{}{}
		}
	}()

	conn, _, err := websocket.DefaultDialer.Dial(*signaling, nil)

	// keep-alive
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			<-ticker.C

			if err := conn.WriteControl(websocket.PingMessage, []byte("keep-alive"), time.Now().Add(10*time.Second)); err != nil {
				log.Println("error:", err)
			}
		}
	}()

	if err != nil {
		panic(err)
	}

	if err := conn.WriteJSON(Handshake{ID: id, Version: VersionLatest}); err != nil {
		panic(err)
	}

	log.Println("Connected to signaling server.")

	isServer := false
	if len(paths) == 0 {
		isServer = true
	}

	var resp string
	if err := conn.ReadJSON(&resp); err != nil {
		panic(err)
	}

	if resp != "CONNECTED" {
		panic("error: " + resp)
	}
	log.Println("Connecting started.")

	p2p := easyp2p.NewP2PConn(stuns, easyp2p.DiscoverIPWithSTUN)

	if _, err := p2p.Listen(0); err != nil {
		conn.Close()

		panic(err)
	}

	if ok, err := p2p.DiscoverIP(); err != nil {
		if !ok {
			panic(err)
		} else {
			log.Println("IP addresses discovery failed: ", err)
			fmt.Print("Continue?(y/n): ")

			var c string
			fmt.Scan(&c)

			if c != "Y" && c != "y" {
				conn.Close()

				return
			}
		}
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	desc, err := p2p.LocalDescription()

	if err != nil {
		panic(err)
	}

	if err := conn.WriteJSON(Message{
		IsServer:         isServer,
		LocalDescription: desc,
		AuthCode:         uuid.String(),
	}); err != nil {
		conn.Close()

		panic(err)
	}

	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		conn.Close()

		panic(err)
	}

	conn.Close()

	if isServer == msg.IsServer {
		log.Println("error: The mode is duplicating.")
		return
	}

	log.Println("local description: ", desc)
	log.Println("remote description: ", msg.LocalDescription)
	if isServer {
		log.Println("mode: reciever")
	} else {
		log.Println("mode: sender")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	if err := p2p.Connect(msg.LocalDescription, isServer, ctx); err != nil {
		panic(err)
	}
	cancel()

	log.Println("Connected: ", p2p.UTPConn.RemoteAddr())

	type AuthMessage struct {
		Filenames []string
		AuthCode  string
	}

	if isServer {
		server, err := smux.Server(p2p, nil)

		if err != nil {
			p2p.Close()
			panic(err)
		}

		stream, err := server.AcceptStream()

		if err != nil {
			p2p.Close()
			panic(err)
		}

		var message AuthMessage

		if err := json.NewDecoder(stream).Decode(&message); err != nil {
			p2p.Close()
			panic(err)
		}

		if err := json.NewEncoder(stream).Encode(msg.AuthCode); err != nil {
			p2p.Close()
			panic(err)
		}

		if message.AuthCode != uuid.String() {
			log.Println("Unauthorized")

			p2p.Close()
			return
		}

		stream.Close()

		for _, name := range message.Filenames {
			stream, err := server.AcceptStream()

			if err != nil {
				p2p.Close()
				panic(err)
			}
			var fp *os.File
			if fp, err = os.Create(name); err != nil {
				p2p.Close()
				panic(err)
			}

			if _, err := io.Copy(fp, stream); err != nil {
				log.Println(err)
			}

			fp.Close()
			stream.Close()
		}
	} else {
		client, err := smux.Client(p2p, nil)

		if err != nil {
			p2p.Close()
			panic(err)
		}

		stream, err := client.OpenStream()

		if err != nil {
			p2p.Close()
			panic(err)
		}

		filenames := make([]string, len(paths))
		for i := range paths {
			filenames[i] = filepath.Base(paths[i])
		}

		if err := json.NewEncoder(stream).Encode(AuthMessage{
			Filenames: filenames,
			AuthCode:  msg.AuthCode,
		}); err != nil {
			p2p.Close()
			panic(err)
		}

		var auth string
		if err := json.NewDecoder(stream).Decode(&auth); err != nil {
			p2p.Close()
			panic(err)
		}

		if auth != uuid.String() {
			log.Println("Unauthorized")

			p2p.Close()
			return
		}

		stream.Close()

		for idx, path := range paths {
			stream, err := client.OpenStream()

			if err != nil {
				p2p.Close()
				panic(err)
			}
			var fp *os.File
			if fp, err = os.Open(path); err != nil {
				p2p.Close()
				panic(err)
			}
			stat, err := fp.Stat()
			if err != nil {
				p2p.Close()
				panic(err)
			}

			p := pb.New64(stat.Size()).SetUnits(pb.U_BYTES)

			p.BarStart = filenames[idx]
			p.Start()

			if _, err := io.Copy(io.MultiWriter(stream, p), fp); err != nil {
				log.Println("copy", err)
			}

			p.Finish()

			fp.Close()
			stream.Close()
		}
	}
}
