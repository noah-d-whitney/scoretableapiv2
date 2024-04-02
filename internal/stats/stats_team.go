package stats

type teamPlayersStatline map[string]playerStatline

func newTeamPlayersStatline(playerPins []string, side TeamSide,
	playerStats []playerStat) teamPlayersStatline {
	teamPlayersStl := teamPlayersStatline{}
	for _, pin := range playerPins {
		teamPlayersStl[pin] = newPlayerStatline(playerStats, side)
	}
	return teamPlayersStl
}

type teamStat struct {
	name    string
	getFunc func(teamPlayersStats teamPlayersStatline) any
	req     []playerStat
}

func (ts teamStat) getReq() []Stat {
	req := make([]Stat, 0)
	for _, r := range ts.req {
		req = append(req, r)
	}
	return req
}

func (ts teamStat) getName() string {
	return ts.name
}

type teamStatline struct {
	stats       map[string]teamStat
	playerStats teamPlayersStatline
}

func (ps *teamStatline) get(stat teamStat) any {
	teamStat := ps.stats[stat.name]
	return teamStat.getFunc(ps.playerStats)
}
func (ps *teamStatline) getAll() map[string]any {
	statline := make(map[string]any)
	for n, s := range ps.stats {
		statline[n] = s.getFunc(ps.playerStats)
	}
	return statline
}

func newTeamStatline(playerPins []string, side TeamSide, teamStats []teamStat) teamStatline {
	statline := teamStatline{
		stats: make(map[string]teamStat),
	}

	playerStatsReq := make(map[string]playerStat)
	for _, s := range teamStats {
		for _, req := range s.req {
			playerStatsReq[req.name] = req
		}
		statline.stats[s.name] = s
	}

	playerStatsReqSl := make([]playerStat, 0)
	for _, req := range playerStatsReq {
		playerStatsReqSl = append(playerStatsReqSl, req)
	}

	statline.playerStats = newTeamPlayersStatline(playerPins, side, playerStatsReqSl)
	return statline
}
