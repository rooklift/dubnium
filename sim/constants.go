package sim

type Constants struct {

	CAPTURE_ENABLED				bool
	CAPTURE_RADIUS				int
	DEFAULT_MAP_HEIGHT			int
	DEFAULT_MAP_WIDTH			int
	DROPOFF_COST				int
	DROPOFF_PENALTY_RATIO		int
	EXTRACT_RATIO				int
	FACTOR_EXP_1				float64
	FACTOR_EXP_2				float64
	INITIAL_ENERGY				int
	INSPIRATION_ENABLED			bool
	INSPIRATION_RADIUS			int
	INSPIRATION_SHIP_COUNT		int
	INSPIRED_BONUS_MULTIPLIER	float64
	INSPIRED_EXTRACT_RATIO		int
	INSPIRED_MOVE_COST_RATIO	int
	MAX_CELL_PRODUCTION			int
	MAX_ENERGY					int
	MAX_PLAYERS					int
	MAX_TURNS					int
	MAX_TURN_THRESHOLD			int
	MIN_CELL_PRODUCTION			int
	MIN_TURNS					int
	MIN_TURN_THRESHOLD			int
	MOVE_COST_RATIO				int
	NEW_ENTITY_ENERGY_COST		int
	PERSISTENCE					float64
	SHIPS_ABOVE_FOR_CAPTURE		int
	STRICT_ERRORS				bool

	GameSeed					int32				`json:"game_seed"`		// Sent to bots but (in official) not to replay.
}

func NewConstants(players, width, height, turns int, seed int32) *Constants {

	return &Constants{

		CAPTURE_ENABLED:			false,
		CAPTURE_RADIUS:				3,
		DEFAULT_MAP_HEIGHT:			height,
		DEFAULT_MAP_WIDTH:			width,
		DROPOFF_COST:				4000,
		DROPOFF_PENALTY_RATIO:		4,			// What's this?
		EXTRACT_RATIO:				4,
		FACTOR_EXP_1:				2,			// What's this?
		FACTOR_EXP_2:				2,			// What's this?
		INITIAL_ENERGY:				5000,
		INSPIRATION_ENABLED:		true,
		INSPIRATION_RADIUS:			4,
		INSPIRATION_SHIP_COUNT:		2,
		INSPIRED_BONUS_MULTIPLIER:	2,
		INSPIRED_EXTRACT_RATIO:		4,
		INSPIRED_MOVE_COST_RATIO:	10,
		MAX_CELL_PRODUCTION:		1000,
		MAX_ENERGY:					1000,
		MAX_PLAYERS:				16,
		MAX_TURNS:					turns,
		MAX_TURN_THRESHOLD:			64,			// What's this?
		MIN_CELL_PRODUCTION:		900,
		MIN_TURNS:					400,
		MIN_TURN_THRESHOLD:			32,			// What's this?
		MOVE_COST_RATIO:			10,
		NEW_ENTITY_ENERGY_COST:		1000,
		PERSISTENCE:				0.7,		// What's this?
		SHIPS_ABOVE_FOR_CAPTURE:	3,
		STRICT_ERRORS:				false,

		GameSeed:					seed,
	}
}
