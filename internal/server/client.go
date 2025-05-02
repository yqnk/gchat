package server

import (
	"bufio"
	"net"

	m "github.com/yqnk/gchat/pkg/message"
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

		jsonData := m.Deserialize(message)

		if jsonData.MType == m.CommandMessage {
			client.server.ExecuteCommand(message, client)
		} else {
			client.server.Broadcast(message, client)
		}
	}
}
