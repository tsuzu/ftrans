package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	easyp2p "git.moxapp.net/tsuzu/go-easyp2p"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"github.com/xtaci/smux"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	mode      = flag.String("mode", "client", "Server(signaling server) or client(sender or reciever)")
	stun      = flag.String("stun", "stun.l.google.com:19302", "STUN server addresses(split with ',')")
	signaling = flag.String("sig", "wss://ftrans.cs3238.com/ws", "Signaling server address")
)

const (
	Version1_0 = "1.0"

	VersionLatest = Version1_0
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	LocalDescription string
	AuthCode         string
	IsServer         bool
}

type Handshake struct {
	Version string
	ID      string
}

func main() {
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of ftrans:")
		fmt.Fprintln(os.Stderr, "  ftrans [options] password [file paths...]")
		fmt.Fprintln(os.Stderr, "  If no path is passed, this runs as a reciever.")
		fmt.Fprintln(os.Stderr)

		flag.PrintDefaults()
	}

	if *mode == "server" {
		runServer()
	} else {
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

		var mut sync.Mutex
		conn, _, err := websocket.DefaultDialer.Dial(*signaling, nil)

		// keep-alive
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			for {
				<-ticker.C

				mut.Lock()
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("error:", err)
				}
				mut.Unlock()
			}
		}()

		if err != nil {
			panic(err)
		}

		mut.Lock()
		if err := conn.WriteJSON(Handshake{ID: id, Version: VersionLatest}); err != nil {
			panic(err)
		}
		mut.Unlock()

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
				log.Println("DiscoverIP fail: ", err)
				log.Println("Continue?(y/n)")

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

		mut.Lock()
		if err := conn.WriteJSON(Message{
			IsServer:         isServer,
			LocalDescription: desc,
			AuthCode:         uuid.String(),
		}); err != nil {
			conn.Close()

			panic(err)
		}
		mut.Unlock()

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
			log.Println("mode: sender")
		} else {
			log.Println("mode: reciever")
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
}

type Subscription struct {
	ID string

	Reciever       chan Message
	SenderReciever chan chan Message
}

func initServer() http.Handler {
	var mut sync.Mutex
	conns := make(map[string]Subscription)

	router := mux.NewRouter()

	router.HandleFunc("/ws", func(rw http.ResponseWriter, req *http.Request) {
		log.Println("connected", req.RemoteAddr)
		defer log.Println("closed", req.RemoteAddr)
		conn, err := upgrader.Upgrade(rw, req, nil)

		if err != nil {
			log.Println(err)

			return
		}

		defer conn.Close()

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		var hs Handshake
		if err := conn.ReadJSON(&hs); err != nil {
			log.Println(err)

			return
		}
		conn.SetReadDeadline(time.Time{})

		if hs.Version != VersionLatest {
			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			conn.WriteJSON("Incorrect Version")
			conn.Close()

			return
		}
		id := hs.ID

		log.Println(req.RemoteAddr, ": ", id)

		mut.Lock()
		subs, ok := conns[id]

		if !ok {
			ch := make(chan Message, 1)
			sch := make(chan chan Message, 1)
			subs = Subscription{
				ID:             id,
				Reciever:       ch,
				SenderReciever: sch,
			}
			conns[id] = subs
		} else {
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			delete(conns, id)
		}

		mut.Unlock()

		var sender, reciever chan Message
		if !ok {
			sender = subs.Reciever
		} else {
			sender = make(chan Message, 1)
			subs.SenderReciever <- sender
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			var msg Message
			defer wg.Done()
			defer close(sender)

			if err := conn.ReadJSON(&msg); err != nil {
				log.Println(err)

				mut.Lock()
				if _, ok := conns[id]; ok {
					delete(conns, id)
				}
				mut.Unlock()

				return
			}

			sender <- msg
		}()

		if !ok {
			func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
			FOR:
				for {
					select {
					case reciever = <-subs.SenderReciever:
						break FOR
					case <-ticker.C:
						if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
							mut.Lock()
							if _, ok := conns[id]; ok {
								delete(conns, id)
							}
							mut.Unlock()

							return
						}
					}
				}
			}()
		} else {
			reciever = subs.Reciever
		}
		if err := conn.WriteJSON("CONNECTED"); err != nil {
			return
		}

		msg, ok := <-reciever

		if ok {
			conn.WriteJSON(msg)
		}
		wg.Wait()
	})
	return router
}

func runServer() {
	router := initServer()
	listen := os.Getenv("LISTEN")

	if listen == "" {
		listen = ":80"
	}
	panic(http.ListenAndServe(listen, router))
}
