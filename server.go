package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func worker(connChan <-chan *websocket.Conn, conn1 *websocket.Conn, canceller func()) {
	defer canceller()
	defer conn1.Close()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C

			if err := conn1.WriteControl(websocket.PingMessage, []byte("keep-alive"), time.Now().Add(3*time.Second)); err != nil {
				canceller()

				return
			}
		}
	}()

	var ok bool
	var conn2 *websocket.Conn
	if conn2, ok = <-connChan; !ok {
		return
	}

	defer conn2.Close()

	conn1.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn2.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn1.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn2.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err := conn1.WriteJSON("CONNECTED"); err != nil {
		return
	}
	if err := conn2.WriteJSON("CONNECTED"); err != nil {
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)

	readWrite := func(a, b *websocket.Conn) {
		var msg Message
		if err := a.ReadJSON(&msg); err != nil {
			return
		}

		b.WriteJSON(msg)
	}

	wg.Add(1)
	go func() {
		readWrite(conn1, conn2)
		wg.Done()
	}()

	readWrite(conn2, conn1)

	wg.Wait()

}

func initServer() http.Handler {
	var mut sync.Mutex
	conns := make(map[string]chan<- *websocket.Conn)

	router := mux.NewRouter()

	router.HandleFunc("/ws", func(rw http.ResponseWriter, req *http.Request) {
		log.Println("connected", req.RemoteAddr)
		defer log.Println("closed", req.RemoteAddr)
		conn, err := upgrader.Upgrade(rw, req, nil)

		if err != nil {
			log.Println(err)

			return
		}

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
		ch, ok := conns[id]

		if !ok {
			ch := make(chan *websocket.Conn, 1)
			conns[id] = ch

			go worker(ch, conn, func() {
				mut.Lock()
				if c, ok := conns[id]; ok && c == ch {
					delete(conns, id)

					close(ch)
				}
				mut.Unlock()
			})
		} else {
			delete(conns, id)

			ch <- conn
		}

		mut.Unlock()

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
