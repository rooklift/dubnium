package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

// -----------------------------------------------------------------------------------------

func bot_handler(cmd string, pid int, io chan string, pregame string) {

	// There are 2 clear places where this handler can hang: the 2 Scan() calls.
	// Therefore it is essential that main() never try to send to the io channel
	// unless it knows that those scans succeeded.

	bot_is_kill := false

	cmd_split := strings.Fields(cmd)
	exec_command := exec.Command(cmd_split[0], cmd_split[1:]...)

	// Note that the command isn't run until we call Start().
	// So the following is just setup for that and shouldn't fail.

	i_pipe, _ := exec_command.StdinPipe()
	o_pipe, _ := exec_command.StdoutPipe()
	e_pipe, _ := exec_command.StderrPipe()

	err := exec_command.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start bot %d (%s)\n", pid, cmd)
		bot_is_kill = true
	}

	if bot_is_kill == false {
		go pipe_to_stderr(e_pipe, pid)
		fmt.Fprintf(i_pipe, pregame)
		if pregame[len(pregame) - 1] != '\n' {
			fmt.Fprintf(i_pipe, "\n")
		}
	}

	scanner := bufio.NewScanner(o_pipe)

	if bot_is_kill == false && scanner.Scan() == false {				// So the Scan() happens if bot at least started.
		fmt.Fprintf(os.Stderr, "Bot %d output reached EOF\n", pid)
		bot_output_chan <- BotOutput{pid, "Non-starter (EOF)"}
		bot_is_kill = true
	} else if bot_is_kill {
		bot_output_chan <- BotOutput{pid, "Non-starter (exec)"}
	} else {
		bot_output_chan <- BotOutput{pid, scanner.Text()}
	}

	for {

		to_send := <- io			// Since this blocks, main() must never send via io unless it knows our last Scan() worked.

		if bot_is_kill == false {

			fmt.Fprintf(i_pipe, to_send)
			if to_send[len(to_send) - 1] != '\n' {
				fmt.Fprintf(i_pipe, "\n")
			}

			if scanner.Scan() == false {
				fmt.Fprintf(os.Stderr, "Bot %d output reached EOF\n", pid)
				bot_is_kill = true
			}

			bot_output_chan <- BotOutput{pid, scanner.Text()}

		} else {

			// Nothing, just let it time out

		}
	}
}

func pipe_to_stderr(p io.ReadCloser, pid int) {
	scanner := bufio.NewScanner(p)
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, "Bot %v: %v\n", pid, scanner.Text())
	}
}

// -----------------------------------------------------------------------------------------

