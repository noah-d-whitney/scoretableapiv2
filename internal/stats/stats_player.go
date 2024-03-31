package stats

type PlayerStat struct {
	name    string
	getFunc func(primStats *PrimitiveStatline) any
	req     []PrimitiveStat
}

func (ps PlayerStat) getReq() []Stat {
	return make([]Stat, 0)
}

type PlayerStatline struct {
	stats     map[string]PlayerStat
	primStats *PrimitiveStatline
}

func (ps *PlayerStatline) get(stat PlayerStat) any {
	statStruct := ps.stats[stat.name]
	return statStruct.getFunc(ps.primStats)
}

func (ps *PlayerStatline) getAll() map[string]any {
	statline := make(map[string]any)
	for n, s := range ps.stats {
		statline[n] = s.getFunc(ps.primStats)
	}
	return statline
}

func newPlayerStatline(playerStats []PlayerStat) (PlayerStatline, error) {
	statline := PlayerStatline{
		stats: make(map[string]PlayerStat),
	}

	primReq := make(map[PrimitiveStat]bool)
	for _, s := range playerStats {
		for _, req := range s.req {
			_, exists := primReq[req]
			if exists {
				return PlayerStatline{}, ErrDuplicateStatKeys
			}
			primReq[req] = true
		}
		statline.stats[s.name] = s
	}

	primReqSl := make([]PrimitiveStat, 0)
	for req, _ := range primReq {
		primReqSl = append(primReqSl, req)
	}

	statline.primStats = newPrimitiveStatline(primReqSl)
	return statline, nil
}
