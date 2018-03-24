package main

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"net"
	"strings"
)

var (
	// ErrUnrecognizedCommand is returned when an unrecognized command is input to the interpreter
	ErrUnrecognizedCommand = errors.New("unrecognized command")
	// ErrUnknownExit is returned when an the a string can't be converted to a direction
	ErrUnknownExit = errors.New("unknown exit")
)

var gamePrompt = []byte("\r\n> ")

type ExitDir int

const (
	ExitDirN ExitDir = iota
	ExitDirE
	ExitDirS
	ExitDirW
	ExitDirNone
)

func (e ExitDir) String() string {
	switch e {
	case ExitDirN:
		return "north"
	case ExitDirE:
		return "east"
	case ExitDirS:
		return "south"
	case ExitDirW:
		return "west"
	}
	return "unk"
}

func StringToDirection(dir string) (ExitDir, error) {
	switch dir {
	case ExitDirN.String():
		return ExitDirN, nil
	case ExitDirE.String():
		return ExitDirE, nil
	case ExitDirS.String():
		return ExitDirS, nil
	case ExitDirW.String():
		return ExitDirW, nil
	}
	return ExitDirNone, ErrUnknownExit
}

type Action interface {
	Do(*GameState, *bufio.Writer)
}

type GoAction struct {
	TargetX int
	TargetY int
}

func (a GoAction) Do(state *GameState, w *bufio.Writer) {
	for idx, room := range state.Rooms {
		if room.PosX == a.TargetX && room.PosY == a.TargetY {
			state.CurrentRoom = &state.Rooms[idx]

			state.WriteCurrentRoomBrief(w)
		}
	}
}

type RoomExit struct {
	Id          string
	Description string
	Open        bool
}

type GameRoom struct {
	Brief       string
	Description string
	PosX        int
	PosY        int
	Exits       map[ExitDir]RoomExit
}

type GameCommand struct {
	Cmd     string
	Aliases []string
	Action  func(action string, arguments []string) (string, Action, error)
}

type GameState struct {
	Commands []GameCommand
	Rooms    []GameRoom

	CurrentRoom *GameRoom
}

func NewGameState() *GameState {
	state := &GameState{}

	state.Commands = []GameCommand{
		{
			Cmd:     "help",
			Aliases: []string{"h", "?"},
			Action: func(action string, arguments []string) (string, Action, error) {
				return "TBD", nil, nil
			},
		},
		{
			Cmd:     "look",
			Aliases: []string{},
			Action: func(action string, arguments []string) (string, Action, error) {
				if len(arguments) == 0 {
					return state.CurrentRoom.Description, nil, nil
				}

				// Match exits
				dir, err := StringToDirection(arguments[0])
				if err != nil {
					id := strings.Join(arguments, " ")
					for exitDir, exit := range state.CurrentRoom.Exits {
						if strings.Contains(exit.Id, id) {
							dir = exitDir
						}
					}
				}

				if exit, ok := state.CurrentRoom.Exits[dir]; ok {
					return exit.Description, nil, nil
				}

				return "no exit that way", nil, nil
			},
		},
		{
			Cmd:     "go",
			Aliases: []string{"enter"},
			Action: func(action string, arguments []string) (string, Action, error) {
				if len(arguments) == 0 {
					return "Where should I go?", nil, nil
				}

				// Match exits
				var dir ExitDir
				dir, err := StringToDirection(arguments[0])
				if err != nil {
					id := strings.Join(arguments, " ")
					for exitDir, exit := range state.CurrentRoom.Exits {
						if strings.Contains(exit.Id, id) {
							dir = exitDir
						}
					}
				}

				var dirX, dirY int
				switch dir {
				case ExitDirN:
					dirX = 0
					dirY = 1
				case ExitDirE:
					dirX = 1
					dirY = 0
				case ExitDirS:
					dirX = 0
					dirY = -1
				case ExitDirW:
					dirX = -1
					dirY = 0
				}
				// TODO: Leave through exit action
				return "navigating", GoAction{
					TargetX: state.CurrentRoom.PosX + dirX,
					TargetY: state.CurrentRoom.PosY + dirY,
				}, nil
			},
		},
	}

	state.Rooms = []GameRoom{
		{
			Brief:       "You're in a small jail cell.",
			Description: "long description",
			PosX:        0,
			PosY:        0,
			// Items, doors, monsters?
			Exits: map[ExitDir]RoomExit{
				ExitDirN: {
					Id:          "cell door",
					Description: "door description",
					Open:        false,
				},
			},
		},
		{
			Brief:       "You're standing in a hallway.",
			Description: "long description",
			PosX:        0,
			PosY:        1,
			Exits: map[ExitDir]RoomExit{
				ExitDirW: {
					Id:          "hallway",
					Description: "door description",
					Open:        true,
				},
				ExitDirE: {
					Id:          "hallway",
					Description: "door description",
					Open:        true,
				},
				ExitDirS: {
					Id:          "cell door",
					Description: "door description",
					Open:        false,
				},
			},
		},
	}

	state.CurrentRoom = &state.Rooms[0]

	return state
}

