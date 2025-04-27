package client

import (
	"bufio"
	"fmt"
	"net"
	"os"

	m "github.com/yqnk/gchat/pkg/message"
)

type Client struct {
	username string
	conn     net.Conn
}

func New(username string, address string) *Client {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	return &Client{
		username: username,
		conn:     conn,
	}
}

func (client *Client) Run() {
	defer client.conn.Close()

	joinBody := fmt.Sprintf("%s joined the room!\n", client.username)
	joinMessage := m.New(m.SystemMessage, client.username, joinBody)
	_, err := client.conn.Write([]byte(m.Serialize(*joinMessage) + "\n"))
	if err != nil {
		return
	}

	go client.Receive()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		_, err := client.conn.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			panic(err)
		}
	}
}

func (client *Client) Receive() {
	reader := bufio.NewReader(client.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			client.conn.Close()
			return
		}

		fmt.Print(message)
	}
}
