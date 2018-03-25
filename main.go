package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

const (
	CMD_BEGIN  byte = 0xff
	CMD_ESCAPE      = 0x1b
	CMD_CSI         = '['
)

func spawnDreamer(conn net.Conn) {
	defer conn.Close()

	w := bufio.NewWriter(conn)
	r := bufio.NewReader(conn)

	fg256 := "38;5;%sm"
	bg256 := "48;5;%sm"
	nyanFg := []string{"15", "0", "0", "0", "15", "15"}
	nyanBg := []string{"196", "214", "226", "34", "20", "91"}

	// DO the command
	w.Write([]byte{CMD_BEGIN, 0xfd, 45}) // Suppress Local Echo
	// WILL the command
	w.Write([]byte{CMD_BEGIN, 0xfb, 1}) // Echo
	// DON'T the command
	w.Write([]byte{CMD_BEGIN, 0xfe, 1})  // Echo
	w.Write([]byte{CMD_BEGIN, 0xfe, 34}) // Linemode
	w.Flush()

	w.Write([]byte("> "))
	w.Flush()

	lineEnd := false
	buffer := make([]byte, 1)
	line := bytes.Buffer{}
	for {
		w.Write([]byte{CMD_ESCAPE, CMD_CSI})
		w.Write([]byte(fmt.Sprintf(fg256, nyanFg[line.Len()%len(nyanFg)])))
		w.Write([]byte{CMD_ESCAPE, CMD_CSI})
		w.Write([]byte(fmt.Sprintf(bg256, nyanBg[line.Len()%len(nyanBg)])))
		w.Flush()

		n, err := r.Read(buffer)
		if err != nil {
			log.Println("could not read from client,", err)
			return
		} else if n <= 0 {
			continue
		}

		// Telnet commands
		if buffer[0] == CMD_BEGIN {
			cmd := make([]byte, 2)
			n, err := r.Read(cmd)
			if n <= 0 || err != nil {
				log.Println("could not read cmd from client,", err)
				return
			}
			log.Println("cmd: ", cmd)

			// TODO: Handle options and settings
		} else if buffer[0] == '\n' || buffer[0] == '\r' {
			if line.Len() == 0 {
				if !lineEnd {
					lineEnd = true
					continue
				}
			}
			lineEnd = false

			msg := strings.TrimRight(line.String(), "\r\n")
			log.Println("Client said", msg)
			if msg == "quit" {
				break
			}

			line.Reset()
			w.Write([]byte{CMD_ESCAPE, CMD_CSI, '0', 'm'})
			w.Write([]byte("\r\n> "))
			w.Flush()
		} else {
			line.WriteByte(buffer[0])
			w.Write(buffer[0:1])
			w.Flush()
		}
	}

	w.Write([]byte("\r\nBYE\r\n"))
	w.Flush()
}

func serveDreamer() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, _ := ln.Accept()
		go spawnDreamer(conn)
	}
}

func serveGame() {
	ln, err := net.Listen("tcp", ":3001")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, _ := ln.Accept()
		go spawnGame(conn)
	}
}

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
		log.Fatal(errors.Wrap(err, "could not read config"))
	}
	cfg.BlockMode = false

	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == "DEFAULT" {
			continue
		}

		log.Println(name)
	}

	go serveDreamer()
	go serveGame()

	var forever chan int
	<-forever
}
