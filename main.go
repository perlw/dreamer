package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

/** Missing codes
 * NULL (NUL)						0
 * Line Feed (LF)				10
 * Carriage Return (CR)	13
 * Bell (BEL)						7
 * BackSpace (BS)				8
 * Horizontal Tab (HT)	9
 * Vertical Tab (VT)		11
 * Form Feed (FF)				12
 */

type Command byte

const (
	CMD_SE Command = iota + 240
	CMD_NOP
	CMD_DATA
	CMD_BREAK
	CMD_IP
	CMD_ABORT
	CMD_AYT
	CMD_ERASE_CHARACTER
	CMD_ERASE_LINE
	CMD_GO
	CMD_SB

	// Options
	CMD_WILL
	CMD_WONT
	CMD_DO
	CMD_DONT

	CMD_IAC = 255
)

func (c Command) String() string {
	switch c {
	case CMD_IAC:
		return "IAC"
	case CMD_SE:
		return "SE"
	case CMD_NOP:
		return "NOP"
	case CMD_DATA:
		return "DATA"
	case CMD_BREAK:
		return "BREAK"
	case CMD_IP:
		return "IP"
	case CMD_ABORT:
		return "ABORT"
	case CMD_AYT:
		return "AYT"
	case CMD_ERASE_CHARACTER:
		return "ERASE_CHARACTER"
	case CMD_ERASE_LINE:
		return "ERASE_LINE"
	case CMD_GO:
		return "GO"
	case CMD_SB:
		return "SB"
	case CMD_WILL:
		return "WILL"
	case CMD_WONT:
		return "WONT"
	case CMD_DO:
		return "DO"
	case CMD_DONT:
		return "DONT"
	default:
		return "UNK"
	}
}

type Option byte

const (
	OPT_BINARY Option = iota
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
	OPT_PRAGMA_LOGON = iota + 0x8a
	OPT_SSPI_LOGON
	OPT_PRAGMA_HEARTBEAT
	// Unassigned options, 141-254
	OPT_EXTENDED = 0xff
)

