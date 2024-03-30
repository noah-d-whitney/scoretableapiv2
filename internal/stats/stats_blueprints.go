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
	playerPointsSimple = PlayerStat{
		name: "Pts",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(PointSimple)
		},
		req: []PrimitiveStat{PointSimple},
	}
	playerTwoPointsAttempted = PlayerStat{
		name: "2PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(TwoPointMiss) + primStats.get(TwoPointMade)
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade},
	}
	playerTwoPointsMade = PlayerStat{
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
	playerThreePointsAttempted = PlayerStat{
		name: "3PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ThreePointMiss) + primStats.get(ThreePointMade)
		},
		req: []PrimitiveStat{ThreePointMiss, ThreePointMade},
	}
	playerThreePointsMade = PlayerStat{
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
	playerFieldGoalsAttempted = PlayerStat{
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
	playerFieldGoalsMade = PlayerStat{
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
	playerDefensiveRebounds = PlayerStat{
		name: "DRebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(DefensiveRebound)
		},
		req: []PrimitiveStat{DefensiveRebound},
	}
	playerOffensiveRebounds = PlayerStat{
		name: "ORebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(OffensiveRebound)
		},
		req: []PrimitiveStat{OffensiveRebound},
	}
	playerRebounds = PlayerStat{
		name: "Rebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(OffensiveRebound) + primStats.get(DefensiveRebound)
		},
		req: []PrimitiveStat{OffensiveRebound, DefensiveRebound},
	}
	playerReboundsSimple = PlayerStat{
		name: "Rebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ReboundSimple)
		},
		req: []PrimitiveStat{ReboundSimple},
	}
	playerAssists = PlayerStat{
		name: "Ast",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Assist)
		},
		req: []PrimitiveStat{Assist},
	}
	playerSteals = PlayerStat{
		name: "Stl",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Steal)
		},
		req: []PrimitiveStat{Steal},
	}
	playerBlocks = PlayerStat{
		name: "Blk",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Block)
		},
		req: []PrimitiveStat{Block},
	}
	playerTurnovers = PlayerStat{
		name: "To",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Turnover)
		},
		req: []PrimitiveStat{Turnover},
	}
	playerFoulsSimple = PlayerStat{
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
				points += sl.get(playerPointsSimple).(int)
			}
			return points
		},
		req: []PlayerStat{playerPointsSimple},
	}
	teamTwoPointAttempted = TeamStat{
		name: "2PtA",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var twoAttempt int
			for _, sl := range teamPlayersStats {
				twoAttempt += sl.get(playerTwoPointsAttempted).(int)
			}
			return twoAttempt
		},
		req: []PlayerStat{playerTwoPointsAttempted},
	}
	teamTwoPointMade = TeamStat{
		name: "2PtM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerTwoPointsMade).(int)
			}
			return count
		},
		req: []PlayerStat{playerTwoPointsMade},
	}
	teamTwoPointPercent = TeamStat{
		name: "2Pt%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerTwoPointsAttempted).(int)
				made += sl.get(playerTwoPointsMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerTwoPointsMade, playerTwoPointsAttempted, playerTwoPointPercent},
	}
	teamThreePointAttempted = TeamStat{
		name: "3PtA",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerThreePointsAttempted).(int)
			}
			return count
		},
		req: []PlayerStat{playerThreePointsAttempted},
	}
	teamThreePointMade = TeamStat{
		name: "3PtM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerThreePointsMade).(int)
			}
			return count
		},
		req: []PlayerStat{playerThreePointsMade},
	}
	teamThreePointPercent = TeamStat{
		name: "3Pt%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerThreePointsAttempted).(int)
				made += sl.get(playerThreePointsMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerThreePointsMade, playerThreePointsAttempted, playerThreePointPercent},
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
				fga += sl.get(playerFieldGoalsAttempted).(int)
			}
			return fga
		},
		req: []PlayerStat{playerFieldGoalsAttempted},
	}
	teamFieldGoalMade = TeamStat{
		name: "FGM",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var fgm int
			for _, sl := range teamPlayersStats {
				fgm += sl.get(playerFieldGoalsMade).(int)
			}
			return fgm
		},
		req: []PlayerStat{playerFieldGoalsMade},
	}
	teamFieldGoalPercent = TeamStat{
		name: "FG%",
		getFunc: func(teamPlayersStats TeamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerFieldGoalsAttempted).(int)
				made += sl.get(playerFieldGoalsMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []PlayerStat{playerFieldGoalsMade, playerFieldGoalsAttempted, playerFieldGoalPercent},
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
