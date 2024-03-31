package stats

import (
	json2 "encoding/json"
	"errors"
	"fmt"
)

type Stat interface {
	getReq() []Stat
}

type GameTeamsStatline struct {
	home TeamStatline
	away TeamStatline
}

type GameStat struct {
	name    string
	getFunc func(gameTeamsStats GameTeamsStatline) any
	req     []TeamStat
}

func (gs GameStat) getReq() []Stat {
	req := make([]Stat, 0)
	for _, r := range gs.req {
		req = append(req, r)
	}
	return req
}

type GameTeamSide int

const (
	TeamHome GameTeamSide = iota
	TeamAway
)

type GameStatline struct {
	stats       map[string]GameStat
	teamStats   GameTeamsStatline
	playerStats map[string]*PlayerStatline
}

func (gsl *GameStatline) Add(playerPin string, stat PrimitiveStat, add int) int {
	statline := gsl.playerStats[playerPin]
	newValue := statline.primStats.set(stat, add)
	return newValue
}

func (gsl *GameStatline) GetGameStat(stat GameStat) any {
	gameStat := gsl.stats[stat.name]
	return gameStat.getFunc(gsl.teamStats)
}

func (gsl *GameStatline) getAll() map[string]any {
	statline := make(map[string]any)
	for n, s := range gsl.stats {
		statline[n] = s.getFunc(gsl.teamStats)
	}
	return statline
}

func (gsl *GameStatline) GetTeamStat(stat TeamStat, side GameTeamSide) any {
	var teamStatline TeamStatline
	switch side {
	case TeamHome:
		teamStatline = gsl.teamStats.home
	case TeamAway:
		teamStatline = gsl.teamStats.away
	}

	return teamStatline.get(stat)
}

func (gsl *GameStatline) GetCleanStatline() CleanGameStatline {
	cleanStatline := CleanGameStatline{}
	cleanStatline.GameStats = gsl.getAll()
	cleanStatline.Teams.Home.TeamStats = gsl.teamStats.home.getAll()
	cleanStatline.Teams.Away.TeamStats = gsl.teamStats.away.getAll()

	homePlayerStats := make(map[string]CleanStatline)
	for p, s := range gsl.teamStats.home.playerStats {
		homePlayerStats[p] = s.getAll()
	}
	cleanStatline.Teams.Home.PlayerStats = homePlayerStats

	awayPlayerStats := make(map[string]CleanStatline)
	for p, s := range gsl.teamStats.away.playerStats {
		awayPlayerStats[p] = s.getAll()
	}
	cleanStatline.Teams.Away.PlayerStats = awayPlayerStats

	return cleanStatline
}

func (gsl *GameStatline) GetPlayerStat(stat PlayerStat, playerPin string) any {
	playerStat := gsl.playerStats[playerPin]
	return playerStat.get(stat)
}

//
//func (gsl *GameStatline) getAllowedStats(stats []Stat) []Stat {
//	allowedStats := make([]Stat, 0)
//	for _, s := range stats {
//		allowedStats = append(allowedStats, s)
//		if len(s.getReq()) != 0 {
//			return gsl.getAllowedStats(s.getReq())
//		}
//	}
//
//	return allowedStats
//}

var (
	ErrDuplicateStatKeys = errors.New("duplicate stat key in statline blueprint")
)

func newGameStatline(homePlayerPins, awayPlayerPins []string,
	blueprint GameStatlineBlueprint) (*GameStatline, error) {
	statline := GameStatline{
		stats: make(map[string]GameStat),
	}

	// add each stat's requirements to map and assign to stats map
	teamStatsReq := make(map[string]TeamStat)
	for _, s := range blueprint {
		for _, req := range s.req {
			_, exists := teamStatsReq[req.name]
			if exists {
				return nil, ErrDuplicateStatKeys
			}
			teamStatsReq[req.name] = req
		}
		statline.stats[s.name] = s
	}

	teamStatsReqSl := make([]TeamStat, 0)
	for _, req := range teamStatsReq {
		teamStatsReqSl = append(teamStatsReqSl, req)
	}

	// create team statlines with each slice of player ids
	homeSl, err := newTeamStatline(homePlayerPins, teamStatsReqSl)
	if err != nil {
		return nil, err
	}
	awaySl, err := newTeamStatline(awayPlayerPins, teamStatsReqSl)
	if err != nil {
		return nil, err
	}
	gameTeamsStatline := GameTeamsStatline{
		home: homeSl,
		away: awaySl,
	}
	statline.teamStats = gameTeamsStatline

	// create map of player pins and pointers to their respective player statlines
	primStats := make(map[string]*PlayerStatline)
	for p, s := range gameTeamsStatline.home.playerStats {
		primStats[p] = &s
	}
	for p, s := range gameTeamsStatline.away.playerStats {
		primStats[p] = &s
	}
	statline.playerStats = primStats

	return &statline, nil
}

type CleanGameStatline struct {
	GameStats CleanStatline `json:"game_stats"`
	Teams     struct {
		Home struct {
			TeamStats   CleanStatline            `json:"team_stats"`
			PlayerStats map[string]CleanStatline `json:"player_stats"`
		} `json:"home"`
		Away struct {
			TeamStats   CleanStatline            `json:"team_stats"`
			PlayerStats map[string]CleanStatline `json:"player_stats"`
		} `json:"away"`
	} `json:"teams"`
}
type CleanStatline map[string]any

func TEST() {
	gameStats := []GameStat{GamePoints, GameFieldGoalsMade, GameFieldGoalsAttempted}
	homePins := []string{"noah", "aaron", "jesse"}
	awayPins := []string{"waylon", "felixx", "lexie"}
	gsl := newGameStatline(homePins, awayPins, gameStats)

	json, _ := json2.MarshalIndent(gsl.GetCleanStatline(), "", "\t")
	statline := string(json)
	fmt.Printf("Start: %s\n", statline)

	gsl.Add("noah", ThreePointMade, 1)
	gsl.Add("noah", ThreePointMiss, 1)
	gsl.Add("noah", ThreePointMade, 1)
	gsl.Add("waylon", TwoPointMade, 1)
	gsl.Add("lexie", ThreePointMade, 1)

	json, _ = json2.MarshalIndent(gsl.GetCleanStatline(), "", "\t")
	statline = string(json)
	fmt.Printf("AFTER: %s\n", statline)
}
