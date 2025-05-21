package main

import (
	"flag"
	"log"
	"os"

	"github.com/yqnk/gchat/internal/server"
)

func main() {
	host := flag.String("host", "0.0.0.0", "IP address of the server")
	port := flag.String("port", "3000", "Server port")

	flag.Parse()

	addr := *host + ":" + *port

	s := server.New(addr)
	if err := s.Start(); err != nil {
		log.Println("Server error: ", err)
		os.Exit(1)
	}
}
