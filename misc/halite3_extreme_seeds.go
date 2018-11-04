package main

// This program is irrelevant for Dubnium itself, but uses
// Dubnium's mapgen to find extreme seeds.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"../sim"
)

type Result struct {
	Key			string
	Seed		uint32
	Size		int
	Players		int
	Score		int
}

func (self Result) String() string {
	halite_string := fmt.Sprintf("%v,", self.Score)		// For the comma with the padding
	return fmt.Sprintf("%9v: -s %-9v (halite: %-8v avg: %v)", self.Key, self.Seed, halite_string, self.Score / (self.Size * self.Size))
}

type SaveFile struct {
	N			uint32
	Seeds		[]uint32
}

var start_time = time.Now()
var end_chan = make(chan bool)

func main() {
	go runner()
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	end_chan <- true
	<- end_chan
}

func runner() {

	n := uint32(0)
	gens := 0

	results := make(map[string]Result)

	defer finish(&n, &gens, results)

	// -------------------------------------------------

	var load SaveFile

	injson, err := ioutil.ReadFile("known_seeds.json")

	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		err = json.Unmarshal(injson, &load)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}

	// -------------------------------------------------

	for _, n = range load.Seeds {

		size := sim.SizeFromSeed(n)

		for players := 2; players <= 4; players += 2 {

			high_key := fmt.Sprintf("%v-%v-%v", "high", players, size)
			low_key := fmt.Sprintf("%v-%v-%v", "low", players, size)

			frame := sim.MapGenOfficial(players, size, size, 5000, n)
			th := frame.TotalHalite()

			if th > results[high_key].Score {
				results[high_key] = Result{
					Key: high_key,
					Seed: n,
					Size: size,
					Players: players,
					Score: th,
				}
			}

			if th < results[low_key].Score || results[low_key].Key == "" {
				results[low_key] = Result{
					Key: low_key,
					Seed: n,
					Size: size,
					Players: players,
					Score: th,
				}
			}
		}
	}

	// -------------------------------------------------

	fmt.Printf("Starting at %v\n", load.N)

	for n = load.N; n < 0xffffffff; n++ {

		if n % 100 == 0 {
			select {
			case <- end_chan:
				return
			default:
				// pass
			}
		}

		gens++

		size := sim.SizeFromSeed(n)

		for players := 2; players <= 4; players += 2 {

			high_key := fmt.Sprintf("%v-%v-%v", "high", players, size)
			low_key := fmt.Sprintf("%v-%v-%v", "low", players, size)

			frame := sim.MapGenOfficial(players, size, size, 5000, n)
			th := frame.TotalHalite()

			if th > results[high_key].Score {
				results[high_key] = Result{
					Key: high_key,
					Seed: n,
					Size: size,
					Players: players,
					Score: th,
				}
				fmt.Printf("%v\n", results[high_key])
			}

			if th < results[low_key].Score || results[low_key].Key == "" {
				results[low_key] = Result{
					Key: low_key,
					Seed: n,
					Size: size,
					Players: players,
					Score: th,
				}
				fmt.Printf("%v\n", results[low_key])
			}
		}
	}
}

func finish(n *uint32, gens *int, results map[string]Result) {
	var all_keys []string
	var good_seeds []uint32
	var save SaveFile

	for key, _ := range results {
		all_keys = append(all_keys, key)
	}

	sort.Strings(all_keys)

	elapsed_seconds := time.Now().Sub(start_time) / time.Second
	seeds_per_second := *gens / int(elapsed_seconds)

	fmt.Printf("Ending at %d - seeds per second: %d\n", *n, seeds_per_second)
	fmt.Printf("--------------------------------------------------------------\n")

	for _, key := range all_keys {
		fmt.Printf("%v\n", results[key])
		good_seeds = append(good_seeds, results[key].Seed)
	}

	save.N = *n
	save.Seeds = good_seeds

	outjson, _ := json.Marshal(save)

	outfile, _ := os.Create("known_seeds.json")
	outfile.Write(outjson)
	outfile.Close()

	end_chan <- true
}
