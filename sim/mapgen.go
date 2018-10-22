package sim

import (
	"math/rand"
)

func choose_sizes(seed int64) (int, int) {

	rand.Seed(seed)

	r := rand.Intn(5)

	width := 32 + (r * 8)
	height := width

	return width, height
}


func mapgen(players, width, height int, seed int64) *Frame {

	rand.Seed(seed)

	frame := new(Frame)

	for pid := 0; pid < players; pid++ {
		frame.budgets = append(frame.budgets, 5000)		// FIXME: don't hardcode
		frame.deposited = append(frame.deposited, 0)
	}

	frame.halite = make([][]int, width)

	for x := 0; x < width; x++ {
		frame.halite[x] = make([]int, height)
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			val := rand.Intn(128)

			frame.halite[x][y] += val

			frame.halite[Mod(x + 1, width)][y] += val
			frame.halite[Mod(x - 1, width)][y] += val

			frame.halite[x][Mod(y + 1, height)] += val
			frame.halite[x][Mod(y - 1, height)] += val
		}
	}

	var dx int
	var dy int

	if players == 2 {
		dx = 8
		dy = 0
	} else if players == 4 {
		dx = 8
		dy = 8
	}

	for pid := 0; pid < players; pid++ {

		var x int
		var y int

		if pid % 2 == 0 {
			x = width / 2 - dx
		} else {
			x = width / 2 + dx
		}

		if pid < players / 2 {
			y = height / 2 - dy
		} else {
			y = height / 2 + dy
		}

		factory := &Dropoff{
			Factory: true,
			Owner: pid,
			Sid: -1,
			X: x,
			Y: y,
		}

		frame.dropoffs = append(frame.dropoffs, factory)
		frame.halite[x][y] = 0
	}

	return frame
}
