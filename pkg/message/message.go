package message

type Message struct {
	To   string
	From string
	Text string
}

func NewMessage(to string, from string, text string) *Message {
	return &Message{
		To:   to,
		From: from,
		Text: text,
	}
}
