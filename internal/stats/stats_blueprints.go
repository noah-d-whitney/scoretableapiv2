package stats

type GameStatlineBlueprint []GameStat

var (
	GameStatBlueprintSimple      GameStatlineBlueprint = []GameStat{GamePointsSimple}
	GameStatlineBpPointsAdvanced GameStatlineBlueprint = []GameStat{GamePoints}
)

// PRIMITIVE STATS
const (
	PointSimple      PrimitiveStat = "PtsSimple"
	ThreePointMiss   PrimitiveStat = "3PtA"
	ThreePointMade   PrimitiveStat = "3PtM"
	TwoPointMiss     PrimitiveStat = "2PtA"
	TwoPointMade     PrimitiveStat = "2PtM"
	FreeThrowMiss    PrimitiveStat = "FTA"
	FreeThrowMade    PrimitiveStat = "FTM"
	Assist           PrimitiveStat = "Ast"
	Block            PrimitiveStat = "Blk"
	Steal            PrimitiveStat = "Stl"
	OffensiveRebound PrimitiveStat = "OReb"
	DefensiveRebound PrimitiveStat = "DReb"
	ReboundSimple    PrimitiveStat = "Reb"
	Turnover         PrimitiveStat = "To"
	FoulSimple       PrimitiveStat = "Fl"
)

// TODO return error for duplicate stat names

// PLAYER STATS
var (
	playerPoint = PlayerStat{
		name: "Pts",
		getFunc: func(primStats *PrimitiveStatline) any {
			var points int
			points += primStats.get(FreeThrowMade)
			points += primStats.get(TwoPointMade) * 2
			points += primStats.get(ThreePointMade) * 3
			return points
		},
		req: []PrimitiveStat{FreeThrowMade, TwoPointMade, ThreePointMade},
	}
	playerPointSimple = PlayerStat{
		name: "Pts",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(PointSimple)
		},
		req: []PrimitiveStat{PointSimple},
	}
	playerTwoPointAttempted = PlayerStat{
		name: "2PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(TwoPointMiss) + primStats.get(TwoPointMade)
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade},
	}
	playerTwoPointMade = PlayerStat{
		name: "2PtM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(TwoPointMade)
		},
		req: []PrimitiveStat{TwoPointMade},
	}
	playerTwoPointPercent = PlayerStat{
		name: "2Pt%",
		getFunc: func(primStats *PrimitiveStatline) any {
			twoMade := float64(primStats.get(TwoPointMade))
			twoMiss := float64(primStats.get(TwoPointMiss))
			return float64ToPercent(twoMade/twoMiss + twoMade)
		},
		req: []PrimitiveStat{TwoPointMade, TwoPointMiss},
	}
	playerThreePointAttempted = PlayerStat{
		name: "3PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ThreePointMiss) + primStats.get(ThreePointMade)
		},
		req: []PrimitiveStat{ThreePointMiss, ThreePointMade},
	}
	playerThreePointMade = PlayerStat{
		name: "3PtM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ThreePointMade)
		},
		req: []PrimitiveStat{ThreePointMade},
	}
	playerThreePointPercent = PlayerStat{
		name: "3Pt%",
		getFunc: func(primStats *PrimitiveStatline) any {
			threeMade := float64(primStats.get(ThreePointMade))
			threeMiss := float64(primStats.get(ThreePointMiss))
			return float64ToPercent(threeMade/threeMiss + threeMade)
		},
		req: []PrimitiveStat{ThreePointMiss, ThreePointMade},
	}
	playerFreeThrowAttempted = PlayerStat{
		name: "FTA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(FreeThrowMiss) + primStats.get(FreeThrowMade)
		},
		req: []PrimitiveStat{FreeThrowMade, FreeThrowMiss},
	}
	playerFreeThrowMade = PlayerStat{
		name: "FTM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(FreeThrowMade)
		},
		req: []PrimitiveStat{FreeThrowMade},
	}
	playerFreeThrowPercent = PlayerStat{
		name: "FT%",
		getFunc: func(primStats *PrimitiveStatline) any {
			return float64ToPercent(
				float64(primStats.get(FreeThrowMade))/
					float64(primStats.get(FreeThrowMade)) +
					float64(primStats.get(FreeThrowMiss)),
			)
		},
		req: []PrimitiveStat{FreeThrowMade, FreeThrowMiss},
	}
	playerFieldGoalAttempted = PlayerStat{
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
	playerFieldGoalMade = PlayerStat{
		name: "FGM",
		getFunc: func(primStats *PrimitiveStatline) any {
			var fgm int
			fgm += primStats.stats[TwoPointMade]
			fgm += primStats.stats[ThreePointMade]
			return fgm
		},
		req: []PrimitiveStat{TwoPointMade, ThreePointMade},
	}
	playerFieldGoalPercent = PlayerStat{
		name: "FG%",
		getFunc: func(primStats *PrimitiveStatline) any {
			var fgMade float64
			fgMade += float64(primStats.get(TwoPointMade))
			fgMade += float64(primStats.get(ThreePointMade))

			var fgMiss float64
			fgMiss += float64(primStats.get(TwoPointMiss))
			fgMiss += float64(primStats.get(TwoPointMade))
			fgMiss += float64(primStats.get(ThreePointMiss))
			fgMiss += float64(primStats.get(ThreePointMade))

			return float64ToPercent(fgMade/fgMiss + fgMade)
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade, ThreePointMiss, ThreePointMade},
	}
	playerDefensiveRebound = PlayerStat{
		name: "DRebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(DefensiveRebound)
		},
		req: []PrimitiveStat{DefensiveRebound},
	}
	playerOffensiveRebound = PlayerStat{
		name: "ORebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(OffensiveRebound)
		},
		req: []PrimitiveStat{OffensiveRebound},
	}
	playerRebound = PlayerStat{
		name: "Rebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(OffensiveRebound) + primStats.get(DefensiveRebound)
		},
		req: []PrimitiveStat{OffensiveRebound, DefensiveRebound},
	}
	playerReboundSimple = PlayerStat{
		name: "Rebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ReboundSimple)
		},
		req: []PrimitiveStat{ReboundSimple},
	}
	playerAssist = PlayerStat{
		name: "Ast",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Assist)
		},
		req: []PrimitiveStat{Assist},
	}
	playerSteal = PlayerStat{
		name: "Stl",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Steal)
		},
		req: []PrimitiveStat{Steal},
	}
	playerBlock = PlayerStat{
		name: "Blk",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Block)
		},
		req: []PrimitiveStat{Block},
	}
	playerTurnover = PlayerStat{
		name: "To",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Turnover)
		},
		req: []PrimitiveStat{Turnover},
	}
	playerFoulSimple = PlayerStat{
		name: "Fls",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(FoulSimple)
		},
		req: []PrimitiveStat{FoulSimple},
	}
)

