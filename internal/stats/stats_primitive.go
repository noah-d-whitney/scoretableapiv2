package stats

import (
	"fmt"
	"sync"
)

// Stat interface is typically implemented by a struct that contains a string name, a getFunc,
// which calculates stat value, and a req slice, which contains all Stat's required to calculate.
type Stat interface {
	getReq() []Stat
	getName() string
}

// getPrimitiveStats recursively traverses the reqs of a list of Stat's and returns a slice of
// PrimitiveStat's required.
func getPrimitiveStats(stats []Stat) []PrimitiveStat {
	fmt.Printf("LEN: %d || ", len(stats))
	fmt.Printf("TYPE: %+v\n", stats[0])
	switch stats[0].(type) {
	case PrimitiveStat:
		//primStatsMap := make(map[PrimitiveStat]bool)
		//for _, s := range stats {
		//	primStatsMap[s.(PrimitiveStat)] = true
		//}
		//
		primStats := make([]PrimitiveStat, 0)
		for _, s := range stats {
			primStats = append(primStats, s.(PrimitiveStat))
		}

		return primStats
	default:
		reqStats := make(map[string]Stat)
		for _, s := range stats {
			req := s.getReq()
			for _, r := range req {
				reqStats[r.getName()] = r
			}
		}
		reqStatsSl := make([]Stat, 0)
		for _, s := range reqStats {
			reqStatsSl = append(reqStatsSl, s)
		}

		reqStatsStr := make([]string, 0)
		for _, s := range reqStatsSl {
			reqStatsStr = append(reqStatsStr, s.getName())
		}
		fmt.Printf("%v\n", reqStatsStr)
		return getPrimitiveStats(reqStatsSl)
	}
}

// PrimitiveStat is a string type to define keys of map in PrimitiveStatline
type PrimitiveStat string

func (ps PrimitiveStat) getReq() []Stat {
	return make([]Stat, 0)
}

func (ps PrimitiveStat) getName() string {
	return string(ps)
}

// PrimitiveStatline holds a map with keys of type PrimitiveStat and value of type int. Int value
// holds current value of stat.
type PrimitiveStatline struct {
	stats map[PrimitiveStat]int // DO NOT access stats map directly. Instead,
	// use get on PrimitiveStatline
	mu sync.Mutex
}

// get(): gets int value for key PrimitiveStat
func (psl *PrimitiveStatline) get(stat PrimitiveStat) int {
	return psl.stats[stat]
}

// set(): locks memory and adds int provided to value for key PrimitiveStat in PrimitiveStatline.
// Returns new value.
func (psl *PrimitiveStatline) set(stat PrimitiveStat, add int) int {
	currentValue := psl.get(stat)
	if currentValue+add < 0 {
		return currentValue
	}
	psl.mu.Lock()
	psl.stats[stat] += add
	psl.mu.Unlock()
	return psl.get(stat)
}

// newPrimitiveStatline receives a slice of PrimitiveStat's,
// returns a pointer to a PrimitiveStatline with initialized map of keys of provided
// PrimitiveStat's and values of 0.
func newPrimitiveStatline(primStats []PrimitiveStat) *PrimitiveStatline {
	statline := PrimitiveStatline{
		stats: make(map[PrimitiveStat]int),
		mu:    sync.Mutex{},
	}
	for _, s := range primStats {
		statline.stats[s] = 0
	}
	return &statline
}