func (s *GameState) InterpretCommand(action string, arguments []string) (string, Action, error) {
	for _, cmd := range s.Commands {
		matched := false
		if cmd.Cmd == action {
			matched = true
		} else {
			for _, alias := range cmd.Aliases {
				if alias == action {
					matched = true
					break
				}
			}
		}

		if matched {
			return cmd.Action(action, arguments)
		}
	}

	return "", nil, ErrUnrecognizedCommand
}

func (s *GameState) WriteCurrentRoomBrief(w *bufio.Writer) {
	w.Write([]byte(s.CurrentRoom.Brief))
	w.Write([]byte("\n\nThere's a "))
	numExits := len(s.CurrentRoom.Exits)
	current := 2
	for dir, exit := range s.CurrentRoom.Exits {
		w.Write([]byte(exit.Id + " to the " + dir.String()))
		if numExits > 1 {
			if current < numExits {
				w.Write([]byte(", a "))
			} else if current == numExits {
				w.Write([]byte(", and a "))
			}
		}
		current++
	}
	w.Write([]byte("."))
	w.Flush()
}

func spawnGame(conn net.Conn) {
	defer conn.Close()

	state := NewGameState()

	w := bufio.NewWriter(conn)
	r := bufio.NewReader(conn)

	w.Write([]byte("The Dungeon\r\n===========\r\n\r\n"))
	state.WriteCurrentRoomBrief(w)
	w.Write(gamePrompt)
	w.Flush()

	lineEnd := false
	buffer := make([]byte, 1)
	line := bytes.Buffer{}
	for {
		n, err := r.Read(buffer)
		if err != nil {
			log.Println("could not read from client,", err)
			return
		} else if n <= 0 {
			continue
		}

		// Telnet commands
		if buffer[0] == 0xff {
			cmd := make([]byte, 2)
			n, err := r.Read(cmd)
			if n <= 0 || err != nil {
				log.Println("could not read cmd from client,", err)
				return
			}
			log.Println("cmd: ", cmd)
		} else if buffer[0] == '\n' || buffer[0] == '\r' {
			if line.Len() == 0 {
				if !lineEnd {
					lineEnd = true
					continue
				}
			}
			lineEnd = false

			msg := strings.TrimRight(line.String(), "\r\n")
			// Interpret the command
			if msg == "quit" {
				break
			} else {
				words := strings.Split(msg, " ")
				command := strings.ToLower(words[0])
				arguments := words[1:]

				output, action, err := state.InterpretCommand(command, arguments)
				if err != nil {
					log.Printf("could not interpret \"%s\", %s", command, err)

					if err == ErrUnrecognizedCommand {
						w.Write([]byte("Sorry, I don't know how to '" + command + "'..."))
					}
				} else {
					w.Write([]byte(output))

					if action != nil {
						action.Do(state, w)
					}
				}
				w.Write([]byte("\r\n"))
			}

			line.Reset()
			w.Write(gamePrompt)
			w.Flush()
		} else {
			line.WriteByte(buffer[0])
		}
	}

	w.Write([]byte("BYE\n\n"))
	w.Flush()
}