func main() {

	start_time := time.Now()

	// This stuff should be a struct I guess...
	width, height, sleep, seed, no_timeout, no_replay, viewer, folder, infile, inPNG, botlist := parse_args()

	var provided_frame *sim.Frame

	if infile != "" {
		provided_frame, seed = sim.FrameFromFile(infile)
	} else if inPNG != "" {
		provided_frame = sim.FrameFromPNG(inPNG)
	}

	if provided_frame != nil {
		width = provided_frame.Width()
		height = provided_frame.Height()
	}

	turns := turns_from_size(width, height)

	players := len(botlist)

	if provided_frame != nil && provided_frame.Players() != players {
		fmt.Fprintf(os.Stderr, "Wrong number of bots (%d) given for this replay (need %d)\n", players, provided_frame.Players())
		return
	}

	if players < 1 || (players > 4 && provided_frame == nil) {
		fmt.Fprintf(os.Stderr, "Bad number of players: %d\n", players)
		return
	}

	io_chans := make([]chan string, players)

	for pid := 0; pid < players; pid++ {
		io_chans[pid] = make(chan string)
	}

	constants := sim.NewConstants(players, width, height, turns, seed)
	game := sim.NewGame(constants)

	if provided_frame == nil {
		game.UseFrame(sim.MapGenOfficial(players, width, height, constants.INITIAL_ENERGY, seed))
	} else {
		game.UseFrame(provided_frame)
	}

	json_blob_bytes, _ := json.Marshal(constants)
	json_blob := string(json_blob_bytes)
	json_blob = strings.Replace(json_blob, " ", "", -1)

	init_string := game.BotInitString()

	var pregame string

	for pid := 0; pid < players; pid++ {
		pregame = fmt.Sprintf("%s\n%d %d\n%s", json_blob, players, pid, init_string)
		go bot_handler(botlist[pid], pid, io_chans[pid], pregame)
	}

	if viewer {
		print_with_newline(pregame)		// The viewer will get the POV of the final player
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

			if no_timeout {
				continue GetNames
			}

			fmt.Fprintf(os.Stderr, "Hit the deadline. Received: %d\n", names_received)

			for pid := 0; pid < players; pid++ {
				if player_names[pid] == "" {
					player_names[pid] = "Non-starter (time)"
					game.Kill(pid, 0)
				}
			}

			break GetNames
		}
	}

	if viewer {
		j, _ := json.Marshal(player_names)
		fmt.Fprintf(os.Stderr, "{\"viewer_info\":{\"names\":%v}}\n", string(j))
	}

	replay := sim.NewReplay(player_names, game, turns, seed)

	move_strings := make([]string, players)

	// -----------------------------------------------------------------------------------------------------------------------

	for turn := 0; turn <= turns; turn++ {				// Don't mess with this now, we expect <= below...

		update_string, rf := game.UpdateFromMoves(move_strings)
		replay.FullFrames = append(replay.FullFrames, rf)

		// Send on every turn except final...

		if turn < turns {
			for pid := 0; pid < players; pid++ {
				if game.IsAlive(pid) {
					io_chans[pid] <- update_string		// THIS WILL HANG THE ENGINE IF THE HANDLER ISN'T AT THE RIGHT PLACE. Care!
				}
			}
		}

		if viewer {
			print_with_newline(update_string)
		}

		received := make([]bool, players)
		received_total := 0

		// Count dead players as already received "".
		// Also do this for all bots on the very final
		// frame (which is not updated).

		for pid := 0; pid < players; pid++ {
			if game.IsAlive(pid) == false || turn == turns {
				move_strings[pid] = ""
				received[pid] = true
				received_total++
			}
		}

		wait_start_time := time.Now()

		if received_total < players {

			deadline := time.NewTimer(2 * time.Second)

			Wait:
			for {

				select {

				case op := <- bot_output_chan:

					if game.IsAlive(op.Pid) {		// Bot hasn't crashed (if it had, we already pretended it sent "")

						received_total++
						received[op.Pid] = true
						move_strings[op.Pid] = op.Output

						if received_total >= players {
							deadline.Stop()
							break Wait
						}
					}

				case <- deadline.C:

					if no_timeout {
						continue Wait
					}

					for pid := 0; pid < players; pid++ {
						if received[pid] == false {
							move_strings[pid] = ""
							game.Kill(pid, -1)
							fmt.Fprintf(os.Stderr, "Hit the deadline. Killing bot %v\n", pid)
						}
					}

					break Wait
				}
			}
		}

		elapsed := time.Now().Sub(wait_start_time)
		wanted := time.Duration(sleep) * time.Millisecond

		if elapsed < wanted {
			time.Sleep(wanted - elapsed)
		}
	}

	// -----------------------------------------------------------------------------------------------------------------------

	_, rf := game.UpdateFromMoves(move_strings)
	replay.FullFrames = append(replay.FullFrames, rf)

	// Now the game is finished, we just need to do some stats and printing...

	replay.Stats = new(sim.ReplayStats)
	replay.Stats.NumTurns = turns + 1

	for pid := 0; pid < players; pid++ {

		replay.Stats.Pstats = append(replay.Stats.Pstats, new(sim.PlayerStats))
		replay.Stats.Pstats[pid].Pid = pid
		replay.Stats.Pstats[pid].Rank = game.GetRank(pid)

		replay.Stats.Pstats[pid].FinalProduction = game.Budget(pid)

		turn_last_alive := turns + 1		// Like in official replays
		if game.IsAlive(pid) == false {
			turn_last_alive = game.DeathTime(pid)
		}

		replay.Stats.Pstats[pid].LastTurnAlive = turn_last_alive
	}

	all_dropoffs := game.GetDropoffs()

	for _, dropoff := range all_dropoffs {
		replay.Stats.Pstats[dropoff.Owner].HalitePerDropoff = append(replay.Stats.Pstats[dropoff.Owner].HalitePerDropoff, dropoff)

		replay.Stats.Pstats[dropoff.Owner].TotalProduction += dropoff.Gathered

		if dropoff.Factory == false {
			replay.Stats.Pstats[dropoff.Owner].NumDropoffs += 1
		}
	}

	replay_filename := ""

	if no_replay == false {

		timestamp := time.Now().Format("20060102-150405-0700")

		if infile != "" {
			replay_filename = fmt.Sprintf("reload-%v-%v-%v-%v.hlt", timestamp, seed, width, height)
		} else {
			replay_filename = fmt.Sprintf("replay-%v-%v-%v-%v.hlt", timestamp, seed, width, height)
		}
		replay_filename = filepath.Join(folder, replay_filename)
		replay.Dump(replay_filename)
	}

	if viewer == false {

		type RankScore struct {
			Rank			int					`json:"rank"`
			Score			int					`json:"score"`
		}

		type PrintedStats struct {
			MapWidth		int					`json:"map_width"`
			MapHeight		int					`json:"map_height"`
			MapSeed			int32				`json:"map_seed"`
			Replay			string				`json:"replay"`
			Stats			map[int]RankScore	`json:"stats"`
			Time			string				`json:"time"`
		}

		ps := new(PrintedStats)

		ps.MapWidth = width
		ps.MapHeight = height
		ps.Replay = replay_filename
		ps.MapSeed = seed
		ps.Stats = make(map[int]RankScore)
		ps.Time = time.Now().Sub(start_time).Round(time.Millisecond).String()

		for pid := 0; pid < players; pid++ {

			rankscore := RankScore{
				Rank: replay.Stats.Pstats[pid].Rank,
				Score: replay.Stats.Pstats[pid].FinalProduction,
			}

			ps.Stats[pid] = rankscore
		}

		foo, _ := json.MarshalIndent(ps, "", "    ")

		fmt.Printf(string(foo))
		fmt.Printf("\n")
	}
}

