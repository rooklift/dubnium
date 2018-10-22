package sim

import (
	"fmt"
	"strconv"
	"strings"
)

type Frame struct {
	turn						int
	budgets						[]int
	deposited					[]int
	halite						[][]int
	ships						[]*Ship			// Index is also Sid. Dead ships are nil.
	dropoffs					[]*Dropoff		// The first <player_count> items are always the factories. Index is arbitrary otherwise.
}

func (self *Frame) Copy() *Frame {

	new_frame := new(Frame)
	new_frame.turn = self.turn		// Meh, copy means copy.

	// ---------------------------------------------------------------

	for _, budget := range self.budgets {
		new_frame.budgets = append(new_frame.budgets, budget)
	}

	for _, deposited := range self.deposited {
		new_frame.deposited = append(new_frame.deposited, deposited)
	}

	// ---------------------------------------------------------------

	new_frame.halite = make([][]int, len(self.halite))
	for x := 0; x < len(self.halite); x++ {
		new_frame.halite[x] = make([]int, len(self.halite[0]))
	}

	for x := 0; x < len(self.halite); x++ {
		for y := 0; y < len(self.halite[0]); y++ {
			new_frame.halite[x][y] = self.halite[x][y]
		}
	}

	// ---------------------------------------------------------------

	for _, ship := range self.ships {

		if ship == nil {
			new_frame.ships = append(new_frame.ships, nil)
			continue
		}

		ship_copy := &Ship{
			Owner: ship.Owner,
			Sid: ship.Sid,
			X: ship.X,
			Y: ship.Y,
			Halite: ship.Halite,
			Inspired: ship.Inspired,
		}

		new_frame.ships = append(new_frame.ships, ship_copy)
	}

	// ---------------------------------------------------------------

	for _, dropoff := range self.dropoffs {

		dropoff_copy := &Dropoff{
			Factory: dropoff.Factory,
			Owner: dropoff.Owner,
			Sid: dropoff.Sid,
			X: dropoff.X,
			Y: dropoff.Y,
		}

		new_frame.dropoffs = append(new_frame.dropoffs, dropoff_copy)
	}

	return new_frame
}

func (self *Frame) FixInspiration() {

	RADIUS := 4

	width := len(self.halite)
	height := len(self.halite[0])

	xy_lookup := make(map[Position]*Ship)

	for _, ship := range self.ships {

		if ship == nil {
			continue
		}

		xy_lookup[Position{ship.X, ship.Y}] = ship
	}

	for _, ship := range self.ships {

		if ship == nil {
			continue
		}

		hits := 0

		for y := 0; y <= RADIUS; y++ {

			startx := y - RADIUS
			endx := RADIUS - y

			for x := startx; x <= endx; x++ {

				other_x := Mod(ship.X + x, width)
				other_y := Mod(ship.Y + y, height)

				other := xy_lookup[Position{other_x, other_y}]

				if other != nil {
					if other.Owner != ship.Owner {
						hits++
					}
				}

				if y != 0 {

					other_y = Mod(ship.Y - y, height)

					other := xy_lookup[Position{other_x, other_y}]

					if other != nil {
						if other.Owner != ship.Owner {
							hits++
						}
					}
				}
			}
		}

		if hits >= 2 {				// FIXME: don't hardcode
			ship.Inspired = true
		} else {
			ship.Inspired = false
		}
	}
}

type Game struct {

	Constants					*Constants

	players						int
	width						int
	height						int

	death						[]int		// time of death

	frame						*Frame
}

type Dropoff struct {
	Factory						bool
	Owner						int
	Sid							int			// Make this -1 if factory I guess (official engine does this too)
	X							int
	Y							int
}

type Ship struct {
	Owner						int		`json:"-"`					// Player ID
	Sid							int		`json:"-"`					// Ship ID		(will equal its index in the ships slice)
	X							int		`json:"x"`
	Y							int		`json:"y"`
	Halite						int		`json:"energy"`
	Inspired					bool	`json:"is_inspired"`
}

type Position struct {
	X							int		`json:"x"`
	Y							int		`json:"y"`
}

func NewGame(players, width, height int, seed int64, constants *Constants) *Game {

	self := new(Game)

	self.Constants = constants
	self.players = players
	self.death = make([]int, players)

	if width < 4 || height < 4 {
		width, height = choose_sizes(seed)
	}

	self.width = width
	self.height = height

	self.frame = mapgen(players, width, height, seed)

	return self
}

func (self *Game) BotInitString() string {

	// This returns the string with the factories and map,
	// but NOT the player count or pid, nor the JSON.

	var lines []string

	for pid := 0; pid < self.players; pid++ {

		factory := self.frame.dropoffs[pid]
		x := factory.X
		y := factory.Y

		lines = append(lines, fmt.Sprintf("%d %d %d", pid, x, y))
	}

	lines = append(lines, fmt.Sprintf("%d %d", self.width, self.height))

	for y := 0; y < self.height; y++ {

		var elements []string

		for x := 0; x < self.width; x++ {
			elements = append(elements, strconv.Itoa(self.frame.halite[x][y]))
		}

		lines = append(lines, strings.Join(elements, " "))
	}

	return strings.Join(lines, "\n")		// There is no final newline returned.
}

func (self *Game) GetRank(pid int) int {

	money := self.frame.budgets[pid]
	rank := 1

	for n := 0; n < self.players; n++ {

		if n == pid {
			continue
		}

		if self.frame.budgets[n] > money {
			rank++
		}
	}

	return rank
}
