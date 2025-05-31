package message

import (
	"fmt"
	"encoding/json"
)
type Message struct {
	From string`json: "from"`
	Content string`json: "content"`
	Type string`json: "type"`
	To string`json: "to"`
}

func (msg *Message) MakeString() string {
	switch msg.Type {
	case "system":
		return fmt.Sprintf("[SYSTEM]: %s", msg.Content)
	case "message":
		return fmt.Sprintf("[%s]: %s", msg.From, msg.Content)
	case "private_message":
		return fmt.Sprintf("[%s]->[%s]: %s", msg.From, msg.To, msg.Content)

	}
	
	return ""
}

func MakeSystemMessage(msg string) Message {
	return Message{
		From: "none",
		Content: msg,
		Type: "system",
		To: "none",
	}
}

func MakeGeneralUserMessage(msg string, from_user string) Message {
	return Message{
		From: from_user,
		Content: msg,
		Type: "message",
		To: "none",
	}
}

func MakePrivateUserMessage(msg string, from_user string, to_user string) Message {
	return Message{
		From: from_user,
		Content: msg,
		Type: "private_message",
		To: to_user,
	}
}

func MessageToJson(msg Message) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		return make([]byte, 0)
	}
	return data
}