// TEAM STATS
var (
	teamPoints = TeamStat{
		name: "Pts",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var points int
			for _, sl := range teamPlayersStats {
				points += sl.get(playerPoint).(int)
			}
			return points
		},
		req: []PlayerStat{playerPoint},
	}
	teamPointsSimple = TeamStat{
		name: "Pts",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var points int
			for _, sl := range teamPlayersStats {
				points += sl.get(playerPointSimple).(int)
			}
			return points
		},
		req: []PlayerStat{playerPointSimple},
	}
	teamTwoPointAttempted = TeamStat{
		name: "2PtA",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var twoAttempt int
			for _, sl := range teamPlayersStats {
				twoAttempt += sl.get(playerTwoPointAttempted).(int)
			}
			return twoAttempt
		},
		req: []PlayerStat{playerTwoPointAttempted},
	}
	teamTwoPointMade = TeamStat{
		name: "2PtM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerTwoPointMade).(int)
			}
			return count
		},
		req: []PlayerStat{playerTwoPointMade},
	}
	teamTwoPointPercent = TeamStat{
		name: "2Pt%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerTwoPointAttempted).(int)
				made += sl.get(playerTwoPointMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerTwoPointMade, playerTwoPointAttempted, playerTwoPointPercent},
	}
	teamThreePointAttempted = TeamStat{
		name: "3PtA",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerThreePointAttempted).(int)
			}
			return count
		},
		req: []PlayerStat{playerThreePointAttempted},
	}
	teamThreePointMade = TeamStat{
		name: "3PtM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerThreePointMade).(int)
			}
			return count
		},
		req: []PlayerStat{playerThreePointMade},
	}
	teamThreePointPercent = TeamStat{
		name: "3Pt%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerThreePointAttempted).(int)
				made += sl.get(playerThreePointMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerThreePointMade, playerThreePointAttempted, playerThreePointPercent},
	}
	teamFreeThrowAttempted = TeamStat{
		name: "FTA",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerFreeThrowAttempted).(int)
			}
			return count
		},
		req: []PlayerStat{playerFreeThrowAttempted},
	}
	teamFreeThrowMade = TeamStat{
		name: "FTM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerFreeThrowMade).(int)
			}
			return count
		},
		req: []PlayerStat{playerFreeThrowMade},
	}
	teamFreeThrowPercent = TeamStat{
		name: "FT%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerFreeThrowAttempted).(int)
				made += sl.get(playerFreeThrowMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerFreeThrowMade, playerFreeThrowAttempted, playerFreeThrowPercent},
	}
	teamFieldGoalAttempted = TeamStat{
		name: "FGA",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var fga int
			for _, sl := range teamPlayersStats {
				fga += sl.get(playerFieldGoalAttempted).(int)
			}
			return fga
		},
		req: []PlayerStat{playerFieldGoalAttempted},
	}
	teamFieldGoalMade = TeamStat{
		name: "FGM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var fgm int
			for _, sl := range teamPlayersStats {
				fgm += sl.get(playerFieldGoalMade).(int)
			}
			return fgm
		},
		req: []PlayerStat{playerFieldGoalMade},
	}
	teamFieldGoalPercent = TeamStat{
		name: "FG%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerFieldGoalAttempted).(int)
				made += sl.get(playerFieldGoalMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerFieldGoalMade, playerFieldGoalAttempted, playerFieldGoalPercent},
	}
	teamReboundSimple = TeamStat{
		name: "Rebs",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var rebs int
			for _, sl := range teamPlayersStats {
				rebs += sl.get(playerReboundSimple).(int)
			}
			return rebs
		},
		req: []PlayerStat{playerReboundSimple},
	}
	teamRebound = TeamStat{
		name: "Rebs",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var rebs int
			for _, sl := range teamPlayersStats {
				rebs += sl.get(playerRebound).(int)
			}
			return rebs
		},
		req: []PlayerStat{playerRebound},
	}
	teamOffensiveRebound = TeamStat{
		name: "ORebs",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var oRebs int
			for _, sl := range teamPlayersStats {
				oRebs += sl.get(playerOffensiveRebound).(int)
			}
			return oRebs
		},
		req: []PlayerStat{playerOffensiveRebound},
	}
	teamDefensiveRebound = TeamStat{
		name: "DRebs",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var dRebs int
			for _, sl := range teamPlayersStats {
				dRebs += sl.get(playerDefensiveRebound).(int)
			}
			return dRebs
		},
		req: []PlayerStat{playerDefensiveRebound},
	}
	teamAssist = TeamStat{
		name: "Ast",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var ast int
			for _, sl := range teamPlayersStats {
				ast += sl.get(playerAssist).(int)
			}
			return ast
		},
		req: []PlayerStat{playerAssist},
	}
	teamSteal = TeamStat{
		name: "Stl",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var stl int
			for _, sl := range teamPlayersStats {
				stl += sl.get(playerSteal).(int)
			}
			return stl
		},
		req: []PlayerStat{playerSteal},
	}
	teamBlock = TeamStat{
		name: "Blk",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var blk int
			for _, sl := range teamPlayersStats {
				blk += sl.get(playerBlock).(int)
			}
			return blk
		},
		req: []PlayerStat{playerBlock},
	}
	teamTurnover = TeamStat{
		name: "To",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var to int
			for _, sl := range teamPlayersStats {
				to += sl.get(playerTurnover).(int)
			}
			return to
		},
		req: []PlayerStat{playerTurnover},
	}
	teamFoulSimple = TeamStat{
		name: "Fls",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var fls int
			for _, sl := range teamPlayersStats {
				fls += sl.get(playerFoulSimple).(int)
			}
			return fls
		},
		req: []PlayerStat{playerFoulSimple},
	}
)

