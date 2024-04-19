package gamehub

import (
	"ScoreTableApi/internal/data"
	"slices"
	strings2 "strings"
)

type lineup []*data.Player
type lineupManager struct {
	home     lineup
	away     lineup
	homeDnp  lineup
	awayDnp  lineup
	teamSize int
}

func (lm *lineupManager) getActive() map[data.GameTeamSide]lineup {
	homeActive := make([]*data.Player, lm.teamSize)
	for i := 0; i < lm.teamSize; i++ {
		homeActive[i] = lm.home[i]
	}
	awayActive := make([]*data.Player, lm.teamSize)
	for i := 0; i < lm.teamSize; i++ {
		awayActive[i] = lm.away[i]
	}

	return map[data.GameTeamSide]lineup{
		data.TeamHome: homeActive,
		data.TeamAway: awayActive,
	}
}

func (lm *lineupManager) getBench() map[data.GameTeamSide]lineup {
	homeBench := make([]*data.Player, 0)
	for i := lm.teamSize; i < len(lm.home); i++ {
		homeBench = append(homeBench, lm.home[i])
	}
	awayBench := make([]*data.Player, 0)
	for i := lm.teamSize; i < len(lm.away); i++ {
		awayBench = append(awayBench, lm.away[i])
	}

	return map[data.GameTeamSide]lineup{
		data.TeamHome: homeBench,
		data.TeamAway: awayBench,
	}
}

func (lm *lineupManager) getDnp() map[data.GameTeamSide]lineup {
	return map[data.GameTeamSide]lineup{
		data.TeamHome: lm.homeDnp,
		data.TeamAway: lm.awayDnp,
	}
}

func (lm *lineupManager) isActive(playerPin string) bool {
	active := lm.getActive()
	return slices.ContainsFunc(active[data.TeamHome], func(p *data.Player) bool {
		return p.PinId.Pin == playerPin
	}) || slices.ContainsFunc(active[data.TeamAway], func(p *data.Player) bool {
		return p.PinId.Pin == playerPin
	})
}

func (lm *lineupManager) substitution(side data.GameTeamSide, outPin string, inPin string) {
	var lnp *lineup
	switch side {
	case data.TeamHome:
		lnp = &lm.home
	case data.TeamAway:
		lnp = &lm.away
	}

	var outPlayer, inPlayer *data.Player
	var outIdx, inIdx int
	for i := 0; i < lm.teamSize; i++ {
		if strings2.Compare((*lnp)[i].PinId.Pin, outPin) == 0 {
			outPlayer = (*lnp)[i]
			outIdx = i
		}
	}
	if outPlayer == nil {
		return
	}

	for i := lm.teamSize; i < len(*lnp); i++ {
		if strings2.Compare((*lnp)[i].PinId.Pin, inPin) == 0 {
			inPlayer = (*lnp)[i]
			inIdx = i
		}
	}
	if inPlayer == nil {
		return
	}
	*inPlayer.LineupPos = outIdx + 1
	*outPlayer.LineupPos = inIdx + 1
	(*lnp)[outIdx] = inPlayer
	(*lnp)[inIdx] = outPlayer

	return
}

func newLineupManager(g *data.Game) *lineupManager {
	homeLnp := make([]*data.Player, 0)
	homeDnp := make([]*data.Player, 0)
	for _, p := range g.Teams.Home.Players {
		if p.LineupPos == nil {
			homeDnp = append(homeDnp, p)
		} else {
			homeLnp = append(homeLnp, p)
		}
	}
	awayLnp := make([]*data.Player, 0)
	awayDnp := make([]*data.Player, 0)
	for _, p := range g.Teams.Away.Players {
		if p.LineupPos == nil {
			awayDnp = append(awayDnp, p)
		} else {
			awayLnp = append(awayLnp, p)
		}
	}

	return &lineupManager{
		home:     homeLnp,
		away:     awayLnp,
		homeDnp:  homeDnp,
		awayDnp:  awayDnp,
		teamSize: int(g.TeamSize),
	}
}
