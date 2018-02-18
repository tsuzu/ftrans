package main

import (
	"context"
	"encoding/json"
	"errors"
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

func runClient() error {
	if len(os.Args) < 2 {
		flag.Usage()

		return nil
	}
	id := os.Args[1]
	paths := os.Args[2:]

	stuns := strings.Split(*stun, ",")
	for i := range stuns {
		stuns[i] = strings.TrimPrefix(stuns[i], " ")
	}

	if err := func() error {
		m := make(map[string]struct{})
		for i := range paths {
			if _, ok := m[filepath.Base(paths[i])]; ok {
				return errors.New("Duplicated filename")
			}
			m[filepath.Base(paths[i])] = struct{}{}
		}

		return nil
	}(); err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(*signaling, nil)

	if err != nil {
		return err
	}

	// keep-alive
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			<-ticker.C

			if err := conn.WriteControl(websocket.PingMessage, []byte("keep-alive"), time.Now().Add(10*time.Second)); err != nil {
				conn.Close()
				return
			}
		}
	}()

	defer conn.Close()

	if err := conn.WriteJSON(Handshake{ID: id, Version: ProtocolVersionLatest}); err != nil {
		return err
	}

	log.Println("Connected to signaling server.")

	isServer := false
	if len(paths) == 0 {
		isServer = true
	}

	var resp string
	if err := conn.ReadJSON(&resp); err != nil {
		return err
	}

	if resp != "CONNECTED" {
		return errors.New("error: " + resp)
	}
	log.Println("Connecting to peer started.")

	p2p := easyp2p.NewP2PConn(stuns, easyp2p.DiscoverIPWithSTUN)

	defer p2p.Close()

	if _, err := p2p.Listen(0); err != nil {
		return err
	}

	if ok, err := p2p.DiscoverIP(); err != nil {
		if !ok {
			return err
		}

		log.Println("IP addresses discovery failed: ", err)
		log.Println("Available IP addresses: ", strings.Join(p2p.LocalAddresses, ", "))
		fmt.Print("Continue?(y/n): ")

		var c string
		fmt.Scan(&c)

		switch strings.ToLower(c) {
		case "yes":
		case "y":

		default:
			return nil
		}
	}

	uuid := uuid.NewV4()

	desc, err := p2p.LocalDescription()

	if err != nil {
		return err
	}

	if err := conn.WriteJSON(Message{
		IsServer:         isServer,
		LocalDescription: desc,
		AuthCode:         uuid.String(),
	}); err != nil {
		return err
	}

	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		return err
	}

	conn.Close()

	if isServer == msg.IsServer {
		var m string
		if isServer {
			m = "receivers"
		} else {
			m = "senders"
		}
		return errors.New("The mode is duplicating(Both are " + m + ")")
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
		cancel()

		return err
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
			return err
		}

		stream, err := server.AcceptStream()

		if err != nil {
			return err
		}

		var message AuthMessage

		if err := json.NewDecoder(stream).Decode(&message); err != nil {
			return err
		}

		if err := json.NewEncoder(stream).Encode(msg.AuthCode); err != nil {
			return err
		}

		if message.AuthCode != uuid.String() {
			return errors.New("Unauthorized")
		}

		stream.Close()

		for _, name := range message.Filenames {
			stream, err := server.AcceptStream()

			if err != nil {
				return err
			}
			var fp *os.File
			if fp, err = os.Create(name); err != nil {
				return err
			}

			if _, err := io.Copy(fp, stream); err != nil {
				log.Printf("%s: %s", name, err.Error())
			}

			fp.Close()
			stream.Close()
		}
	} else {
		client, err := smux.Client(p2p, nil)

		if err != nil {
			return err
		}

		stream, err := client.OpenStream()

		if err != nil {
			return err
		}

		filenames := make([]string, len(paths))
		for i := range paths {
			filenames[i] = filepath.Base(paths[i])
		}

		if err := json.NewEncoder(stream).Encode(AuthMessage{
			Filenames: filenames,
			AuthCode:  msg.AuthCode,
		}); err != nil {
			return err
		}

		var auth string
		if err := json.NewDecoder(stream).Decode(&auth); err != nil {
			return err
		}

		if auth != uuid.String() {
			return errors.New("Unauthorized")
		}

		stream.Close()

		for idx, path := range paths {
			stream, err := client.OpenStream()

			if err != nil {
				return err
			}
			var fp *os.File
			if fp, err = os.Open(path); err != nil {
				return err
			}
			stat, err := fp.Stat()
			if err != nil {
				return err
			}

			p := pb.New64(stat.Size()).SetUnits(pb.U_BYTES)

			p.BarStart = filenames[idx]
			p.Start()

			if _, err := io.Copy(io.MultiWriter(stream, p), fp); err != nil {
				log.Printf("%s: %s", path, err.Error())
			}

			p.Finish()

			fp.Close()
			stream.Close()
		}
	}

	return nil
}
