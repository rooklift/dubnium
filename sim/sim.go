package sim

import (
	"fmt"
	"strconv"
	"strings"
)

type Frame struct {
	turn						int
	last_alive					[]int
	budgets						[]int
	deposited					[]int
	halite						[][]int
	ships						[]*Ship			// Index is also Sid. Dead ships are nil.
	dropoffs					[]*Dropoff		// The first <player_count> items are always the factories. Index is arbitrary otherwise.
}

func (self *Frame) Width() int {
	return len(self.halite)
}

func (self *Frame) Height() int {
	return len(self.halite[0])
}

func (self *Frame) Players() int {
	return len(self.budgets)
}

func (self *Frame) TotalHalite() int {
	count := 0
	for x := 0; x < self.Width(); x++ {
		for y := 0; y < self.Height(); y++ {
			count += self.halite[x][y]
		}
	}
	return count
}

func (self *Frame) IsAlive(pid int) bool {
	return self.last_alive[pid] == -1
}

func (self *Frame) Kill(pid int) {
	self.last_alive[pid] = self.turn
}

func (self *Frame) DeathTime(pid int) int {
	return self.last_alive[pid]
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

	for _, la := range self.last_alive {
		new_frame.last_alive = append(new_frame.last_alive, la)
	}

	// ---------------------------------------------------------------

	width, height := self.Width(), self.Height()

	new_frame.halite = make_2d_int_array(width, height)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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
			Gathered: dropoff.Gathered,
		}

		new_frame.dropoffs = append(new_frame.dropoffs, dropoff_copy)
	}

	return new_frame
}

func (self *Frame) fix_inspiration(RADIUS int, SHIPS_NEEDED int) {

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

				other_x := mod(ship.X + x, width)
				other_y := mod(ship.Y + y, height)

				other := xy_lookup[Position{other_x, other_y}]

				if other != nil {
					if other.Owner != ship.Owner {
						hits++
					}
				}

				if y != 0 {

					other_y = mod(ship.Y - y, height)

					other := xy_lookup[Position{other_x, other_y}]

					if other != nil {
						if other.Owner != ship.Owner {
							hits++
						}
					}
				}
			}
		}

		if hits >= SHIPS_NEEDED {
			ship.Inspired = true
		} else {
			ship.Inspired = false
		}
	}
}

// ------------------------------------------------------------------------------------------

type Game struct {
	Constants					*Constants
	frame						*Frame
}

func NewGame(constants *Constants) *Game {

	self := new(Game)

	self.Constants = constants
	self.frame = nil				// To be set by caller.

	return self
}

func (self *Game) UseFrame(f *Frame) {
	self.frame = f
}

func (self *Game) BotInitString() string {

	// This returns the string with the factories and map,
	// but NOT the player count or pid, nor the JSON.

	var lines []string

	for pid := 0; pid < self.frame.Players(); pid++ {

		factory := self.frame.dropoffs[pid]
		x := factory.X
		y := factory.Y

		lines = append(lines, fmt.Sprintf("%d %d %d", pid, x, y))
	}

	lines = append(lines, fmt.Sprintf("%d %d", self.frame.Width(), self.frame.Height()))

	for y := 0; y < self.frame.Height(); y++ {

		var elements []string

		for x := 0; x < self.frame.Width(); x++ {
			elements = append(elements, strconv.Itoa(self.frame.halite[x][y]))
		}

		lines = append(lines, strings.Join(elements, " "))
	}

	return strings.Join(lines, "\n")		// There is no final newline returned.
}

func (self *Game) GetRank(pid int) int {

	money := self.frame.budgets[pid]
	rank := 1

	for n := 0; n < self.frame.Players(); n++ {

		if n == pid {
			continue
		}

		if self.frame.budgets[n] > money {
			rank++
		}
	}

	return rank
}

func (self *Game) GetDropoffs() []*Dropoff {				// Needed for replay stats
	var ret []*Dropoff
	for _, dropoff := range self.frame.dropoffs {
		ret = append(ret, dropoff)
	}
	return ret
}

func (self *Game) Budget(pid int) int {
	return self.frame.budgets[pid]
}

func (self *Game) IsAlive(pid int) bool {
	return self.frame.IsAlive(pid)
}

func (self *Game) Kill(pid int) {
	self.frame.Kill(pid)
}

func (self *Game) DeathTime(pid int) int {
	return self.frame.DeathTime(pid)
}

type Dropoff struct {
	Factory						bool
	Owner						int
	Sid							int			// Make this -1 if factory I guess (official engine does this too)
	X							int
	Y							int
	Gathered					int			// Halite gathered here - for stats
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
