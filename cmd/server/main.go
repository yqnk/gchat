package main

import (
	"flag"
	"fmt"

	s "github.com/yqnk/gchat/internal/server"
)

func main() {
	var host string
	flag.StringVar(&host, "host", "localhost", "a string")

	var port string
	flag.StringVar(&port, "port", "3333", "a string")

	flag.Parse()

	server := s.New(host, port)
	fmt.Printf("Starting on %s:%s...\n", host, port)
	server.Run()
}
