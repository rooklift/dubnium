package sim

/*
	This file is under the MIT License.

	Copyright (c) 2016 Michael Truell and Benjamin Spector

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
*/

import (
	"math"
	"math/rand"
)

func MapGenOfficial(players, width, height, player_energy int, seed int32) *Frame {

	// AFAIK, perfectly faithful reproduction of the official mapgen in file:
	// game_engine/mapgen/FractalValueNoiseTileGenerator.cpp

	rand.Seed(int64(seed))

	frame := new(Frame)

	for pid := 0; pid < players; pid++ {
		frame.budgets = append(frame.budgets, player_energy)
		frame.deposited = append(frame.deposited, 0)
	}

	frame.halite = make_2d_int_array(width, height)

	// ----------------------------------------------------------------------------------------------------

	tile_width := width / 2
	tile_height := height / 2

	if width % 2 == 1 { tile_width += 1 }
	if height % 2 == 1 { tile_height += 1 }

	tile := make_tile(tile_width, tile_height, 950)

	for x := 0; x < tile_width; x++ {

		for y := 0; y < tile_height; y++ {

			val := tile[x][y]

			frame.halite[x][y] = val
			frame.halite[width - 1 - x][y] = val
			frame.halite[x][height - 1 - y] = val
			frame.halite[width - 1 - x][height - 1 - y] = val
		}
	}

	place_factories(frame, players)

	return frame
}

func generate_smooth_noise(source_noise [][]float64, wavelength int) [][]float64 {

	// For consistency with original implementation, arrays are [y][x] format.

	mini_source := make_2d_float_array(
						int(math.Ceil(float64(len(source_noise)) / float64(wavelength))),
						int(math.Ceil(float64(len(source_noise[0])) / float64(wavelength))))

	for y := 0; y < len(mini_source); y++ {
		for x := 0; x < len(mini_source[0]); x++ {
			mini_source[y][x] = source_noise[wavelength * y][wavelength * x]
		}
	}

	smoothed_source := make_2d_float_array(len(source_noise), len(source_noise[0]))

	for y := 0; y < len(source_noise); y++ {

		var y_i int = y / wavelength
		var y_f int = (y / wavelength + 1) % len(mini_source)

		var vertical_blend float64 = float64(y) / float64(wavelength) - float64(y_i)

		for x := 0; x < len(source_noise[0]); x++ {

			var x_i int = x / wavelength
			var x_f int = (x / wavelength + 1) % len(mini_source[0])

			var horizontal_blend float64 = float64(x) / float64(wavelength) - float64(x_i)

			var top_blend float64 = (1 - horizontal_blend) * mini_source[y_i][x_i] + horizontal_blend * mini_source[y_i][x_f]
			var bottom_blend float64 = (1 - horizontal_blend) * mini_source[y_f][x_i] + horizontal_blend * mini_source[y_f][x_f]

			smoothed_source[y][x] = (1 - vertical_blend) * top_blend + vertical_blend * bottom_blend
		}
	}

	return smoothed_source
}

func make_tile(tile_width, tile_height, max_production int) [][]int {

	// Although various things here use [y][x] format, the tile itself is in our normal [x][y]...

	tile := make_2d_int_array(tile_width, tile_height)

	source_noise := make_2d_float_array(tile_height, tile_width)
	region := make_2d_float_array(tile_height, tile_width)

	const (
		FACTOR_EXP_1 = 2
		FACTOR_EXP_2 = 2
		PERSISTENCE = 0.7
	)

	for y := 0; y < tile_height; y++ {
		for x := 0; x < tile_width; x++ {
			source_noise[y][x] = math.Pow(rand.Float64(), FACTOR_EXP_1)
		}
	}

	var MAX_OCTAVE int = int(math.Floor(math.Log2(math.Min(float64(tile_width), float64(tile_height)))) + 1)
	var amplitude float64 = 1.0
	for octave := 2; octave <= MAX_OCTAVE; octave++ {
		smoothed_source := generate_smooth_noise(source_noise, int(math.Round(math.Pow(2, float64(MAX_OCTAVE - octave)))))
		for y := 0; y < tile_height; y++ {
			for x := 0; x < tile_width; x++ {
				region[y][x] += amplitude * smoothed_source[y][x]
			}
		}
		amplitude *= PERSISTENCE
	}
	for y := 0; y < tile_height; y++ {
		for x := 0; x < tile_width; x++ {
			region[y][x] += amplitude * source_noise[y][x]
		}
	}

	// Make productions spikier using exponential. Also find max value.
	var max_value float64
	for y := 0; y < tile_height; y++ {
		for x := 0; x < tile_width; x++ {
			region[y][x] = math.Pow(region[y][x], FACTOR_EXP_2)
			if region[y][x] > max_value { max_value = region[y][x] }
		}
	}

	// Normalize to highest value.

	MAX_CELL_PRODUCTION := max_production

	for y := 0; y < tile_height; y++ {
		for x := 0; x < tile_width; x++ {
			region[y][x] *= float64(MAX_CELL_PRODUCTION) / max_value
			tile[x][y] = int(region[y][x])								// Note the [x][y] and [y][x]
		}
	}

	return tile
}

func place_factories(frame *Frame, players int) {

	width := frame.Width()
	height := frame.Height()

	frame.dropoffs = nil

	for pid := 0; pid < players; pid++ {

		var x int
		var y int

		if pid == 0 {
			x = width / 4
			y = height / 4
		} else if pid == 1 {
			x = width - 1 - width / 4
			y = height / 4
		} else if pid == 2 {
			x = width / 4
			y = height - 1 - height / 4
		} else {
			x = width - 1 - width / 4
			y = height - 1 - height / 4
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
}
