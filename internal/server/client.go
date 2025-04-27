package server

import (
	"bufio"
	"net"
)

type Client struct {
	conn   net.Conn
	server *Server
}

func (client *Client) handleRequest() {
	reader := bufio.NewReader(client.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			client.conn.Close()
			return
		}

		client.server.Broadcast(message, client)
	}
}
