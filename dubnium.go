package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"./sim"
)

type BotOutput struct {
	Pid						int
	Output					string
}

var bot_output_chan = make(chan BotOutput)		// Shared by all bot handlers.


func bot_handler(cmd string, pid int, query chan string, io chan string, pregame string) {

	bot_is_kill := false

	cmd_split := strings.Fields(cmd)
	exec_command := exec.Command(cmd_split[0], cmd_split[1:]...)

	i_pipe, err := exec_command.StdinPipe()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	o_pipe, err := exec_command.StdoutPipe()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
/*
	e_pipe, err := exec_command.StderrPipe()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
*/
	err = exec_command.Start()
	if err != nil {
		fmt.Printf("Failed to start bot %d (%s)\n", pid, cmd)
		bot_is_kill = true
	}

	if bot_is_kill == false {
		fmt.Fprintf(i_pipe, pregame)
		if pregame[len(pregame) - 1] != '\n' {
			fmt.Fprintf(i_pipe, "\n")
		}
	}

	scanner := bufio.NewScanner(o_pipe)		// Is this OK if the exec failed? Probably.
	if scanner.Scan() == false {
		fmt.Printf("Bot %d output reached EOF\n", pid)
		bot_output_chan <- BotOutput{pid, "Non-starter (EOF)"}
		bot_is_kill = true
	} else {
		bot_output_chan <- BotOutput{pid, scanner.Text()}
	}

	for {

		to_send := <- io

		if bot_is_kill == false {

			fmt.Fprintf(i_pipe, to_send)
			if to_send[len(to_send) - 1] != '\n' {
				fmt.Fprintf(i_pipe, "\n")
			}

			if scanner.Scan() == false {
				fmt.Printf("Bot %d output reached EOF\n", pid)
				bot_is_kill = true
			}

			bot_output_chan <- BotOutput{pid, scanner.Text()}

		} else {

			bot_output_chan <- BotOutput{pid, ""}

		}
	}
}