func (o Option) String() string {
	switch o {
	case OPT_BINARY:
		return "BINARY"
	case OPT_ECHO:
		return "ECHO"
	case OPT_RECONNECTION:
		return "RECONNECTION"
	case OPT_SUPPRESS_GO_AHEAD:
		return "SUPPRESS_GO_AHEAD"
	case OPT_APPROX_MESSAGE_SIZE:
		return "APPROX_MESSAGE_SIZE"
	case OPT_STATUS:
		return "STATUS"
	case OPT_TIMING_MARK:
		return "TIMING_MARK"
	case OPT_REMOTE_CONTROLLED:
		return "REMOTE_CONTROLLED"
	case OPT_LINE_WIDTH:
		return "LINE_WIDTH"
	case OPT_PAGE_SIZE:
		return "PAGE_SIZE"
	case OPT_CARRIAGE_RETURN:
		return "CARRIAGE_RETURN"
	case OPT_HORIZ_TABS:
		return "HORIZ_TABS"
	case OPT_FORMFEED_DISP:
		return "FORMFEED_DISP"
	case OPT_VERT_TABS:
		return "VERT_TABS"
	case OPT_VERT_TAB_DISP:
		return "VERT_TAB_DISP"
	case OPT_LINEFEED_DISP:
		return "LINEFEED_DISP"
	case OPT_EXTENDED_ASCII:
		return "EXTENDED_ASCII"
	case OPT_LOGOUT:
		return "LOGOUT"
	case OPT_BYTE_MACRO:
		return "BYTE_MACRO"
	case OPT_DATA_ENTRY:
		return "DATA_ENTRY"
	case OPT_SUPDUP:
		return "SUPDUP"
	case OPT_SUPDUP_OUTPUT:
		return "SUPDUP_OUTPUT"
	case OPT_SEND_LOCATION:
		return "SEND_LOCATION"
	case OPT_TERMINAL_TYPE:
		return "TERMINAL_TYPE"
	case OPT_END_OF_RECORD:
		return "END_OF_RECORD"
	case OPT_TACACS:
		return "TACACS"
	case OPT_OUTPUT_MARKING:
		return "OUTPUT_MARKING"
	case OPT_TERMINAL:
		return "TERMINAL"
	case OPT_TELNET_3270:
		return "TELNET_3270"
	case OPT_X3_PAD:
		return "X3_PAD"
	case OPT_WINDOW_SIZE:
		return "WINDOW_SIZE"
	case OPT_TERMINAL_SPEED:
		return "TERMINAL_SPEED"
	case OPT_REMOTE_FLOW_CONTROL:
		return "REMOTE_FLOW_CONTROL"
	case OPT_LINEMODE:
		return "LINEMODE"
	case OPT_X_DISPLAY_LOCATION:
		return "X_DISPLAY_LOCATION"
	case OPT_ENVIRONMENT:
		return "ENVIRONMENT"
	case OPT_AUTHENTICATION:
		return "AUTHENTICATION"
	case OPT_ENCRYPTION:
		return "ENCRYPTION"
	case OPT_NEW_ENVIRONMENT:
		return "NEW_ENVIRONMENT"
	case OPT_TN3270E:
		return "TN3270E"
	case OPT_XAUTH:
		return "XAUTH"
	case OPT_CHARSET:
		return "CHARSET"
	case OPT_RSP:
		return "RSP"
	case OPT_COM_PORT_CONTROL:
		return "COM_PORT_CONTROL"
	case OPT_SUPPRESS_LOCAL_ECHO:
		return "SUPPRESS_LOCAL_ECHO"
	case OPT_START_TLS:
		return "START_TLS"
	case OPT_KERMIT:
		return "KERMIT"
	case OPT_SEND_URL:
		return "SEND_URL"
	case OPT_FORWARD_X:
		return "FORWARD_X"
	case OPT_PRAGMA_LOGON:
		return "PRAGMA_LOGON"
	case OPT_SSPI_LOGON:
		return "SSPI_LOGON"
	case OPT_PRAGMA_HEARTBEAT:
		return "PRAGMA_HEARTBEAT"
	case OPT_EXTENDED:
		return "EXTENDED"
	default:
		return "UNK"
	}
}

type AnsiSeq byte

const (
	ANSI_ESCAPE AnsiSeq = 0x1b

	ANSI_SS2 = 'N'
	ANSI_SS3 = 'O'
	ANSI_DCS = 'P'
	ANSI_CSI = '['
	ANSI_ST  = '\\'
	ANSI_OSC = ']'
	ANSI_SOS = 'X'
	ANSI_PM  = '^'
	ANSI_APC = '_'
	ANSI_RIS = 'c'
)

func (a AnsiSeq) String() string {
	switch a {
	case ANSI_ESCAPE:
		return "ESCAPE"
	case ANSI_SS2:
		return "SS2"
	case ANSI_SS3:
		return "SS3"
	case ANSI_DCS:
		return "DCS"
	case ANSI_CSI:
		return "CSI"
	case ANSI_ST:
		return "ST"
	case ANSI_OSC:
		return "OSC"
	case ANSI_SOS:
		return "SOS"
	case ANSI_PM:
		return "PM"
	case ANSI_APC:
		return "APC"
	case ANSI_RIS:
		return "RIS"
	default:
		return "UNK"
	}
}

func NewOptionSequence(cmd Command, opt Option) []byte {
	return []byte{byte(CMD_IAC), byte(cmd), byte(opt)}
}

type Blocked struct {
	Count int
	Since time.Time
}

var blockList map[string]Blocked

