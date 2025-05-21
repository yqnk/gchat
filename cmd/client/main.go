package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yqnk/gchat/internal/client"
)

func main() {
	host := flag.String("host", "localhost", "IP address of the server")
	port := flag.String("port", "3000", "Server port")
	// add a randomly generated uuid as default username to allow the absence of --name ?
	username := flag.String("name", "", "User name")

	flag.Parse()

	if *username == "" {
		fmt.Println("Error: --name is required")
		flag.Usage()
		os.Exit(1)
	}

	addr := *host + ":" + *port

	client.New(*username, addr)
}
