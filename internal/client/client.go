package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

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
	defer client.Disconnect()

	joinBody := fmt.Sprintf("%s joined the room!\n", client.username)
	joinMessage := m.New(m.SystemMessage, client.username, joinBody)
	_, err := client.conn.Write([]byte(m.Serialize(*joinMessage) + "\n"))
	if err != nil {
		return
	}

	go client.Receive()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if len(scanner.Text()) > 140 {
			// TODO: show where the limit is, just like Rust's pretty errors
			// e.g. : ... this is too long ...
			// 140th character ---^
			fmt.Print("\n\t--- Max 140 characters ---\n\n")
			continue
		}

		// TODO: handle commands
		// TODO: handle private message (like commands ?)

		// avoid empty messages
		if scanner.Text() != "" {
			var message *m.Message
			if strings.HasPrefix(scanner.Text(), "/") {
				message = m.New(m.CommandMessage, client.username, scanner.Text())
			} else {
				message = m.New(m.PublicMessage, client.username, scanner.Text())
			}

			_, err := client.conn.Write([]byte(m.Serialize(*message) + "\n"))
			if err != nil {
				panic(err)
			}
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

		jsonData := m.Deserialize(message)

		if jsonData.MType == m.SystemMessage {
			fmt.Println(jsonData.Body)
		} else {
			fmt.Printf("%s: %s\n", jsonData.Author, jsonData.Body)
		}
	}
}

func (client *Client) Disconnect() {
	defer client.conn.Close()

	disconnectMessage := m.New(m.CommandMessage, client.username, "/quit")
	_, err := client.conn.Write([]byte(m.Serialize(*disconnectMessage) + "\n"))

	if err != nil {
		panic(err)
	}
}
