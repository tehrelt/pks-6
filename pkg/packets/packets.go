package packets

import (
	"xmpp/pkg/util"
)

const (
	PACKET_AUTH = iota
	PACKET_SEND
	PACKET_RECV
	PACKET_CLOSE
	PACKET_ERROR
	PACKET_OK
)

type Packet struct {
	Type int8
}

type AuthPacket struct {
	Packet
	Username string
}

type SendPacket struct {
	Packet
	From string
	To   string
	Data string
}

type ReceivePacket struct {
	Packet
	From string
	Data string
}

type ClosePacket struct {
	Packet
	Username string
}

type ErrorPacket struct {
	Packet
	Error string
}

func NewReceivePacket(from string, data string) *ReceivePacket {
	return &ReceivePacket{From: from, Data: data, Packet: NewPacket(PACKET_RECV)}
}

func NewPacket(Type int8) Packet {
	return Packet{Type: Type}
}

func NewErrorPacket(error string) *ErrorPacket {
	return &ErrorPacket{Error: error, Packet: NewPacket(PACKET_ERROR)}
}

func NewSendPacket(from string, to string, data string) *SendPacket {
	return &SendPacket{From: from, To: to, Data: data, Packet: NewPacket(PACKET_SEND)}
}

func NewAuthPacket(username string) *AuthPacket {
	return &AuthPacket{
		Packet:   NewPacket(PACKET_AUTH),
		Username: username,
	}
}

func UnmarshalPacket(data []byte) *Packet {
	p := new(Packet)
	p.Type = int8(data[0])
	return p
}

func UnmarshalAuthPacket(data []byte) *AuthPacket {
	p := new(AuthPacket)

	p.Type = int8(data[0])

	p.Username = string(data[1:])
	p.Username = util.StripNull(p.Username)

	return p
}

func UnmarshalSendPacket(data []byte) *SendPacket {
	p := new(SendPacket)

	cursor := 0

	p.Type = int8(data[0])
	cursor += 1

	p.From = string(data[cursor:])
	p.From = util.StripNull(p.From)
	cursor += len(p.From) + 1

	p.To = string(data[cursor:])
	p.To = util.StripNull(p.To)
	cursor += len(p.To) + 1

	p.Data = string(data[cursor:])
	p.Data = util.StripNull(p.Data)

	return p
}

func UnmarshalReceivePacket(data []byte) *ReceivePacket {
	p := new(ReceivePacket)

	cursor := 0

	p.Type = int8(data[0])
	cursor += 1

	p.From = string(data[cursor:])
	p.From = util.StripNull(p.From)
	cursor += len(p.From) + 1

	p.Data = string(data[cursor:])
	p.Data = util.StripNull(p.Data)

	return p
}

func UnmarshalErrorPacket(data []byte) *ErrorPacket {
	p := new(ErrorPacket)

	cursor := 0

	p.Type = int8(data[0])
	cursor += 1

	p.Error = string(data[cursor:])

	return p
}

func MarshalPacket(p Packet) []byte {
	data := make([]byte, 1)
	data[0] = byte(p.Type)
	return data
}

func MarshalAuthPacket(p *AuthPacket) []byte {
	data := make([]byte, 1+len(p.Username))
	data[0] = byte(p.Type)
	copy(data[1:], p.Username)
	return data
}

func MarshalSendPacket(p *SendPacket) []byte {
	data := make([]byte, 1+len(p.From)+1+len(p.To)+1+len(p.Data))

	data[0] = byte(p.Type)
	cursor := 1

	copy(data[cursor:], p.From)
	cursor += len(p.From) + 1

	copy(data[cursor:], p.To)
	cursor += len(p.To) + 1

	copy(data[cursor:], p.Data)
	return data
}

func MarshalReceivePacket(p *ReceivePacket) []byte {
	data := make([]byte, 1+len(p.From)+1+len(p.Data))

	data[0] = byte(p.Type)
	cursor := 1

	copy(data[cursor:], p.From)
	cursor += len(p.From) + 1

	copy(data[cursor:], p.Data)
	return data
}

func MarshalErrorPacket(p *ErrorPacket) []byte {
	data := make([]byte, 1+len(p.Error))
	data[0] = byte(p.Type)
	copy(data[1:], p.Error)
	return data
}

func UnmarshalClosePacket(bytes []byte) *ClosePacket {
	p := new(ClosePacket)
	p.Type = int8(bytes[0])
	p.Username = string(bytes[1:])
	p.Username = util.StripNull(p.Username)
	return p
}

func NewClosePacket(username string) *ClosePacket {
	return &ClosePacket{
		Username: username,
		Packet:   NewPacket(PACKET_CLOSE),
	}
}

func MarshalClosePacket(packet *ClosePacket) []byte {
	data := make([]byte, 1+len(packet.Username))
	data[0] = byte(packet.Type)
	copy(data[1:], packet.Username)
	return data
}
