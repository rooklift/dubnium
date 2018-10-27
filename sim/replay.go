package sim

import (
	"encoding/json"
	"fmt"
	"os"
)

const PRETTY_PRINT = false

type Replay struct {

	EngineVersion				string				`json:"ENGINE_VERSION"`
	Constants					*Constants			`json:"GAME_CONSTANTS"`
	FileVersion					int					`json:"REPLAY_FILE_VERSION"`
	FullFrames					[]*ReplayFrame		`json:"full_frames"`
	Stats						*ReplayStats		`json:"game_statistics"`
	Seed						int32				`json:"map_generator_seed"`
	NumPlayers					int					`json:"number_of_players"`
	Players						[]*ReplayPlayer		`json:"players"`
	ProductionMap				*ReplayMap			`json:"production_map"`

}

func NewReplay(names []string, game *Game, turns int, seed int32) *Replay {

	self := new(Replay)

	self.EngineVersion = "Dubnium Engine"
	self.Constants = game.Constants
	self.FileVersion = 3
	self.Seed = seed
	self.NumPlayers = game.frame.Players()

	for pid := 0; pid < self.NumPlayers; pid++ {
		player := NewReplayPlayer(names[pid], pid, game)
		self.Players = append(self.Players, player)
	}

	self.ProductionMap = ReplayMapFromFrame(game.frame)

	return self
}

func (self *Replay) Dump(filename string) {

	outfile, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer outfile.Close()

	enc := json.NewEncoder(outfile)

	if PRETTY_PRINT {
		enc.SetIndent("", "\t")			// Horrifically wasteful
	}

	err = enc.Encode(self)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}

type ReplayPlayer struct {												// This is created at start and not updated
	Energy						int			`json:"energy"`				// Always 5000
	Entities					[]bool		`json:"entities"`			// Unused
	FactoryLocation				Position	`json:"factory_location"`
	Name						string		`json:"name"`
	Pid							int			`json:"player_id"`
}

func NewReplayPlayer(name string, pid int, game *Game) *ReplayPlayer {

	self := new(ReplayPlayer)
	self.Energy = game.Constants.INITIAL_ENERGY
	self.Entities = make([]bool, 0)
	self.FactoryLocation = Position{game.frame.dropoffs[pid].X, game.frame.dropoffs[pid].Y}
	self.Name = name
	self.Pid = pid

	return self
}

type CellUpdate struct {
	Production				int							`json:"production"`
	X						int							`json:"x"`
	Y						int							`json:"y"`
}

type ReplayFrame struct {

	Cells					[]*CellUpdate				`json:"cells"`
	Deposited				map[int]int					`json:"deposited"`
	Energy					map[int]int					`json:"energy"`
	Entities				map[int]map[int]*Ship		`json:"entities"`
	Events					[]*ReplayEvent				`json:"events"`
	Moves					map[int][]*ReplayMove		`json:"moves"`

}

type ReplayEvent struct {

	// Not all of these are used in every event...
	// 3 types: spawn, shipwreck, construct

	Energy					int							`json:"energy"`
	Sid						int							`json:"id"`
	Location				*Position					`json:"location"`
	Owner					int							`json:"owner_id"`
	Type					string						`json:"type"`
	WreckedSids				[]int						`json:"ships"`
}

type ReplayMove struct {

	Direction				string						`json:"direction"`
	Sid						int							`json:"id"`
	Type					string						`json:"type"`
}

type ReplayStats struct {
	NumTurns				int							`json:"number_turns"`
	Pstats					[]*PlayerStats				`json:"player_statistics"`
}

type PlayerStats struct {

	AllCollisions			int							`json:"all_collisions"`
	AvgEntityDist			int							`json:"average_entity_distance"`
	FinalProduction			int							`json:"final_production"`
	HalitePerDropoff		[]*Dropoff					`json:"halite_per_dropoff"`		// a bit magical - the dropoff implements custom marshaler
	InteractionOpps			int							`json:"interaction_opportunities"`
	LastTurnAlive			int							`json:"last_turn_alive"`
	MaxEntityDist			int							`json:"max_entity_distance"`
	MiningEfficiency		float64						`json:"mining_efficiency"`
	NumDropoffs				int							`json:"number_dropoffs"`
	Pid						int							`json:"player_id"`
	RandomId				int							`json:"random_id"`
	Rank					int							`json:"rank"`
	SelfCollisions			int							`json:"self_collisions"`
	ShipsCaptured			int							`json:"ships_captured"`
	ShipsGiven				int							`json:"ships_given"`
	TotalBonus				int							`json:"total_bonus"`
	TotalMined				int							`json:"total_mined"`
	TotalMinedFromCap		int							`json:"total_mined_from_captured"`
	TotalProduction			int							`json:"total_production"`
}

type EnergyHolder struct {
	Energy					int							`json:"energy"`
}

type ReplayMap struct {
	Grid					[][]EnergyHolder			`json:"grid"`
	Width					int							`json:"width"`
	Height					int							`json:"height"`
}

func ReplayMapFromFrame(frame *Frame) *ReplayMap {

	self := new(ReplayMap)

	self.Width = frame.Width()
	self.Height = frame.Height()

	// Note y/x format...

	self.Grid = make([][]EnergyHolder, self.Height)

	for y := 0; y < self.Height; y++ {
		self.Grid[y] = make([]EnergyHolder, self.Width)
	}

	for y := 0; y < self.Height; y++ {
		for x := 0; x < self.Width; x++ {
			self.Grid[y][x].Energy = frame.halite[x][y]
		}
	}

	return self
}

func (d Dropoff) MarshalJSON() ([]byte, error) {		// Strictly for replay halite_per_dropoff stat.
	return []byte(fmt.Sprintf(`[{"x":%d,"y":%d},%d]`, d.X, d.Y, d.Gathered)), nil
}
