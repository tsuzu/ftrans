package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	easyp2p "github.com/cs3238-tsuzu/go-easyp2p"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/xtaci/smux"
	pb "gopkg.in/cheggaaa/pb.v1"
)

func runClient(isReceiver bool, pass string, paths, stuns []string, signaling string) error {
	err := func() error {
		if !isReceiver && len(paths) == 0 {
			log.Println("To look at help of \"ftrans send\", execute \"ftrans send -h\"")
			return errors.New("Insufficient paths")
		}

		m := make(map[string]struct{})
		for i := range paths {
			if _, ok := m[filepath.Base(paths[i])]; ok {
				return errors.New("Duplicated filename")
			}
			m[filepath.Base(paths[i])] = struct{}{}
		}

		return nil
	}()

	if err != nil {
		return err
	}

	if len(pass) == 0 {
		pass = randomSecurePassword()

		log.Println("Your password:", pass)
	}

	header := http.Header(make(map[string][]string))
	header.Add(protocolVersionHeaderKey, protocolVersionLatest)

	conn, r, err := dialer.Dial(signaling, header)

	if err != nil {
		if r != nil && r.Body != nil {
			b, e := ioutil.ReadAll(r.Body)

			r.Body.Close()
			if e == nil && len(b) != 0 {
				err = fmt.Errorf("%s: %s", err.Error(), string(b))
			}
		}
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

	if err := conn.WriteJSON(Handshake2_0{Pass: pass, Version: protocolVersionLatest}); err != nil {
		return err
	}

	log.Println("Connected to signaling server.")

	var resp string
	if err := conn.ReadJSON(&resp); err != nil {
		return err
	}

	if resp != "CONNECTED" {
		return errors.New("error: " + resp)
	}
	log.Println("Connecting to peer started.")

	p2p := easyp2p.NewP2PConn(stuns)

	defer p2p.Close()

	if _, err := p2p.Listen(0); err != nil {
		return err
	}

	if ok, err := p2p.DiscoverIP(); err != nil {
		if !ok {
			return err
		}

		v := askYesNo(func() {
			log.Println("IP addresses discovery failed:", err)
			log.Println("Available IP addresses:", strings.Join(p2p.LocalAddresses, ", "))
			fmt.Print("Continue?")
		}, nil)

		if !v {
			return nil
		}
	}

	uuid, err := uuid.NewV4()

	if err != nil {
		return err
	}
	desc, err := p2p.LocalDescription()

	if err != nil {
		return err
	}

	if err := conn.WriteJSON(Message2_0{
		IsReceiver:       isReceiver,
		LocalDescription: desc,
		AuthCode:         uuid.String(),
	}); err != nil {
		return err
	}

	var msg Message2_0
	if err := conn.ReadJSON(&msg); err != nil {
		return err
	}

	conn.Close()

	if isReceiver == msg.IsReceiver {
		var m string
		if isReceiver {
			m = "receivers"
		} else {
			m = "senders"
		}
		return errors.New("The mode is duplicating(Both are " + m + ")")
	}

	log.Println("local description:", desc)
	log.Println("remote description:", msg.LocalDescription)
	if isReceiver {
		log.Println("mode: receiver")
	} else {
		log.Println("mode: sender")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	if err := p2p.Connect(ctx, msg.LocalDescription); err != nil {
		cancel()

		return err
	}
	cancel()

	log.Println("Connected:", p2p.ReliableConn.RemoteAddr())

	type AuthMessage struct {
		FileNames []string
		FileSizes []int64
		AuthCode  string
	}

	if isReceiver {
		server, err := smux.Server(p2p, nil)

		if err != nil {
			return err
		}

		defer server.Close()

		var message AuthMessage

		err = func() error {
			stream, err := server.AcceptStream()

			if err != nil {
				return err
			}

			defer stream.Close()

			if err := json.NewDecoder(stream).Decode(&message); err != nil {
				return err
			}

			if err := json.NewEncoder(stream).Encode(msg.AuthCode); err != nil {
				return err
			}

			if message.AuthCode != uuid.String() {
				return errors.New("Unauthorized")
			}

			return nil
		}()

		if err != nil {
			return err
		}

		for _, name := range message.FileNames {
			if filepath.Dir(filepath.Clean(name)) != "." {
				return errors.New("Invalid path(avoiding security risk): " + name)
			}
		}

		for idx, name := range message.FileNames {
			err = func() error {
				stream, err := server.AcceptStream()

				if err != nil {
					return err
				}

				defer stream.Close()

				if fp, err := os.Open(name); err == nil {
					fp.Close()

					v := askYesNo(func() {
						log.Println("File already exists:", name)
						fmt.Print("Skip this?")
					}, nil)

					if v {
						return nil
					}
				}

				var fp *os.File
				if fp, err = os.Create(name); err != nil {
					return err
				}
				defer fp.Close()

				var p *pb.ProgressBar
				var writer io.Writer
				if len(message.FileSizes) > idx {
					p = pb.New64(message.FileSizes[idx]).SetUnits(pb.U_BYTES)

					p.BarStart = name
					p.Start()
					writer = io.MultiWriter(fp, p)
				} else {
					writer = fp
					log.Println(name)
				}
				_, err = io.Copy(writer, stream)

				if p != nil {
					p.Finish()
				}

				if err != nil {
					log.Printf("error: %s(%s)", name, err.Error())
				}

				return nil
			}()

			if err != nil {
				return err
			}
		}
	} else {
		client, err := smux.Client(p2p, nil)

		if err != nil {
			return err
		}

		defer client.Close()

		fileNames := make([]string, len(paths))
		fileSizes := make([]int64, len(paths))
		for i := range paths {
			paths[i] = filepath.Clean(paths[i])

			var fp *os.File
			if fp, err = os.Open(paths[i]); err != nil {
				return errors.New(paths[i] + ": " + err.Error())
			}

			stat, err := fp.Stat()
			fp.Close()
			if err != nil {
				return err
			}
			fileSizes[i] = stat.Size()
			fileNames[i] = filepath.Base(paths[i])
		}

		err = func() error {
			stream, err := client.OpenStream()

			if err != nil {
				return err
			}
			defer stream.Close()

			if err := json.NewEncoder(stream).Encode(AuthMessage{
				FileNames: fileNames,
				FileSizes: fileSizes,
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

			return nil
		}()

		if err != nil {
			return err
		}

		for idx, path := range paths {
			err = func() error {
				stream, err := client.OpenStream()

				if err != nil {
					return err
				}

				defer stream.Close()

				var fp *os.File
				if fp, err = os.Open(path); err != nil {
					return err
				}

				defer fp.Close()

				stat, err := fp.Stat()
				if err != nil {
					return err
				}

				p := pb.New64(stat.Size()).SetUnits(pb.U_BYTES)

				p.BarStart = fileNames[idx]
				p.Start()

				_, err = io.Copy(io.MultiWriter(stream, p), fp)

				p.Finish()
				if err != nil {
					log.Printf("error: %s(%s)", path, err.Error())
				}

				return nil
			}()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
