package sim

import (
	"math"
	"math/rand"
)

func MapGen(players, width, height, energy int, seed int32) *Frame {

	rand.Seed(int64(seed))

	frame := new(Frame)

	for pid := 0; pid < players; pid++ {
		frame.budgets = append(frame.budgets, energy)
		frame.deposited = append(frame.deposited, 0)
	}

	frame.halite = make_2d_int_array(width, height)

	// ----------------------------------------------------------------------------------------------------

	smooth_basis := smooth_noise(width, height, 1)		// Do we need this? Official has something similar.

	float_map := make_2d_float_array(width, height)

	p := NewPerlin(2, 2, 7, int64(seed))

	lowest := 99999.0
	highest := -99999.0

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			fx := float64(x) + 0.5
			fy := float64(y) + 0.5

			dx := math.Abs((float64(width) / 2) - fx) / float64(width / 2)
			dy := math.Abs((float64(height) / 2) - fy) / float64(height / 2)

			initial := p.Noise2D(dx, dy) * smooth_basis[x][y]
			float_map[x][y] = math.Pow(initial, 2)

			if float_map[x][y] < lowest { lowest = float_map[x][y] }
			if float_map[x][y] > highest { highest = float_map[x][y] }
		}
	}

	// Normalise to a sane range...

	const (
		MAX_WANTED = 1000.0
		MIN_WANTED = 0.0
	)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			fval := (MAX_WANTED - MIN_WANTED) / (highest - lowest) * (float_map[x][y] - highest) + MAX_WANTED
			frame.halite[x][y] = int(fval)

			if frame.halite[x][y] < 0 {
				frame.halite[x][y] = 0
			}
		}
	}

	// ----------------------------------------------------------------------------------------------------

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


func smooth_noise(width, height, cycles int) [][]float64 {

	noisemap := make_2d_float_array(width, height)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			noisemap[x][y] = rand.Float64()
		}
	}

	for n := 0; n < cycles; n++ {

		next_phase := make_2d_float_array(width, height)

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				neighbours := neighbours8(x, y, width, height)
				var sum float64
				for _, neigh := range neighbours {
					sum += noisemap[neigh.X][neigh.Y]
				}
				next_phase[x][y] = sum / float64(len(neighbours))
			}
		}

		noisemap = next_phase
	}

	// Make the whole thing symmetric... (which renders a lot of the above work pointless, but meh)

	for x := 0; x < width / 2; x++ {

		for y := 0; y < height / 2; y++ {

			val := noisemap[x][y]

			noisemap[width - 1 - x][y] = val
			noisemap[x][height - 1 - y] = val
			noisemap[width - 1 - x][height - 1 - y] = val

		}
	}

	return noisemap
}


func neighbours8(x, y, width, height int) []Position {

	var ret []Position

	ret = append(ret, Position{x - 1, y - 1})
	ret = append(ret, Position{x - 1, y + 0})
	ret = append(ret, Position{x - 1, y + 1})
	ret = append(ret, Position{x + 0, y - 1})
	ret = append(ret, Position{x + 0, y + 1})
	ret = append(ret, Position{x + 1, y - 1})
	ret = append(ret, Position{x + 1, y + 0})
	ret = append(ret, Position{x + 1, y + 1})

	for n := 0; n < len(ret); n++ {
		ret[n].X = mod(ret[n].X, width)
		ret[n].Y = mod(ret[n].Y, height)
	}

	return ret
}
