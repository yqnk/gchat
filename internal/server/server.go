package server

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	host    string
	port    string
	clients []*Client
}

func New(host string, port string) *Server {
	return &Server{host: host, port: port}
}

var wg = sync.WaitGroup{}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.host, server.port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		client := &Client{
			conn:   conn,
			server: server,
		}
		server.clients = append(server.clients, client)
		fmt.Printf("added 1 client (%v)\n", client)

		go client.handleRequest()
	}
}

func (server *Server) Broadcast(message string, sender *Client) {
	for _, client := range server.clients {
		if client != sender {
			client.conn.Write([]byte(message))
		}
	}
}
