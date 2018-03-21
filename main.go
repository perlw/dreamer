package main

import (
	"log"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

/*
func proxy(client net.Conn, server net.Conn) {
	serverClosed := make(chan int, 1)
	clientClosed := make(chan int, 1)

	go streamer(server, client, clientClosed)
	go streamer(client, server, serverClosed)

	var waitFor chan int
	select {
	case <-clientClosed:
		server.Close()
		waitFor = serverClosed
	case <-serverClosed:
		client.Close()
		waitFor = clientClosed
	}

	<-waitFor
}

func (r Rule) String() string {
	return fmt.Sprintf("telnet:%s->%d", r.Name, r.Destination)
}
*/

func main() {
	cfg, err := ini.Load("dreamer.ini")
	if err != nil {
		return errors.Wrap(err, "could not read config")
	}
	cfg.BlockMode = false

	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == "DEFAULT" {
			continue
		}

		log.Println(name)
	}
}
