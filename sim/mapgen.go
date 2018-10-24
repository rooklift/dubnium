package sim

import (
	"math"
)

func MapGen(players, width, height, energy int, seed int32) *Frame {

	frame := new(Frame)

	for pid := 0; pid < players; pid++ {
		frame.budgets = append(frame.budgets, energy)
		frame.deposited = append(frame.deposited, 0)
	}

	frame.halite = make([][]int, width)

	for x := 0; x < width; x++ {
		frame.halite[x] = make([]int, height)
	}

	noise := make_2d_float_array(width, height)

	p := NewPerlin(2, 2, 20, int64(seed))
	q := NewPerlin(2, 2, 10, int64(seed))
	r := NewPerlin(2, 2, 3, int64(seed))

	lowest := 99999.0
	highest := -99999.0

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			fx := float64(x) + 0.5
			fy := float64(y) + 0.5

			dx := math.Abs((float64(width) / 2) - fx) / float64(width / 2)
			dy := math.Abs((float64(height) / 2) - fy) / float64(height / 2)

			a := q.Noise2D(dx, dy)
			b := p.Noise2D(dx, dy)
			c := r.Noise2D(dx, dy)

			noise[x][y] = a + b - c

			if noise[x][y] < lowest { lowest = noise[x][y] }
			if noise[x][y] > highest { highest = noise[x][y] }
		}
	}

	// Normalise to a sane range...

	const (
		MAX_WANTED = 800.0
		MIN_WANTED = -300.0
	)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			// Initial normalised value...

			fval := (MAX_WANTED - MIN_WANTED) / (highest - lowest) * (noise[x][y] - highest) + MAX_WANTED
			frame.halite[x][y] = int(fval)

			if frame.halite[x][y] < 0 {
				frame.halite[x][y] = 0
			}
		}
	}

	// Place factories...

	dx := 8
	dy := 8

	for pid := 0; pid < players; pid++ {

		var x int
		var y int

		if pid == 0 {
			x = width / 2 - dx - 1
			y = height / 2 - dy - 1
		} else if pid == 1 {
			x = width / 2 + dx
			y = height / 2 + dy
		} else if pid == 2 {
			x = width / 2 - dx - 1
			y = height / 2 + dy
		} else {
			x = width / 2 + dx
			y = height / 2 - dy - 1
		}

		factory := &Dropoff{
			Factory: true,
			Owner: pid,
			Sid: -1,
			X: x,
			Y: y,
			Gathered: 0,
		}

		frame.dropoffs = append(frame.dropoffs, factory)
		frame.halite[x][y] = 0
	}

	return frame
}


func make_2d_float_array(width, height int) [][]float64 {
	ret := make([][]float64, width)
	for x := 0; x < width; x++ {
		ret[x] = make([]float64, height)
	}
	return ret
}