// TODO add rest of stats to team.
// TODO rename player stats to singular

// GAME STATS
var (
	GamePoints = GameStat{
		name: "Pts",
		getFunc: func(gameTeamsStats GameTeamsStatline) any {
			var points int
			points += gameTeamsStats.home.get(teamPoints).(int)
			points += gameTeamsStats.away.get(teamPoints).(int)
			return points
		},
		req: []TeamStat{teamPoints},
	}
	GamePointsSimple = GameStat{
		name: "Pts",
		getFunc: func(gameTeamsStats GameTeamsStatline) any {
			var points int
			points += gameTeamsStats.home.get(teamPoints).(int)
			points += gameTeamsStats.away.get(teamPointsSimple).(int)
			return points
		},
		req: []TeamStat{teamPointsSimple},
	}
	GameFieldGoalsAttempted = GameStat{
		name: "FGA",
		getFunc: func(gameTeamsStats GameTeamsStatline) any {
			var fga int
			fga += gameTeamsStats.home.get(teamFieldGoalAttempted).(int)
			return fga
		},
		req: []TeamStat{teamFieldGoalAttempted},
	}
	GameFieldGoalsMade = GameStat{
		name: "FGM",
		getFunc: func(gameTeamsStats GameTeamsStatline) any {
			var fgm int
			fgm += gameTeamsStats.home.get(teamFieldGoalMade).(int)
			return fgm
		},
		req: []TeamStat{teamFieldGoalMade},
	}
)
