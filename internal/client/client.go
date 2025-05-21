package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yqnk/gchat/pkg/message"
)

const gap = "\n\n"

type model struct {
	conn        net.Conn
	username    string
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []string
	senderStyle lipgloss.Style
	receiveChan chan string
	width       int
	height      int
	initialized bool
}

func New(username, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	join := &message.Message{Type: "join", Sender: username, Body: ""}
	conn.Write([]byte(serialize(join) + "\n"))

	m := initialModel(conn, username)

	go receiveLoop(conn, m.receiveChan)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func initialModel(conn net.Conn, username string) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.CharLimit = 140
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.Prompt = "┃ "

	vp := viewport.New(40, 10)
	vp.SetContent("Welcome to the chat room!\nType a message and press Enter to send.")

	return model{
		conn:        conn,
		username:    username,
		textarea:    ta,
		viewport:    vp,
		messages:    []string{},
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiveChan: make(chan string),
		initialized: false,
	}
}

func receiveLoop(conn net.Conn, ch chan string) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			close(ch)
			return
		}
		ch <- strings.TrimSpace(msg)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForMessage())
}

func (m model) waitForMessage() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.receiveChan
		if !ok {
			return tea.Quit()
		}
		return string(msg)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var taCmd tea.Cmd
	m.textarea, taCmd = m.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		m.textarea.SetWidth(msg.Width)
		m.initialized = true
		m.updateViewport()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.sendLeaveMessage()
			return m, tea.Quit
		case tea.KeyEnter:
			text := strings.TrimSpace(m.textarea.Value())
			if text == "" {
				return m, nil
			}
			if len(text) > 140 {
				m.messages = append(m.messages, "-- Max 140 characters --")
			} else {
				msg := &message.Message{
					Type:   "message",
					Sender: m.username,
					Body:   text,
				}
				m.conn.Write([]byte(serialize(msg) + "\n"))
				m.messages = append(m.messages, m.senderStyle.Render("You: ")+text)
			}
			m.textarea.Reset()
			m.updateViewport()
		}

	case string:
		var mes message.Message
		err := json.Unmarshal([]byte(msg), &mes)
		if err != nil {
			panic(err)
		}

		m.messages = append(m.messages, m.senderStyle.Render(mes.Sender+": ")+mes.Body)
		m.updateViewport()
		return m, m.waitForMessage()
	}

	return m, taCmd
}

func (m *model) updateViewport() {
	content := lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n"))
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m model) View() string {
	if !m.initialized {
		return "Loading..."
	}
	return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
}

func (m model) sendLeaveMessage() {
	leave := &message.Message{Type: "leave", Sender: m.username, Body: ""}
	m.conn.Write([]byte(serialize(leave) + "\n"))
}

func serialize(msg *message.Message) string {
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return string(data)
}
