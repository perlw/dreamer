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

type Command byte

const (
	CMD_IAC Command = 0xff

	CMD_SE    = 0xf0
	CMD_NOP   = 0xf1
	CMD_DATA  = 0xf2
	CMD_BREAK = 0xf3
	CMD_IP    = 0xf4
	CMD_ABORT = 0xf5
	CMD_AYT   = 0xf6
	CMD_ERASE = 0xf7
	CMD_GO    = 0xf8
	CMD_SB    = 0xfa

	// Options
	CMD_WILL = 0xfb
	CMD_WONT = 0xfc
	CMD_DO   = 0xfd
	CMD_DONT = 0xfe
)

type Option byte

const (
	OPT_BINARY Option = 0x00
	OPT_ECHO
	OPT_RECONNECTION
	OPT_SUPPRESS_GO_AHEAD
	OPT_APPROX_MESSAGE_SIZE
	OPT_STATUS
	OPT_TIMING_MARK
	OPT_REMOTE_CONTROLLED
	OPT_LINE_WIDTH
	OPT_PAGE_SIZE
	OPT_CARRIAGE_RETURN
	OPT_HORIZ_TABS
	OPT_FORMFEED_DISP
	OPT_VERT_TABS
	OPT_VERT_TAB_DISP
	OPT_LINEFEED_DISP
	OPT_EXTENDED_ASCII
	OPT_LOGOUT
	OPT_BYTE_MACRO
	OPT_DATA_ENTRY
	OPT_SUPDUP
	OPT_SUPDUP_OUTPUT
	OPT_SEND_LOCATION
	OPT_TERMINAL_TYPE
	OPT_END_OF_RECORD
	OPT_TACACS
	OPT_OUTPUT_MARKING
	OPT_TERMINAL
	OPT_TELNET_3270
	OPT_X3_PAD
	OPT_WINDOW_SIZE
	OPT_TERMINAL_SPEED
	OPT_REMOTE_FLOW_CONTROL
	OPT_LINEMODE
	OPT_X_DISPLAY_LOCATION
	OPT_ENVIRONMENT
	OPT_AUTHENTICATION
	OPT_ENCRYPTION
	OPT_NEW_ENVIRONMENT
	OPT_TN3270E
	OPT_XAUTH
	OPT_CHARSET
	OPT_RSP
	OPT_COM_PORT_CONTROL
	OPT_SUPPRESS_LOCAL_ECHO
	OPT_START_TLS
	OPT_KERMIT
	OPT_SEND_URL
	OPT_FORWARD_X
	// Unassigned options, 50-137
	OPT_PRAGMA_LOGON = 0x8a
	OPT_SSPI_LOGON
	OPT_PRAGMA_HEARTBEAT
	// Unassigned options, 141-254
	OPT_EXTENDED
)

type AnsiSeq byte

const (
	ANSI_ESCAPE AnsiSeq = 0x1b
	ANSI_CSI            = '['
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
	w.Write([]byte{byte(CMD_IAC), 0xfd, 45}) // Suppress Local Echo
	// WILL the command
	w.Write([]byte{byte(CMD_IAC), 0xfb, 1}) // Echo
	// DON'T the command
	w.Write([]byte{byte(CMD_IAC), 0xfe, 1})  // Echo
	w.Write([]byte{byte(CMD_IAC), 0xfe, 34}) // Linemode
	w.Flush()

	w.Write([]byte("> "))
	w.Flush()

	lineEnd := false
	buffer := make([]byte, 1)
	line := bytes.Buffer{}
	for {
		w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI)})
		w.Write([]byte(fmt.Sprintf(fg256, nyanFg[line.Len()%len(nyanFg)])))
		w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI)})
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
		if buffer[0] == byte(CMD_IAC) {
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
			w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI), '0', 'm'})
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
