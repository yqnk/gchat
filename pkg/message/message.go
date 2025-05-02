package message

import "encoding/json"

type MessageType int

const (
	SystemMessage  MessageType = iota // broadcasted towards every client
	PublicMessage                     // broadcasted towards everyone but the sender
	CommandMessage                    // broadcasted towards the sender only
)

type Message struct {
	MType  MessageType `json:"type"`
	Author string      `json:"author"`
	Body   string      `json:"body"`
}

func New(broadcastType MessageType, author string, body string) *Message {
	return &Message{
		MType:  broadcastType,
		Author: author,
		Body:   body,
	}
}

func Deserialize(strData string) Message {
	var jsonData Message
	err := json.Unmarshal([]byte(strData), &jsonData)
	if err != nil {
		panic(err)
	}

	return jsonData
}

func Serialize(jsonData Message) string {
	strData, err := json.Marshal(jsonData)
	if err != nil {
		panic(err)
	}

	return string(strData)
}
