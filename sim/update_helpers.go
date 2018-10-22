package sim

import (
	"fmt"
	"strconv"
	"strings"
)

func make_bot_update_string(old, current *Frame) string {

	// The string to send to a bot after the turn (before the next turn, whatever)

	players := len(current.budgets)
	width := len(current.halite)
	height := len(current.halite[0])

	var lines []string

	ship_counts := make([]int, players)
	dropoff_counts := make([]int, players)

	for _, ship := range current.ships {

		if ship == nil {
			continue
		}

		ship_counts[ship.Owner] += 1
	}

	for _, dropoff := range current.dropoffs {

		if dropoff.Factory {
			continue
		}

		dropoff_counts[dropoff.Owner] += 1
	}

	// ----------------------------------------------------

	lines = append(lines, strconv.Itoa(current.turn))

	// ----------------------------------------------------

	for pid := 0; pid < players; pid++ {

		lines = append(lines, fmt.Sprintf("%d %d %d %d", pid, ship_counts[pid], dropoff_counts[pid], current.budgets[pid]))

		for _, ship := range current.ships {

			if ship == nil {
				continue
			}

			if ship.Owner != pid {
				continue
			}

			lines = append(lines, fmt.Sprintf("%d %d %d %d", ship.Sid, ship.X, ship.Y, ship.Halite))
		}

		for _, dropoff := range current.dropoffs {

			if dropoff.Factory {
				continue
			}

			if dropoff.Owner != pid {
				continue
			}

			lines = append(lines, fmt.Sprintf("%d %d %d", dropoff.Sid, dropoff.X, dropoff.Y))
		}
	}

	// ----------------------------------------------------

	var update_lines []string

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if current.halite[x][y] != old.halite[x][y] {
				update_lines = append(update_lines, fmt.Sprintf("%d %d %d", x, y, current.halite[x][y]))
			}
		}
	}

	lines = append(lines, strconv.Itoa(len(update_lines)))
	lines = append(lines, update_lines...)

	return strings.Join(lines, "\n")		// There is no final newline returned.
}

func make_cell_updates(old, current *Frame) []*CellUpdate {

	// For the replay

	ret := make([]*CellUpdate, 0)

	width := len(current.halite)
	height := len(current.halite[0])

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if current.halite[x][y] != old.halite[x][y] {

				ret = append(ret, &CellUpdate{
					Production: current.halite[x][y],
					X: x,
					Y: y,
				})
			}
		}
	}

	return ret
}
