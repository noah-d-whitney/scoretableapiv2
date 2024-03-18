package data

import (
	"ScoreTableApi/internal/assert"
	"testing"
)

func TestParsePlayerAsgList(t *testing.T) {
	tests := []struct {
		name         string
		idList       []string
		assignWant   []string
		unassignWant []string
	}{
		{
			name:         "Parse Assignment",
			idList:       []string{"ABC", "EFG", "HIJ"},
			assignWant:   []string{"ABC", "EFG", "HIJ"},
			unassignWant: []string{},
		},
		{
			name:         "Parse Unassignment",
			idList:       []string{"-ABC", "-EFG", "-HIJ"},
			assignWant:   []string{},
			unassignWant: []string{"ABC", "EFG", "HIJ"},
		},
		{
			name:         "Parse Assign & Unassign",
			idList:       []string{"ABC", "EFG", "-HIJ"},
			assignWant:   []string{"ABC", "EFG"},
			unassignWant: []string{"HIJ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as, ua := parsePlayerAsgList(&Team{PlayerIDs: tt.idList})
			assert.StringSliceEqual(t, as, tt.assignWant)
			assert.StringSliceEqual(t, ua, tt.unassignWant)
		})
	}
}
