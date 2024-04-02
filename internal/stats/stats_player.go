package stats

type playerStat struct {
	name    string
	getFunc func(primStats *PrimitiveStatline) any
	req     []PrimitiveStat
}

func (ps playerStat) getReq() []Stat {
	req := make([]Stat, 0)
	for _, s := range ps.req {
		req = append(req, s)
	}
	return req
}

func (ps playerStat) getName() string {
	return ps.name
}

type playerStatline struct {
	stats     map[string]playerStat
	primStats *PrimitiveStatline
	side      TeamSide
}

func (ps *playerStatline) get(stat playerStat) any {
	statStruct := ps.stats[stat.name]
	return statStruct.getFunc(ps.primStats)
}

func (ps *playerStatline) getAll() map[string]any {
	statline := make(map[string]any)
	for n, s := range ps.stats {
		statline[n] = s.getFunc(ps.primStats)
	}
	return statline
}

func newPlayerStatline(playerStats []playerStat, side TeamSide) playerStatline {
	statline := playerStatline{
		stats: make(map[string]playerStat),
		side:  side,
	}

	primReq := make(map[PrimitiveStat]bool)
	for _, s := range playerStats {
		for _, req := range s.req {
			primReq[req] = true
		}
		statline.stats[s.name] = s
	}

	primReqSl := make([]PrimitiveStat, 0)
	for req, _ := range primReq {
		primReqSl = append(primReqSl, req)
	}

	statline.primStats = newPrimitiveStatline(primReqSl)
	return statline
}