// -----------------------------------------------------------------------------------------

func parse_args() (
		width, height, sleep int,
		seed int32,
		no_timeout, no_replay, viewer bool,
		folder, infile, inPNG string,
		botlist []string) {

	seed = int32(time.Now().UTC().Unix())
	folder = "./"

	dealt_with := make([]bool, len(os.Args))
	dealt_with[0] = true

	var err error

	for n, arg := range os.Args {

		if dealt_with[n] {
			continue
		}

		if arg == "--width" || arg == "-w" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			width, err = strconv.Atoi(os.Args[n + 1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't understand stated width.\n")
				os.Exit(1)
			}
			continue
		}

		if arg == "--height" || arg == "-h" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			height, err = strconv.Atoi(os.Args[n + 1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't understand stated height.\n")
				os.Exit(1)
			}
			continue
		}

		if arg == "--seed" || arg == "-s" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			seed64, err := strconv.ParseInt(os.Args[n + 1], 10, 32)		// ParseInt only here (I forget why)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't understand stated seed.\n")
				os.Exit(1)
			}
			seed = int32(seed64)
			continue
		}

		if arg == "--sleep" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			sleep, err = strconv.Atoi(os.Args[n + 1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Couldn't understand stated sleep.\n")
				os.Exit(1)
			}
			continue
		}

		if arg == "--file" || arg == "-f" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			infile = os.Args[n + 1]
			continue
		}

		if arg == "--png" || arg == "-g" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			inPNG = os.Args[n + 1]
			continue
		}

		if arg == "--replay-directory" || arg == "-i" {
			dealt_with[n] = true
			dealt_with[n + 1] = true
			folder = os.Args[n + 1]
			continue
		}

		if arg == "--viewer" || arg == "-u" {
			dealt_with[n] = true
			viewer = true
			continue
		}

		if arg == "--no-timeout" {
			dealt_with[n] = true
			no_timeout = true
			continue
		}

		if arg == "--no-replay" {
			dealt_with[n] = true
			no_replay = true
			continue
		}

		if arg == "--no-compression" {		// We already don't...
			dealt_with[n] = true
			continue
		}

		if arg == "--results-as-json" {		// We always do...
			dealt_with[n] = true
			continue
		}

		if arg == "--no-logs" {				// We already don't...
			dealt_with[n] = true
			continue
		}
	}

	for n, arg := range os.Args {

		if dealt_with[n] {
			continue
		}

		if len(arg) > 0 && arg[0] == '-' {
			fmt.Fprintf(os.Stderr, "Couldn't understand flag %v (not implemented)\n", arg)
			os.Exit(1)
		}
	}

	for n, arg := range os.Args {

		if dealt_with[n] {
			continue
		}

		botlist = append(botlist, arg)
	}

	if width == 0 && height > 0 { width = height }
	if height == 0 && width > 0 { height = width }

	if width < 2 || width > 128 || height < 2 || height > 128 {
		width = sim.SizeFromSeed(uint32(seed))
		height = width
	}

	return width, height, sleep, seed, no_timeout, no_replay, viewer, folder, infile, inPNG, botlist
}

// -----------------------------------------------------------------------------------------

func turns_from_size(width, height int) int {

	size := width
	if height > size {
		size = height
	}

	return ((size * 25) / 8) + 300
}

func print_with_newline(s string) {
	fmt.Printf(s)
	if len(s) == 0 || s[len(s) - 1] != '\n' {
		fmt.Printf("\n")
	}
}
