package sim

import (
	"fmt"
	"strconv"
	"strings"
)

func (self *Game) UpdateFromMoves(all_player_moves []string) (string, *ReplayFrame) {

	players := self.frame.Players()
	width := self.frame.Width()
	height := self.frame.Height()

	if len(all_player_moves) != players {
		panic("len(all_player_moves) != players")
	}

	rf := new(ReplayFrame)

	rf.Cells = make([]*CellUpdate, 0)
	rf.Deposited = make(map[int]int)
	rf.Energy = make(map[int]int)
	rf.Entities = make(map[int]map[int]*Ship)
	rf.Events = make([]*ReplayEvent, 0)
	rf.Moves = make(map[int][]*ReplayMove)

	for pid := 0; pid < players; pid++ {
		rf.Entities[pid] = make(map[int]*Ship)
	}

	for _, ship := range self.frame.ships {

		if ship == nil {
			continue
		}

		rf.Entities[ship.Owner][ship.Sid] = ship
	}

	for pid := 0; pid < players; pid++ {
		rf.Moves[pid] = make([]*ReplayMove, 0)
	}

	gens := make(map[int]bool)		// pid --> generating?
	moves := make(map[int]string)	// sid --> move (n, s, e, w, o, c, "")

	fails := make(map[int]string)	// pid --> reason, or "" if success

	for pid, s := range all_player_moves {

		tokens := strings.Fields(s)

		command := ""				// g, c, m
		sid := -1

		TokenLoop:
		for _, token := range tokens {

			if command == "" {

				command = token

				if command != "g" && command != "m" && command != "c" {
					fails[pid] = fmt.Sprintf("Got unknown command \"%s\"", command)
					break TokenLoop
				}

				if command == "g" {
					if gens[pid] {
						fails[pid] = "Got 2 or more generate commands"
						break TokenLoop
					}
					gens[pid] = true
					command = ""
				}
				continue
			}

			if sid == -1 {

				var err error
				sid, err = strconv.Atoi(token)

				if err != nil {
					fails[pid] = err.Error()
					break TokenLoop
				}

				if sid >= len(self.frame.ships) || sid < 0 || self.frame.ships[sid] == nil {
					fails[pid] = fmt.Sprintf("Got command for non-existent ship %d", sid)
					break TokenLoop
				}

				// So the sid is a valid ship...

				ship := self.frame.ships[sid]

				if ship.Owner != pid {
					fails[pid] = fmt.Sprintf("Got command for ship %d owned by player %d", sid, ship.Owner)
					break TokenLoop
				}

				// So the ship is indeed owned by the player...

				if moves[sid] != "" {
					fails[pid] = fmt.Sprintf("Got 2 or more commands for ship %d", sid)
					break TokenLoop
				}

				if command == "c" {

					for _, dropoff := range self.frame.dropoffs {
						if dropoff.X == ship.X && dropoff.Y == ship.Y {
							fails[pid] = fmt.Sprintf("Got construct command from ship %d over a structure", sid)
							break TokenLoop
						}
					}

					moves[sid] = "c"
					command = ""
					sid = -1
				}
				continue
			}

			direction := token

			if direction != "n" && direction != "s" && direction != "e" && direction != "w" && direction != "o" {
				fails[pid] = fmt.Sprintf("Got unknown direction \"%s\"", direction)
				break TokenLoop
			}

			moves[sid] = direction

			command = ""
			sid = -1
		}
	}

	// ---------------------------------------------------------------

	new_frame := self.frame.Copy()
	new_frame.turn += 1

	// Best just pretend ships with no move did in fact issue a "o" order...

	for _, ship := range self.frame.ships {
		if ship == nil {
			continue
		}
		if moves[ship.Sid] == "" {
			moves[ship.Sid] = "o"
		}
	}

	// Adjust budgets...

	for pid := 0; pid < players; pid++ {
		if gens[pid] {
			new_frame.budgets[pid] -= 1000		// FIXME: don't hardcode
		}
	}

	for sid, move := range moves {

		if move == "c" {

			ship := new_frame.ships[sid]		// Note we already checked this sid exists and is not nil.
			pid := ship.Owner

			new_frame.budgets[pid] -= 4000		// FIXME: don't hardcode
			new_frame.budgets[pid] += ship.Halite
			new_frame.budgets[pid] += new_frame.halite[ship.X][ship.Y]
		}
	}

	for pid := 0; pid < players; pid++ {
		if new_frame.budgets[pid] < 0 {
			fails[pid] = "Went over budget"
		}
	}

	// Print info on fails...

	for pid := 0; pid < players; pid++ {
		if fails[pid] != "" {
			fmt.Printf("%s\n", fails[pid])
		}
	}

	// Clear gens / budgets of dying players...

	for pid := 0; pid < players; pid++ {
		if fails[pid] != "" {
			gens[pid] = false
			new_frame.budgets[pid] = 0
		}
	}

	// Clear all ships of dying players...

	for i, ship := range new_frame.ships {
		if ship == nil {
			continue
		}
		if fails[ship.Owner] != "" {
			new_frame.ships[i] = nil
		}
	}

	// All surviving moves are valid... (I hope)...

	// Make dropoffs...

	for sid, move := range moves {

		if move != "c" {
			continue
		}

		ship := new_frame.ships[sid]		// Note we already checked this sid exists, but it may have been made nil above.

		if ship == nil {
			continue
		}

		// We also previously checked that there's no structure here already.
		// And we already updated the scores.

		dropoff := &Dropoff{
			Factory: false,
			Owner: ship.Owner,
			Sid: ship.Sid,
			X: ship.X,
			Y: ship.Y,
		}

		new_frame.dropoffs = append(new_frame.dropoffs, dropoff)
		new_frame.ships[ship.Sid] = nil

		new_frame.deposited[ship.Owner] += ship.Halite
		new_frame.deposited[ship.Owner] += new_frame.halite[ship.X][ship.Y]

		new_frame.halite[ship.X][ship.Y] = 0	// Do this after the above lol

		rf.Events = append(rf.Events, &ReplayEvent{
			Sid: ship.Sid,
			Location: &Position{ship.X, ship.Y},
			Owner: ship.Owner,
			Type: "construct",
		})
	}

	// Move ships...

	ship_positions := make(map[Position][]*Ship)

	for sid, move := range moves {

		ship := new_frame.ships[sid]		// Note we already checked this sid exists, but it may have been made nil above.

		if ship == nil {
			continue
		}

		if ship.Halite >= new_frame.halite[ship.X][ship.Y] / 10 {		// We can move

			if move != "" && move != "o" && move != "c" {				// We did move

				ship.Halite -= new_frame.halite[ship.X][ship.Y] / 10

				dx, dy := string_to_dxdy(move)

				ship.X += dx
				ship.Y += dy
				ship.X = mod(ship.X, width)
				ship.Y = mod(ship.Y, height)
			}
		}

		ship_positions[Position{ship.X, ship.Y}] = append(ship_positions[Position{ship.X, ship.Y}], ship)
	}

	// Find places that want to spawn, so we can check for collisions...

	attempted_spawn_points := make(map[Position]bool)

	for pid := 0; pid < players; pid++ {

		if gens[pid] {

			factory := new_frame.dropoffs[pid]
			x := factory.X
			y := factory.Y

			attempted_spawn_points[Position{x, y}] = true
		}
	}

	// Delete ships that collide...

	collision_points := make(map[Position]bool)

	for x := 0; x < width; x++ {

		for y := 0; y < height; y++ {

			ships_here := ship_positions[Position{x, y}]

			if len(ships_here) == 0 {
				continue
			}

			if len(ships_here) == 1 && attempted_spawn_points[Position{x, y}] == false {
				continue
			}

			// Collision...

			collision_points[Position{x, y}] = true

			var wreckedsids []int

			for _, ship := range ships_here {
				new_frame.ships[ship.Sid] = nil
				new_frame.halite[x][y] += ship.Halite			// Dump the halite on the ground.
				wreckedsids = append(wreckedsids, ship.Sid)
			}

			rf.Events = append(rf.Events, &ReplayEvent{
				Location: &Position{x, y},
				WreckedSids: wreckedsids,
				Type: "shipwreck",
			})
		}
	}

	// Deliveries...

	for _, dropoff := range new_frame.dropoffs {

		pid := dropoff.Owner
		x := dropoff.X
		y := dropoff.Y
		ships_here := ship_positions[Position{x, y}]

		// First, handle halite that is on the ground (due to collisions)...

		if new_frame.halite[x][y] > 0 {
			new_frame.budgets[pid] += new_frame.halite[x][y]
			new_frame.deposited[pid] += new_frame.halite[x][y]
			new_frame.halite[x][y] = 0
		}

		// Now do normal deliveries...

		if len(ships_here) == 1 && collision_points[Position{x, y}] == false {

			if ships_here[0].Owner == pid {
				new_frame.budgets[pid] += ships_here[0].Halite
				new_frame.deposited[pid] += ships_here[0].Halite
				ships_here[0].Halite = 0
			}
		}
	}

	// Gen...

	for pid := 0; pid < players; pid++ {

		if gens[pid] {

			factory := new_frame.dropoffs[pid]

			x := factory.X
			y := factory.Y

			if collision_points[Position{x, y}] {		// Cancel the spawn due to collision
				continue
			}

			sid := len(new_frame.ships)

			ship := &Ship{
				Owner: pid,
				Sid: sid,
				X: x,
				Y: y,
				Halite: 0,
				Inspired: false,		// Doesn't really matter; but in official is this set correctly?
			}

			new_frame.ships = append(new_frame.ships, ship)

			rf.Events = append(rf.Events, &ReplayEvent{
				Energy: 0,
				Sid: sid,
				Location: &Position{x, y},
				Owner: pid,
				Type: "spawn",
			})
		}
	}

	// Mining...

	for sid, ship := range new_frame.ships {

		if ship == nil {
			continue
		}

		if sid >= len(self.frame.ships) {
			break
		}

		old_ship := self.frame.ships[sid]

		if old_ship.X == ship.X && old_ship.Y == ship.Y {

			// Normal mining...

			amount_to_mine := (new_frame.halite[ship.X][ship.Y] + 3) / 4		// FIXME, don't hardcode, round properly

			if amount_to_mine + ship.Halite >= 1000 {
				amount_to_mine = 1000 - ship.Halite
			}

			ship.Halite += amount_to_mine
			new_frame.halite[ship.X][ship.Y] -= amount_to_mine

			// Inspired bonus... (doesn't remove halite from ground)

			if ship.Inspired {

				inspired_bonus := amount_to_mine * 2

				if inspired_bonus + ship.Halite >= 1000 {
					inspired_bonus = 1000 - ship.Halite
				}

				ship.Halite += inspired_bonus
			}
		}
	}

	// Fix inspiration of the new frame...

	new_frame.fix_inspiration(
		self.Constants.INSPIRATION_RADIUS,
		self.Constants.INSPIRATION_SHIP_COUNT)

	// Some replay stuff...

	for pid := 0; pid < players; pid++ {
		if fails[pid] != "" {
			continue
		}
		if gens[pid] {
			rf.Moves[pid] = append(rf.Moves[pid], &ReplayMove{Type: "g"})
		}
	}

	for sid, move := range moves {

		pid := self.frame.ships[sid].Owner

		if fails[pid] != "" {
			continue
		}

		if move == "c" {
			rf.Moves[pid] = append(rf.Moves[pid], &ReplayMove{
				Type: "c",
				Sid: sid,
			})
		} else {
			rf.Moves[pid] = append(rf.Moves[pid], &ReplayMove{
				Type: "m",
				Sid: sid,
				Direction: move,
			})
		}
	}

	rf.Cells = make_cell_updates(self.frame, new_frame)

	for pid := 0; pid < players; pid++ {
		rf.Energy[pid] = new_frame.budgets[pid]
	}

	for pid := 0; pid < players; pid++ {
		rf.Deposited[pid] = new_frame.deposited[pid]
	}

	s := make_bot_update_string(self.frame, new_frame)
	self.frame = new_frame
	return s, rf
}
