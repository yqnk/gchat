package message

type Message struct {
	Type        string `json:"type"`
	Sender      string `json:"sender"`
	Body        string `json:"body"`
	SenderStyle string `json:"sender_style"`
}
