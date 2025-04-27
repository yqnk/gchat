package main

import (
	"flag"
	"fmt"
	gclient "github.com/yqnk/gchat/internal/client"
)

func main() {
	var host string
	flag.StringVar(&host, "host", "localhost", "a string")

	var port string
	flag.StringVar(&port, "port", "3333", "a string")

	var username string
	flag.StringVar(&username, "name", "unnamed", "a string")

	flag.Parse()

	fmt.Printf("Logging to %s:%s as %s...\n", host, port, username)

	client := gclient.New(username, fmt.Sprintf("%s:%s", host, port))
	client.Run()
}
