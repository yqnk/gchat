package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yqnk/gchat/pkg/message"
)

const gap = "\n\n"

// NOTE: senderStyle is not required if we dont want to add BackgroundStyles or any other kind of specific styles
type model struct {
	conn        net.Conn
	username    string
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []string
	stringStyle string
	// senderStyle lipgloss.Style
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

	m := initialModel(conn, username)

	join := &message.Message{Type: "join", Sender: username, SenderStyle: m.stringStyle, Body: ""}
	conn.Write([]byte(serialize(join) + "\n"))

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

	randColor := fmt.Sprintf("#%.2x%.2x%.2x", rand.IntN(256), rand.IntN(256), rand.IntN(256))

	return model{
		conn:        conn,
		username:    username,
		textarea:    ta,
		viewport:    vp,
		messages:    []string{},
		stringStyle: randColor,
		// senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color(randColor)),
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
				if isCommand(text) {
					line := strings.Split(text[1:], " ")
					m.handleCommand(line[0], line[1:])
				} else {
					msg := &message.Message{
						Type:        "message",
						Sender:      m.username,
						SenderStyle: m.stringStyle,
						Body:        text,
					}
					m.conn.Write([]byte(serialize(msg) + "\n"))

					var style lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.stringStyle))
					m.messages = append(m.messages, style.Render("You: ")+text)
				}
				m.textarea.Reset()
				m.updateViewport()
			}
		}

	case string:
		var mes message.Message
		err := json.Unmarshal([]byte(msg), &mes)
		if err != nil {
			panic(err)
		}

		m.handleMessage(mes)
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
	leave := &message.Message{Type: "leave", Sender: m.username, SenderStyle: m.stringStyle, Body: ""}
	m.conn.Write([]byte(serialize(leave) + "\n"))
}

func (m *model) handleMessage(mes message.Message) {
	switch mes.Type {
	case "private_message":
		var style lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(mes.SenderStyle)).Italic(true)
		m.messages = append(m.messages, style.Render(mes.Sender+" whispers to you: ")+mes.Body)
	default: // message
		var style lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(mes.SenderStyle))
		m.messages = append(m.messages, style.Render(mes.Sender+": ")+mes.Body)
	}
}

func isCommand(text string) bool {
	return text[0] == 47 // 47 = "/"
}

func (m *model) handleCommand(cmd string, args []string) {

	switch cmd {
	case "color":
		if len(args) > 1 { // TODO: add check if the arg has the correct format
			m.messages = append(m.messages, "`color` command takes only one argument.")
		} else {
			m.stringStyle = args[0]
			var style lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.stringStyle))
			m.messages = append(m.messages, "`color` changed to "+style.Render(m.stringStyle))
		}
	case "w", "whisper", "msg":
		// TODO: check args
		text := args[1:]
		msg := &message.Message{
			Type:        "private_message",
			Sender:      m.username,
			SenderStyle: m.stringStyle,
			Body:        strings.Join(text, " "),
			Command:     cmd,
			Args:        args,
		}
		m.conn.Write([]byte(serialize(msg) + "\n"))
	default: // unknown command
		m.messages = append(m.messages, "Unknown command: "+cmd)
	}
}

// this should obviously be moved to pkg/message
func serialize(msg *message.Message) string {
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return string(data)
}
