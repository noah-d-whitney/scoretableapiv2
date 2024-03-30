package stats

type PlayerStat struct {
	name    string
	getFunc func(primStats *PrimitiveStatline) any
	req     []PrimitiveStat
}

func (ps PlayerStat) getReq() []Stat {
	return make([]Stat, 0)
}

var (
	PlayerPoints = PlayerStat{
		name: "Pts",
		getFunc: func(primStats *PrimitiveStatline) any {
			var points int
			points += primStats.stats[FreeThrowMade]
			points += primStats.stats[TwoPointMade] * 2
			points += primStats.stats[ThreePointMade] * 3
			return points
		},
		req: []PrimitiveStat{FreeThrowMade, TwoPointMade, ThreePointMade},
	}
	PlayerTwoPointsAttempted = PlayerStat{
		name: "2PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.stats[TwoPointMiss] + primStats.stats[TwoPointMade]
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade},
	}
	PlayerTwoPointsMade = PlayerStat{
		name: "2PtM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.stats[TwoPointMade]
		},
		req: []PrimitiveStat{TwoPointMade},
	}
	PlayerThreePointsAttempted = PlayerStat{
		name: "3PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.stats[ThreePointMiss] + primStats.stats[ThreePointMade]
		},
		req: []PrimitiveStat{ThreePointMiss, ThreePointMade},
	}
	PlayerThreePointsMade = PlayerStat{
		name: "3PtM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.stats[ThreePointMade]
		},
		req: []PrimitiveStat{ThreePointMade},
	}
	PlayerFieldGoalsAttempted = PlayerStat{
		name: "FGA",
		getFunc: func(primStats *PrimitiveStatline) any {
			var fga int
			fga += primStats.stats[TwoPointMiss]
			fga += primStats.stats[TwoPointMade]
			fga += primStats.stats[ThreePointMiss]
			fga += primStats.stats[ThreePointMade]
			return fga
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade, ThreePointMiss, ThreePointMade},
	}
	PlayerFieldGoalsMade = PlayerStat{
		name: "FGM",
		getFunc: func(primStats *PrimitiveStatline) any {
			var fgm int
			fgm += primStats.stats[TwoPointMade]
			fgm += primStats.stats[ThreePointMade]
			return fgm
		},
		req: []PrimitiveStat{TwoPointMade, ThreePointMade},
	}
)

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

func newPlayerStatline(playerStats []PlayerStat) PlayerStatline {
	statline := PlayerStatline{
		stats: make(map[string]PlayerStat),
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
