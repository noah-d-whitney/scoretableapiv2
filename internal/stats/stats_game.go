package stats

type gameTeamsStatline struct {
	home teamStatline
	away teamStatline
}

// GameStat is implements Stat and makes calculations with teamStat type.
type GameStat struct {
	name    string
	getFunc func(gameTeamsStats gameTeamsStatline) any
	req     []teamStat
}

func (gs GameStat) getReq() []Stat {
	req := make([]Stat, 0)
	for _, r := range gs.req {
		req = append(req, r)
	}
	return req
}

func (gs GameStat) getName() string {
	return gs.name
}

type GameTeamSide int

const (
	TeamHome GameTeamSide = iota
	TeamAway
)

// GameStatline is a struct that contains a map of PrimitiveStat's at its root,
// and layers of Stat's over top. Stat's are simply functions that calculate values based on layer
// of Stat's beneath. For example, PrimitiveStat's are used to calculate playerStats,
// which are used to calculate teamStats. All writes to statline are done to PrimitiveStat types,
// while all reads are from Stat types.
type GameStatline struct {
	stats       map[string]GameStat
	teamStats   gameTeamsStatline
	playerStats map[string]*playerStatline
}

// Add receives a playerPin, PrimitiveStat,
// and int. Adds value of add arg to primitive statline for provided playerID
func (gsl *GameStatline) Add(playerPin string, stat PrimitiveStat, add int) int {
	statline := gsl.playerStats[playerPin]
	newValue := statline.primStats.set(stat, add)
	return newValue
}

// GetDtoFromPrimitive return a GameStatlineDto containing only Stat's that are dependent
// on provided PrimitiveStat.
func (gsl *GameStatline) GetDtoFromPrimitive(playerPin string, stat PrimitiveStat) GameStatlineDto {
	statline := GameStatlineDto{}
	playerStl := statlineDto{}
	teamStl := statlineDto{}
	gameSl := statlineDto{}

	playerStats := gsl.playerStats[playerPin]
	playerStatsKept := make([]playerStat, 0)
	for _, s := range playerStats.stats {
		reqs := assertAndCopyStatsToMap[PrimitiveStat](s.getReq())
		if _, ok := reqs[stat.getName()]; ok {
			playerStl[s.name] = gsl.getPlayerStat(s, playerPin)
			playerStatsKept = append(playerStatsKept, s)
		}
	}

	var teamStats teamStatline
	switch playerStats.side {
	case TeamHome:
		teamStats = gsl.teamStats.home
	case TeamAway:
		teamStats = gsl.teamStats.away
	}
	teamStatsKept := make([]teamStat, 0)
	for _, s := range teamStats.stats {
		reqs := assertAndCopyStatsToMap[playerStat](s.getReq())
		for _, ps := range playerStatsKept {
			if _, ok := reqs[ps.getName()]; ok {
				teamStl[s.name] = gsl.getTeamStat(s, playerStats.side)
				teamStatsKept = append(teamStatsKept, s)
			}
		}
	}

	gameStats := gsl.stats
	for _, s := range gameStats {
		reqs := assertAndCopyStatsToMap[teamStat](s.getReq())
		for _, ts := range teamStatsKept {
			if _, ok := reqs[ts.getName()]; ok {
				gameSl[s.name] = gsl.getGameStat(s)
			}
		}
	}

	statline.GameStats = gameSl
	switch playerStats.side {
	case TeamHome:
		statline.Teams.Home.TeamStats = teamStl
		statline.Teams.Home.PlayerStats = make(map[string]statlineDto)
		statline.Teams.Home.PlayerStats[playerPin] = playerStl
	case TeamAway:
		statline.Teams.Away.TeamStats = teamStl
		statline.Teams.Away.PlayerStats = make(map[string]statlineDto)
		statline.Teams.Away.PlayerStats[playerPin] = playerStl
	}

	return statline
}

// GetPrimitiveStats returns a slice of all PrimitiveStat's contained in GameStatline.
func (gsl *GameStatline) GetPrimitiveStats() []PrimitiveStat {
	gameStats := make([]Stat, 0)
	for _, s := range gsl.stats {
		gameStats = append(gameStats, s)
	}

	print(len(gameStats))

	return getPrimitiveStats(gameStats)
}

