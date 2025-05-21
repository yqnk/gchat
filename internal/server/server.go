package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/yqnk/gchat/pkg/message"
)

type Server struct {
	Addr    string
	clients map[net.Conn]string
	mu      sync.Mutex
}

func New(addr string) *Server {
	return &Server{
		Addr:    addr,
		clients: make(map[net.Conn]string),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("[E] Listener error: %w", err)
	}
	log.Printf("[I] Server started at %s...", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[W] Connection error: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// save the client address
	addr := conn.RemoteAddr().String()
	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		log.Printf("[W] No initial message from %s", addr)
		return
	}

	var initMessage message.Message
	if err := json.Unmarshal(scanner.Bytes(), &initMessage); err != nil || initMessage.Type != "join" {
		log.Printf("[W] Invalid join message from %s (message type: %s): %v", addr, initMessage.Type, err)
		return
	}

	username := initMessage.Sender

	s.mu.Lock()
	s.clients[conn] = username
	s.mu.Unlock()

	log.Printf("[I] Client connected: %s (%s)", addr, username)
	s.broadcast(message.Message{Type: "message", Sender: "server", Body: fmt.Sprintf("%s joined the chat :)", username)})

	for scanner.Scan() {
		var msg message.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			log.Printf("[W] JSON Error from %s: %v", addr, err)
		}

		log.Printf("[%s] %s", msg.Sender, msg.Body)
		s.broadcast(msg)
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()

	log.Printf("[I] Client disconnected: %s (%s)", addr, username)
	s.broadcast(message.Message{Type: "message", Sender: "server", Body: fmt.Sprintf("%s left the chat :(", username)})
}

func (s *Server) broadcast(msg message.Message) {
	data, _ := json.Marshal(msg)
	data = append(data, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		s.handleMessage(msg, data, conn)
	}
}

func (s *Server) handleMessage(msg message.Message, data []byte, conn net.Conn) {
	log.Println(msg.Type)
	switch msg.Type {
	case "private_message":
		receiver := msg.Args[0]
		log.Println("[I] DM to " + receiver)
		if s.clients[conn] == receiver {
			if _, err := conn.Write(data); err != nil {
				log.Printf("[W] Failed to send %s: %v", s.clients[conn], err)
			}
		}
	default: // "message"
		if msg.Sender != s.clients[conn] {
			if _, err := conn.Write(data); err != nil {
				log.Printf("[W] Failed to send %s: %v", s.clients[conn], err)
			}
		}
	}
}
