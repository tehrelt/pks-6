package main

import (
	"flag"
	"log"
	"xmpp/internal/server"
)

var (
	host    string
	port    int
	verbose bool
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.IntVar(&port, "port", 5222, "port")
	flag.BoolVar(&verbose, "v", false, "verbose")
}

func main() {

	flag.Parse()

	s, err := server.New(host, port, &server.Options{
		Verbose: verbose,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}

}
