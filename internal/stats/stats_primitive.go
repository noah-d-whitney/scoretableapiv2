package stats

import "sync"

// PrimitiveStat is a string type to define keys of map in PrimitiveStatline
type PrimitiveStat string

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
