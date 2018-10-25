package sim

import "./mt19937_32"

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
)

func MapGenOfficial(players, width, height, player_energy int, seed int32) *Frame {

	// AFAIK, perfectly faithful reproduction of the official mapgen in file:
	// game_engine/mapgen/FractalValueNoiseTileGenerator.cpp

	// For the RNG, probably I need to reproduce std::uniform_real_distribution

	mt19937_32.Seed(uint32(seed))

	frame := new(Frame)

	for pid := 0; pid < players; pid++ {
		frame.budgets = append(frame.budgets, player_energy)
		frame.deposited = append(frame.deposited, 0)
	}

	frame.halite = make_2d_int_array(width, height)

	// ----------------------------------------------------------------------------------------------------

	tile_width := width
	tile_height := height

	tile_cols := 1
	tile_rows := 1

	if players > 1 {
		tile_width = width / 2
		tile_cols = 2
	}

	if players > 2 {
		tile_height = height / 2
		tile_rows = 2
	}

	if width % 2 == 1 && tile_cols >= 2 { tile_width += 1 }
	if height % 2 == 1 && tile_rows >= 2 { tile_height += 1 }

	tile := make_tile(tile_width, tile_height)

	for x := 0; x < tile_width; x++ {

		for y := 0; y < tile_height; y++ {

			val := tile[x][y]

			frame.halite[x][y] = val

			if tile_cols > 1 { frame.halite[width - 1 - x][y] = val }

			if tile_rows > 1 {
				frame.halite[x][height - 1 - y] = val
				frame.halite[width - 1 - x][height - 1 - y] = val
			}
		}
	}

	place_factories(frame, players, tile_width, tile_height)

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

func make_tile(tile_width, tile_height int) [][]int {

	// Although various things here use [y][x] format, the tile itself is in our normal [x][y]...

	tile := make_2d_int_array(tile_width, tile_height)

	source_noise := make_2d_float_array(tile_height, tile_width)
	region := make_2d_float_array(tile_height, tile_width)

	const (
		FACTOR_EXP_1 = 2
		FACTOR_EXP_2 = 2
		PERSISTENCE = 0.7
		MAX_CELL_PRODUCTION = 1000
		MIN_CELL_PRODUCTION = 900
	)

	for y := 0; y < tile_height; y++ {
		for x := 0; x < tile_width; x++ {
			source_noise[y][x] = math.Pow(mt19937_32.Urd(), FACTOR_EXP_1)
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

	actual_max := mt19937_32.Uint32() % (1 + MAX_CELL_PRODUCTION - MIN_CELL_PRODUCTION) + MIN_CELL_PRODUCTION

	for y := 0; y < tile_height; y++ {
		for x := 0; x < tile_width; x++ {
			region[y][x] *= float64(actual_max) / max_value
			tile[x][y] = int(region[y][x])								// Note the [x][y] and [y][x]
		}
	}

	// The function now makes 2 RNG calls which end up not being used,
	// because the factory location generated is ignored. We duplicate
	// the calls here to keep the RNG consistent...

	mt19937_32.Uint32()
	mt19937_32.Uint32()

	return tile
}

func place_factories(frame *Frame, players, tile_width, tile_height int) {

	width := frame.Width()
	height := frame.Height()

	frame.dropoffs = nil

	dx := tile_width / 2
	dy := tile_height / 2

	if tile_width >= 16 && tile_width <= 40 && tile_height >= 16 && tile_height <= 40 {
		dx = int(8.0 + (float64(tile_width - 16) / 24.0) * 20.0)
		if players > 2 {
			dy = int(8.0 + (float64(tile_height - 16) / 24.0) * 20.0)
		}
	}

	for pid := 0; pid < players; pid++ {

		var x int
		var y int

		if pid == 0 {
			x = dx
			y = dy
		} else if pid == 1 {
			x = width - 1 - dx
			y = dy
		} else if pid == 2 {
			x = dx
			y = height - 1 - dy
		} else {
			x = width - 1 - dx
			y = height - 1 - dy
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
