package server

import (
	"fmt"
	"log"
	"net"

	m "github.com/yqnk/gchat/pkg/message"
)

type Server struct {
	host    string
	port    string
	clients []*Client
}

func New(host string, port string) *Server {
	return &Server{host: host, port: port}
}

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

		go client.handleRequest()
	}
}

func (server *Server) Broadcast(message string, sender *Client) {
	jsonData := m.Deserialize(message)
	if jsonData.MType == m.SystemMessage {
		if jsonData.Body == "/quit" {
			// TODO: find a way to remove the corresponding client from the server's `clients` list
			// hint: use a map, and mutex rwlock immediately
		}

		// maybe use a command instead of a system message, like /join
		fmt.Printf("[SYSTEM by %s] %s", jsonData.Author, jsonData.Body)
	}

	for _, client := range server.clients {
		if jsonData.MType == m.SystemMessage {
			client.conn.Write([]byte(jsonData.Body))
		} else {
			if client != sender {
				client.conn.Write([]byte(jsonData.Author + " > " + jsonData.Body + "\n"))
			}
		}
	}
}