func main() {

	width, height, seed, botlist, infile := parse_args()

	var initial_frame *sim.Frame

	if infile != "" {
		initial_frame, seed = sim.FrameFromFile(infile)
		width = initial_frame.Width()
		height = initial_frame.Height()
	}

	turns := turns_from_size(width, height)

	players := len(botlist)

	if initial_frame != nil && initial_frame.Players() != players {
		fmt.Printf("Wrong number of bots (%d) given for this replay (need %d)\n", players, initial_frame.Players())
		return
	}

	if players < 1 || players > 4 {
		fmt.Printf("Bad number of players: %d\n", players)
		return
	}

	query_chans := make([]chan string, players)
	io_chans := make([]chan string, players)

	for pid := 0; pid < players; pid++ {
		query_chans[pid] = make(chan string)
		io_chans[pid] = make(chan string)
	}

	crash_list := make([]bool, players)

	constants := sim.NewConstants(players, width, height, turns, seed)
	game := sim.NewGame(players, width, height, seed, constants)

	if initial_frame == nil {
		game.UseFrame(sim.MapGen(players, width, height, seed))
	} else {
		game.UseFrame(initial_frame)
	}

	json_blob_bytes, _ := json.Marshal(constants)
	json_blob := string(json_blob_bytes)
	json_blob = strings.Replace(json_blob, " ", "", -1)

	init_string := game.BotInitString()

	for pid := 0; pid < players; pid++ {
		pregame := fmt.Sprintf("%s\n%d %d\n%s", json_blob, players, pid, init_string)
		go bot_handler(botlist[pid], pid, query_chans[pid], io_chans[pid], pregame)
	}

	var player_names []string
	for pid := 0; pid < players; pid++ {
		player_names = append(player_names, "")
	}

	// Get names...

	names_received := 0
	deadline := time.NewTimer(30 * time.Second)

	GetNames:
	for {

		select {

		case op := <- bot_output_chan:

			names_received++
			player_names[op.Pid] = op.Output

			if player_names[op.Pid] == "" {
				player_names[op.Pid] = "(blank)"
			}

			if names_received >= players {
				deadline.Stop()
				break GetNames
			}

		case <- deadline.C:

			fmt.Printf("Hit the deadline. Received: %d\n", names_received)

			for pid := 0; pid < players; pid++ {
				if player_names[pid] == "" {
					player_names[pid] = "Non-starter (time)"
					crash_list[pid] = true
				}
			}

			break GetNames
		}
	}

	replay := sim.NewReplay(player_names, game, turns, seed)

	move_strings := make([]string, players)

	for turn := 0; turn <= turns; turn++ {

		update_string, rf := game.UpdateFromMoves(move_strings)

		replay.FullFrames = append(replay.FullFrames, rf)

		for pid := 0; pid < players; pid++ {
			if crash_list[pid] == false {
				io_chans[pid] <- update_string
			}
		}

		received := make([]bool, players)
		received_total := 0

		// Count dead players as already received ""

		for pid := 0; pid < players; pid++ {
			if crash_list[pid] {
				move_strings[pid] = ""
				received[pid] = true
				received_total++
			}
		}

		deadline := time.NewTimer(2 * time.Second)

		if received_total < players {

			Wait:
			for {

				select {

				case op := <- bot_output_chan:

					if crash_list[op.Pid] == false {		// If the bot has officially crashed we already pretended it sent ""

						received_total++
						received[op.Pid] = true
						move_strings[op.Pid] = op.Output

						if received_total >= players {
							deadline.Stop()
							break Wait
						}
					}

				case <- deadline.C:

					for pid := 0; pid < players; pid++ {
						if received[pid] == false {
							move_strings[pid] = ""
							crash_list[pid] = true
						}
					}

					break Wait
				}
			}
		} else {
			fmt.Printf("Skipped\n")
		}
	}

	_, rf := game.UpdateFromMoves(move_strings)
	replay.FullFrames = append(replay.FullFrames, rf)

	replay.Stats = new(sim.ReplayStats)

	for pid := 0; pid < players; pid++ {

		replay.Stats.Pstats = append(replay.Stats.Pstats, new(sim.PlayerStats))

		replay.Stats.Pstats[pid].Pid = pid
		replay.Stats.Pstats[pid].LastTurnAlive = turns					// FIXME
		replay.Stats.Pstats[pid].HalitePerDropoff = make([]bool, 0)
		replay.Stats.Pstats[pid].Rank = game.GetRank(pid)
	}

	if infile != "" {
		replay.Dump(fmt.Sprintf("reload-%v-%v-%v.hlt", seed, width, height))
	} else {
		replay.Dump(fmt.Sprintf("replay-%v-%v-%v.hlt", seed, width, height))
	}
}


func parse_args() (int, int, int64, []string, string) {

	var botlist []string
	infile := ""

	width := 0
	height := 0
	seed := time.Now().UTC().UnixNano()

	dealt_with := make([]bool, len(os.Args))
	dealt_with[0] = true

	for n, arg := range os.Args {

		if dealt_with[n] {
			continue
		}

		if arg == "--width" || arg == "-w" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			width, _ = strconv.Atoi(os.Args[n + 1])
			continue
		}

		if arg == "--height" || arg == "-h" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			height, _ = strconv.Atoi(os.Args[n + 1])
			continue
		}

		if arg == "--seed" || arg == "-s" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			seed, _ = strconv.ParseInt(os.Args[n + 1], 10, 64)
			continue
		}

		if arg == "--file" || arg == "-f" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			infile = os.Args[n + 1]
			continue
		}
	}

	for n, arg := range os.Args {

		if dealt_with[n] {
			continue
		}

		botlist = append(botlist, arg)
	}

	rand.Seed(seed)		// Use the seed to get width/height, if needed...

	if width == 0 && height > 0 { width = height }
	if height == 0 && width > 0 { height = width }

	if width < 32 || width > 64 || height < 32 || height > 64 {
		width = 32 + rand.Intn(5) * 8
		height = width
	}

	return width, height, seed, botlist, infile
}


func turns_from_size(width, height int) int {

	size := width
	if height > size {
		size = height
	}

	return (((size - 32) * 25) / 8) + 400
}
