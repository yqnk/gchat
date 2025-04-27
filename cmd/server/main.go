package main

import (
	"fmt"

	s "github.com/yqnk/gchat/internal/server"
)

func main() {
	server := s.New("localhost", "3333")
	fmt.Printf("Starting on %s:%s...\n", "localhost", "3333")
	server.Run()
}
