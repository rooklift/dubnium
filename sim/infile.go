package sim

import (
	"encoding/json"
	"io/ioutil"
	"os"
)


type StuffWeWant struct {
	Players					[]*ReplayPlayer		`json:"players"`
	ProductionMap			*ReplayMap			`json:"production_map"`
	Seed					int64				`json:"map_generator_seed"`
}


func FrameFromFile(infile string) (*Frame, int64) {

	f, err := os.Open(infile)
	if err != nil {
		panic("Couldn't read infile")
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)

	if err != nil {
		panic("Couldn't read infile")
	}

	foo := new(StuffWeWant)

	err = json.Unmarshal(bytes, foo)

	if err != nil {
		panic("Couldn't parse infile")
	}

	players := len(foo.Players)
	width := foo.ProductionMap.Width
	height := foo.ProductionMap.Height

	frame := new(Frame)

	for pid := 0; pid < players; pid++ {
		frame.budgets = append(frame.budgets, 5000)						// FIXME: don't hardcode
		frame.deposited = append(frame.deposited, 0)
	}

	frame.halite = make([][]int, width)

	for x := 0; x < width; x++ {
		frame.halite[x] = make([]int, height)
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			frame.halite[x][y] = foo.ProductionMap.Grid[y][x].Energy	// y/x inversion
		}
	}

	for pid := 0; pid < players; pid++ {

		x := foo.Players[pid].FactoryLocation.X
		y := foo.Players[pid].FactoryLocation.Y

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

	return frame, foo.Seed
}


