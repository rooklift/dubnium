package sim

func string_to_dxdy(s string) (int, int) {

	switch s {

	case "e":
		return 1, 0
	case "w":
		return -1, 0
	case "s":
		return 0, 1
	case "n":
		return 0, -1
	case "c":
		return 0, 0
	case "o":
		return 0, 0
	case "":
		return 0, 0
	}

	panic("string_to_dxdy() got illegal string")
}

func mod(x, n int) int {

	// Works for negative x
	// https://dev.to/maurobringolf/a-neat-trick-to-compute-modulo-of-negative-numbers-111e

	return (x % n + n) % n
}

func make_2d_float_array(width, height int) [][]float64 {
	ret := make([][]float64, width)
	for x := 0; x < width; x++ {
		ret[x] = make([]float64, height)
	}
	return ret
}

func make_2d_int_array(width, height int) [][]int {
	ret := make([][]int, width)
	for x := 0; x < width; x++ {
		ret[x] = make([]int, height)
	}
	return ret
}
