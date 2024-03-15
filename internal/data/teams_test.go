package data

import (
	"ScoreTableApi/internal/assert"
	"testing"
)

func TestParsePlayerAsgList(t *testing.T) {
	tests := []struct {
		name         string
		idList       []int64
		assignWant   []int64
		unassignWant []int64
	}{
		{
			name:         "Parse Assignment",
			idList:       []int64{1, 2, 3, 4, 5},
			assignWant:   []int64{1, 2, 3, 4, 5},
			unassignWant: []int64{},
		},
		{
			name:         "Parse Unassignment",
			idList:       []int64{-1, -2, -3, -4, -5},
			assignWant:   []int64{},
			unassignWant: []int64{1, 2, 3, 4, 5},
		},
		{
			name:         "Parse Assign & Unassign",
			idList:       []int64{1, 2, 3, -4, -5},
			assignWant:   []int64{1, 2, 3},
			unassignWant: []int64{4, 5},
		},
		{
			name:         "Parse With Zero",
			idList:       []int64{0, 1, 2, 3, 4},
			assignWant:   []int64{1, 2, 3, 4},
			unassignWant: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as, ua := parsePlayerAsgList(&Team{PlayerIDs: tt.idList})
			assert.Int64SliceEqual(t, as, tt.assignWant)
			assert.Int64SliceEqual(t, ua, tt.unassignWant)
		})
	}
}
