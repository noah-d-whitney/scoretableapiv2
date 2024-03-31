package stats

type TeamPlayersStatline map[string]PlayerStatline

func newTeamPlayersStatline(playerPins []string, playerStats []PlayerStat) (TeamPlayersStatline,
	error) {
	teamPlayersStatline := TeamPlayersStatline{}
	for _, pin := range playerPins {
		psl, err := newPlayerStatline(playerStats)
		if err != nil {
			return TeamPlayersStatline{}, err
		}
		teamPlayersStatline[pin] = psl
	}
	return teamPlayersStatline, nil
}

type TeamStat struct {
	name    string
	getFunc func(teamPlayersStats TeamPlayersStatline) any
	req     []PlayerStat
}

func (ts TeamStat) getReq() []Stat {
	req := make([]Stat, 0)
	for _, r := range ts.req {
		req = append(req, r)
	}
	return req
}

type TeamStatline struct {
	stats       map[string]TeamStat
	playerStats TeamPlayersStatline
}

func (ps *TeamStatline) get(stat TeamStat) any {
	teamStat := ps.stats[stat.name]
	return teamStat.getFunc(ps.playerStats)
}
func (ps *TeamStatline) getAll() map[string]any {
	statline := make(map[string]any)
	for n, s := range ps.stats {
		statline[n] = s.getFunc(ps.playerStats)
	}
	return statline
}

func newTeamStatline(playerPins []string, teamStats []TeamStat) (TeamStatline, error) {
	statline := TeamStatline{
		stats: make(map[string]TeamStat),
	}

	playerStatsReq := make(map[string]PlayerStat)
	for _, s := range teamStats {
		for _, req := range s.req {
			_, exists := playerStatsReq[req.name]
			if exists {
				return TeamStatline{}, ErrDuplicateStatKeys
			}
			playerStatsReq[req.name] = req
		}
		statline.stats[s.name] = s
	}

	playerStatsReqSl := make([]PlayerStat, 0)
	for _, req := range playerStatsReq {
		playerStatsReqSl = append(playerStatsReqSl, req)
	}

	teamPlayersSl, err := newTeamPlayersStatline(playerPins, playerStatsReqSl)
	if err != nil {
		return TeamStatline{}, err
	}
	statline.playerStats = teamPlayersSl
	return statline, nil
}
