package client

import (
	"fmt"
	"log"
	"net"
	"os"
	"xmpp/pkg/message"
	"xmpp/pkg/packets"
)

type Client struct {
	net.Conn
	username string
	messages chan *message.Message
}

func New(conn net.Conn, username string) *Client {
	return &Client{
		Conn:     conn,
		username: username,
		messages: make(chan *message.Message),
	}
}

func (c *Client) Close() error {

	closePacket := packets.NewClosePacket(c.username)
	log.Printf("Closing connection to %s", c.username)
	_, err := c.Write(packets.MarshalClosePacket(closePacket))
	if err != nil {
		return err
	}

	return c.Conn.Close()
}

func (c *Client) Auth() error {

	packet := packets.NewAuthPacket(c.username)

	_, err := c.Write(packets.MarshalAuthPacket(packet))
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)

	if _, err := c.Read(buf); err != nil {
		return err
	}

	if int8(buf[0]) == packets.PACKET_ERROR {

		p := packets.UnmarshalErrorPacket(buf)

		return fmt.Errorf("authentication failed: %s", p.Error)
	}

	go c.readMessages()

	return nil
}

func (c *Client) Send(to string, text string) error {
	p := packets.NewSendPacket(c.username, to, text)
	if _, err := c.Write(packets.MarshalSendPacket(p)); err != nil {
		return err
	}
	return nil
}

func (c *Client) readMessages() {

	log.Println("looking for messages")

	for {
		buf := make([]byte, 1024)

		//fmt.Println("waiting for message")
		if _, err := c.Read(buf); err != nil {
			fmt.Println("read message error: ", err)
			os.Exit(1)
		}

		t := int8(buf[0])
		//fmt.Println("message type: ", t)

		if t == packets.PACKET_RECV {
			p := packets.UnmarshalReceivePacket(buf)
			m := message.NewMessage(c.username, p.From, string(p.Data))
			fmt.Printf("%s: %s\n", m.From, m.Text)
			continue
		}

		if t == packets.PACKET_ERROR {
			p := packets.UnmarshalErrorPacket(buf)
			fmt.Println("ERROR: ", p.Error)
		}
	}
}
