package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"xmpp/internal/client"
)

var (
	host     string
	port     int
	username string
	verbose  bool
)

func init() {
	flag.StringVar(&host, "host", "localhost", "host")
	flag.IntVar(&port, "port", 5222, "port")
	flag.StringVar(&username, "u", "", "username")
	flag.BoolVar(&verbose, "v", false, "verbose")
}

func main() {

	flag.Parse()

	if username == "" {
		log.Fatal("username is empty")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected")

	c := client.New(conn, username)
	defer c.Close()

	if err := c.Auth(); err != nil {
		log.Fatal(err)
	}

	log.Println("authenticated")

	scanner := bufio.NewScanner(os.Stdin)

	go func() {
		for {
			fmt.Print("receiver > ")
			scanner.Scan()
			to := scanner.Text()

			fmt.Print("> ")
			scanner.Scan()
			text := scanner.Text()
			if err := c.Send(to, text); err != nil {
				log.Fatal(err)
			}
		}
	}()

	<-sigChan

	log.Println("disconnecting...")

	err = c.Close()
	if err != nil {
		return
	}
}
