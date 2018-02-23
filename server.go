package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func initServer() http.Handler {
	// Initialize handlers
	init1_1()
	init1_2()

	router := mux.NewRouter()

	redirect := func(rw http.ResponseWriter, version string) {
		rw.Header().Add("Location", "./ws/"+version)
		rw.WriteHeader(http.StatusFound)
	}
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		redirect(rw, "https://git.mox.si/tsuzu/ftrans")
	})

	/*isolate := func(rw http.ResponseWriter) {
		rw.WriteHeader(http.StatusGone)
		rw.Write([]byte(`This protocol version is isolate`))
	}*/

	router.HandleFunc("/ws", func(rw http.ResponseWriter, req *http.Request) {
		switch req.Header.Get(ProtocolVersionHeaderKey) {
		case ProtocolVersion1_2:
			serverHandler1_2(rw, req)
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