// GetDto executes all Stat's in GameStatline and returns a GameStatlineDto.
func (gsl *GameStatline) GetDto() GameStatlineDto {
	cleanStatline := GameStatlineDto{}
	cleanStatline.GameStats = gsl.getAll()
	cleanStatline.Teams.Home.TeamStats = gsl.teamStats.home.getAll()
	cleanStatline.Teams.Away.TeamStats = gsl.teamStats.away.getAll()

	homePlayerStats := make(map[string]statlineDto)
	for p, s := range gsl.teamStats.home.playerStats {
		homePlayerStats[p] = s.getAll()
	}
	cleanStatline.Teams.Home.PlayerStats = homePlayerStats

	awayPlayerStats := make(map[string]statlineDto)
	for p, s := range gsl.teamStats.away.playerStats {
		awayPlayerStats[p] = s.getAll()
	}
	cleanStatline.Teams.Away.PlayerStats = awayPlayerStats

	return cleanStatline
}

// NewGameStatline returns a pointer to a GameStatline with specified GameStatlineBlueprint and
// player pins.
func NewGameStatline(homePlayerPins, awayPlayerPins []string, blueprint GameStatlineBlueprint) *GameStatline {
	statline := GameStatline{
		stats: make(map[string]GameStat),
	}

	// add each stat's requirements to map and assign to stats map
	teamStatsReq := make(map[string]teamStat)
	for _, s := range blueprint {
		for _, req := range s.req {
			teamStatsReq[req.name] = req
		}
		statline.stats[s.name] = s
	}

	teamStatsReqSl := make([]teamStat, 0)
	for _, req := range teamStatsReq {
		teamStatsReqSl = append(teamStatsReqSl, req)
	}

	// create team statlines with each slice of player ids
	gameTeamsStl := gameTeamsStatline{
		home: newTeamStatline(homePlayerPins, TeamHome, teamStatsReqSl),
		away: newTeamStatline(awayPlayerPins, TeamAway, teamStatsReqSl),
	}
	statline.teamStats = gameTeamsStl

	// create map of player pins and pointers to their respective player statlines
	primStats := make(map[string]*playerStatline)
	for p, s := range gameTeamsStl.home.playerStats {
		primStats[p] = &s
	}
	for p, s := range gameTeamsStl.away.playerStats {
		primStats[p] = &s
	}
	statline.playerStats = primStats

	return &statline
}

// GameStatlineDto is a GameStatline with all Stat's executed and returned as a struct with no
// functionality.
type GameStatlineDto struct {
	GameStats statlineDto `json:"game_stats"`
	Teams     struct {
		Home struct {
			TeamStats   statlineDto            `json:"team_stats,omitempty"`
			PlayerStats map[string]statlineDto `json:"player_stats,omitempty"`
		} `json:"home,omitempty"`
		Away struct {
			TeamStats   statlineDto            `json:"team_stats,omitempty"`
			PlayerStats map[string]statlineDto `json:"player_stats,omitempty"`
		} `json:"away,omitempty"`
	} `json:"teams"`
}

func (gsl *GameStatline) getAll() map[string]any {
	statline := make(map[string]any)
	for n, s := range gsl.stats {
		statline[n] = s.getFunc(gsl.teamStats)
	}
	return statline
}

func (gsl *GameStatline) getGameStat(stat GameStat) any {
	gameStat := gsl.stats[stat.name]
	return gameStat.getFunc(gsl.teamStats)
}

func (gsl *GameStatline) getTeamStat(stat teamStat, side GameTeamSide) any {
	var teamStatline teamStatline
	switch side {
	case TeamHome:
		teamStatline = gsl.teamStats.home
	case TeamAway:
		teamStatline = gsl.teamStats.away
	}

	return teamStatline.get(stat)
}

func (gsl *GameStatline) getPlayerStat(stat playerStat, playerPin string) any {
	playerSt := gsl.playerStats[playerPin]
	return playerSt.get(stat)
}

type statlineDto map[string]any
