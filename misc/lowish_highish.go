package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"../sim"
)

var end_chan = make(chan bool)

func main() {
	go runner()
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	end_chan <- true
	<- end_chan
}

func runner() {

	results := make(map[string][]int)

	var n uint32

	Loop:
	for {

		if n % 100 == 0 {
			select {
			case <- end_chan:
				break Loop
			default:
				// pass
			}
		}

		size := sim.SizeFromSeed(n)

		for players := 2; players <= 4; players += 2 {
			frame := sim.MapGenOfficial(players, size, size, 5000, n)
			th := frame.TotalHalite()
			s := fmt.Sprintf("%d-%d", players, size)
			results[s] = append(results[s], th)
		}

		n++
	}

	fmt.Printf("n = %d\n\n", n)

	for size := 32; size <= 64; size += 8 {

		for players := 2; players <= 4; players += 2 {

			s := fmt.Sprintf("%d-%d", players, size)
			sort.Ints(results[s])

			half := len(results[s]) / 2
			third := len(results[s]) / 3
			quarter := len(results[s]) / 4

			fmt.Printf("%s: quartile: %d, tertile: %d, median: %d, tertile: %d, quartile: %d\n", s,
				results[s][quarter],
				results[s][third],
				results[s][half],
				results[s][2 * third],
				results[s][3 * quarter],
			)
		}
	}

	end_chan <- true
}
