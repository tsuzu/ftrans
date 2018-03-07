package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func initServer() http.Handler {
	// Initialize handlers
	init1_1()
	init1_2()
	init2_0()

	router := mux.NewRouter()

	redirect := func(rw http.ResponseWriter, address string) {
		rw.Header().Add("Location", address)
		rw.WriteHeader(http.StatusFound)
	}
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		redirect(rw, "https://github.comi/cs3238-tsuzu/ftrans")
	})

	/*isolate := func(rw http.ResponseWriter) {
		rw.WriteHeader(http.StatusGone)
		rw.Write([]byte(`This protocol version is isolate`))
	}*/

	router.HandleFunc("/ws", func(rw http.ResponseWriter, req *http.Request) {
		switch req.Header.Get(protocolVersionHeaderKey) {
		case protocolVersion1_2:
			serverHandler1_2(rw, req)
		case protocolVersion2_0:
			serverHandler2_0(rw, req)
		default:
			serverHandler1_1(rw, req)
		}
	})
	return router
}

func runServer(addr string) error {
	router := initServer()

	if addr == "" {
		addr = ":80"
	}
	return http.ListenAndServe(addr, router)
}
