package main

import (
	"github.com/nfisher/nraft/server"
	"log"
	"net/http"
)

func main() {
	log.Println("starting server")
	srv := &server.Raft{}
	log.Println(http.ListenAndServe(":8080", server.Mux(srv)))
}