func spawnDreamer(conn net.Conn) {
	defer conn.Close()

	w := bufio.NewWriter(conn)
	r := bufio.NewReader(conn)

	fg256 := "38;5;%dm"
	bg256 := "48;5;%dm"
	nyanFg := []int{15, 0, 0, 0, 15, 15}
	nyanBg := []int{196, 214, 226, 34, 20, 91}

	w.Write(NewOptionSequence(CMD_DO, OPT_SUPPRESS_LOCAL_ECHO))
	w.Write(NewOptionSequence(CMD_WILL, OPT_ECHO))
	w.Write(NewOptionSequence(CMD_DONT, OPT_ECHO))
	w.Write(NewOptionSequence(CMD_DONT, OPT_LINEMODE))
	w.Write(NewOptionSequence(CMD_DONT, OPT_TERMINAL_SPEED))
	w.Write(NewOptionSequence(CMD_WONT, OPT_TERMINAL_SPEED))
	w.Flush()

	w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_OSC), '0', ';'})
	w.Write([]byte("dreamer"))
	w.Write([]byte{7})
	w.Flush()

	w.Write([]byte("Speak friend and enter"))
	w.Write([]byte{'\r', '\n', 0})
	w.Write([]byte("> "))
	w.Flush()

	accepted := false
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
			kind := Command(cmd[0])
			option := Option(cmd[1])
			log.Println("cmd:", cmd, kind, option)

			if Command(cmd[0]) == CMD_SB {
				buffer := make([]byte, 1)
				line := bytes.Buffer{}
				for Command(buffer[0]) != CMD_SE {
					n, err := r.Read(buffer)
					if err != nil {
						log.Println("subnegotiation: could not read from client,", err)
						return
					} else if n <= 0 {
						continue
					}
					line.WriteByte(buffer[0])
				}

				log.Println("SUBNEG:", Option(cmd[1]), []byte(line.String()))
				/*switch Option(cmd[1]) {
				case OPT_TERMINAL_SPEED:
				default:
				}*/
			}
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
			log.Println("Client said", []byte(msg), msg)
			if msg == "quit" {
				break
			}

			if !accepted {
				if msg != "mellon" {
					// +Blocklist checking
					var ip string
					parts := strings.Split(conn.RemoteAddr().String(), ":")
					if len(parts) > 2 {
						ip = strings.Join(parts[:len(parts)-1], ":")
					} else {
						ip = parts[0]
					}

					item, ok := blockList[ip]
					if !ok {
						item = Blocked{
							Count: 0,
							Since: time.Now(),
						}
					}
					item.Count++
					item.Since = time.Now()
					blockList[ip] = item
					// -Blocklist checking

					w.Write([]byte{'\r', '\n', 0})
					w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI)})
					w.Write([]byte(fmt.Sprintf(fg256, 11)))
					w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI)})
					w.Write([]byte(fmt.Sprintf(bg256, 9)))
					w.Write([]byte("YOU ARE NOT FRIEND; BEGONE"))
					w.Flush()
					break
				}
				accepted = true
			}

			line.Reset()
			w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI), '0', 'm'})
			w.Write([]byte{'\r', '\n', 0})
			w.Write([]byte("> "))
			w.Flush()
		} else {
			if buffer[0] >= 32 {
				line.WriteByte(buffer[0])
			}
			w.Write(buffer[0:1])
			w.Flush()
		}
	}

	w.Write([]byte{byte(ANSI_ESCAPE), byte(ANSI_CSI), '0', 'm'})
	w.Write([]byte{'\r', '\n', 0})
	w.Write([]byte("BYE"))
	w.Write([]byte{'\r', '\n', 0})
	w.Flush()
}

func serveDreamer() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, _ := ln.Accept()
		log.Println("Connection r[", conn.RemoteAddr(), "] l[", conn.LocalAddr(), "]")

		var ip string
		parts := strings.Split(conn.RemoteAddr().String(), ":")
		if len(parts) > 2 {
			ip = strings.Join(parts[:len(parts)-1], ":")
		} else {
			ip = parts[0]
		}
		if item, ok := blockList[ip]; ok {
			end := item.Since.Add(time.Minute * 5)
			if time.Now().Before(end) {
				log.Println("IS BLOCKED!", item.Count, "time")
				conn.Close()
				continue
			}
		}

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

	blockList = make(map[string]Blocked)

	go serveDreamer()
	//go serveGame()

	var forever chan int
	<-forever
}
