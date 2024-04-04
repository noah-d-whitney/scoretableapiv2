package stats

// Blueprint is a slice of GameStat's used to initialize a GameStatline
type Blueprint []GameStat

var (
	Standard Blueprint = []GameStat{PointsCompound, FieldGoalsAttempted,
		FieldGoalsMade, FieldGoalPercent, FreeThrowsAttempted, FreeThrowsMade, FreeThrowPercent,
		TwosAttempted, TwosMade, TwoPointPercent, ThreesAttempted, ThreesMade, ThreePointPercent,
		ReboundsCompound, DefensiveRebounds, OffensiveRebounds, Steals, Blocks, Assists, Turnovers,
		FoulsSimple}
	Simple Blueprint = []GameStat{PointsSimple, ReboundsSimple, Steals, Blocks, Assists,
		Turnovers, FoulsSimple}
	NoMisses Blueprint = []GameStat{PointsCompound, FieldGoalsMade, FreeThrowsMade,
		TwosMade, ThreesMade, ReboundsSimple, Steals, Blocks, Assists, FoulsSimple}
)

// PRIMITIVE STATS
const (
	Point            PrimitiveStat = "Pts"
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
	Rebound          PrimitiveStat = "Reb"
	Turnover         PrimitiveStat = "To"
	Foul             PrimitiveStat = "Fl"
)

// PLAYER STATS
var (
	playerPointCompound = playerStat{
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
	playerPointSimple = playerStat{
		name: "Pts",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Point)
		},
		req: []PrimitiveStat{Point},
	}
	playerTwoPointAttempt = playerStat{
		name: "2PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(TwoPointMiss) + primStats.get(TwoPointMade)
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade},
	}
	playerTwoPointMade = playerStat{
		name: "2PtM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(TwoPointMade)
		},
		req: []PrimitiveStat{TwoPointMade},
	}
	playerTwoPointPercent = playerStat{
		name: "2Pt%",
		getFunc: func(primStats *PrimitiveStatline) any {
			twoMade := primStats.get(TwoPointMade)
			twoMiss := primStats.get(TwoPointMiss)
			return float64ToPercent(float64(twoMade) / float64(twoMiss+twoMade))
		},
		req: []PrimitiveStat{TwoPointMade, TwoPointMiss},
	}
	playerThreePointAttempt = playerStat{
		name: "3PtA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ThreePointMiss) + primStats.get(ThreePointMade)
		},
		req: []PrimitiveStat{ThreePointMiss, ThreePointMade},
	}
	playerThreePointMade = playerStat{
		name: "3PtM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(ThreePointMade)
		},
		req: []PrimitiveStat{ThreePointMade},
	}
	playerThreePointPercent = playerStat{
		name: "3Pt%",
		getFunc: func(primStats *PrimitiveStatline) any {
			return float64ToPercent(float64(primStats.get(ThreePointMade))/float64(primStats.get(
				ThreePointMade)) + float64(primStats.get(ThreePointMiss)))
		},
		req: []PrimitiveStat{ThreePointMiss, ThreePointMade},
	}
	playerFreeThrowAttempt = playerStat{
		name: "FTA",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(FreeThrowMiss) + primStats.get(FreeThrowMade)
		},
		req: []PrimitiveStat{FreeThrowMade, FreeThrowMiss},
	}
	playerFreeThrowMade = playerStat{
		name: "FTM",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(FreeThrowMade)
		},
		req: []PrimitiveStat{FreeThrowMade},
	}
	playerFreeThrowPercent = playerStat{
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
	playerFieldGoalAttempt = playerStat{
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
	playerFieldGoalMade = playerStat{
		name: "FGM",
		getFunc: func(primStats *PrimitiveStatline) any {
			var fgm int
			fgm += primStats.stats[TwoPointMade]
			fgm += primStats.stats[ThreePointMade]
			return fgm
		},
		req: []PrimitiveStat{TwoPointMade, ThreePointMade},
	}
	playerFieldGoalPercent = playerStat{
		name: "FG%",
		getFunc: func(primStats *PrimitiveStatline) any {
			var fgMade int
			fgMade += primStats.get(TwoPointMade)
			fgMade += primStats.get(ThreePointMade)

			var fgMiss int
			fgMiss += primStats.get(TwoPointMiss)
			fgMiss += primStats.get(ThreePointMiss)

			return float64ToPercent(float64(fgMade) / float64(fgMiss+fgMade))
		},
		req: []PrimitiveStat{TwoPointMiss, TwoPointMade, ThreePointMiss, ThreePointMade},
	}
	playerDefensiveRebound = playerStat{
		name: "DRebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(DefensiveRebound)
		},
		req: []PrimitiveStat{DefensiveRebound},
	}
	playerOffensiveRebound = playerStat{
		name: "ORebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(OffensiveRebound)
		},
		req: []PrimitiveStat{OffensiveRebound},
	}
	playerReboundCompound = playerStat{
		name: "Rebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(OffensiveRebound) + primStats.get(DefensiveRebound)
		},
		req: []PrimitiveStat{OffensiveRebound, DefensiveRebound},
	}
	playerReboundSimple = playerStat{
		name: "Rebs",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Rebound)
		},
		req: []PrimitiveStat{Rebound},
	}
	playerAssist = playerStat{
		name: "Ast",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Assist)
		},
		req: []PrimitiveStat{Assist},
	}
	playerSteal = playerStat{
		name: "Stl",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Steal)
		},
		req: []PrimitiveStat{Steal},
	}
	playerBlock = playerStat{
		name: "Blk",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Block)
		},
		req: []PrimitiveStat{Block},
	}
	playerTurnover = playerStat{
		name: "To",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Turnover)
		},
		req: []PrimitiveStat{Turnover},
	}
	playerFoulSimple = playerStat{
		name: "Fls",
		getFunc: func(primStats *PrimitiveStatline) any {
			return primStats.get(Foul)
		},
		req: []PrimitiveStat{Foul},
	}
)

// TEAM STATS
var (
	teamPointCompound = teamStat{
		name: "Pts",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var points int
			for _, sl := range teamPlayersStats {
				points += sl.get(playerPointCompound).(int)
			}
			return points
		},
		req: []playerStat{playerPointCompound},
	}
	teamPointSimple = teamStat{
		name: "Pts",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var points int
			for _, sl := range teamPlayersStats {
				points += sl.get(playerPointSimple).(int)
			}
			return points
		},
		req: []playerStat{playerPointSimple},
	}
	teamTwoPointAttempt = teamStat{
		name: "2PtA",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var twoAttempt int
			for _, sl := range teamPlayersStats {
				twoAttempt += sl.get(playerTwoPointAttempt).(int)
			}
			return twoAttempt
		},
		req: []playerStat{playerTwoPointAttempt},
	}
	teamTwoPointMake = teamStat{
		name: "2PtM",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerTwoPointMade).(int)
			}
			return count
		},
		req: []playerStat{playerTwoPointMade},
	}
	teamTwoPointPercent = teamStat{
		name: "2Pt%",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerTwoPointAttempt).(int)
				made += sl.get(playerTwoPointMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []playerStat{playerTwoPointMade, playerTwoPointAttempt, playerTwoPointPercent},
	}
	teamThreePointAttempt = teamStat{
		name: "3PtA",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerThreePointAttempt).(int)
			}
			return count
		},
		req: []playerStat{playerThreePointAttempt},
	}
	teamThreePointMake = teamStat{
		name: "3PtM",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerThreePointMade).(int)
			}
			return count
		},
		req: []playerStat{playerThreePointMade},
	}
	teamThreePointPercent = teamStat{
		name: "3Pt%",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerThreePointAttempt).(int)
				made += sl.get(playerThreePointMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []playerStat{playerThreePointMade, playerThreePointAttempt, playerThreePointPercent},
	}
	teamFreeThrowAttempt = teamStat{
		name: "FTA",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerFreeThrowAttempt).(int)
			}
			return count
		},
		req: []playerStat{playerFreeThrowAttempt},
	}
	teamFreeThrowMake = teamStat{
		name: "FTM",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var count int
			for _, sl := range teamPlayersStats {
				count += sl.get(playerFreeThrowMade).(int)
			}
			return count
		},
		req: []playerStat{playerFreeThrowMade},
	}
	teamFreeThrowPercent = teamStat{
		name: "FT%",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerFreeThrowAttempt).(int)
				made += sl.get(playerFreeThrowMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []playerStat{playerFreeThrowMade, playerFreeThrowAttempt, playerFreeThrowPercent},
	}
	teamFieldGoalAttempt = teamStat{
		name: "FGA",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var fga int
			for _, sl := range teamPlayersStats {
				fga += sl.get(playerFieldGoalAttempt).(int)
			}
			return fga
		},
		req: []playerStat{playerFieldGoalAttempt},
	}
	teamFieldGoalMake = teamStat{
		name: "FGM",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var fgm int
			for _, sl := range teamPlayersStats {
				fgm += sl.get(playerFieldGoalMade).(int)
			}
			return fgm
		},
		req: []playerStat{playerFieldGoalMade},
	}
	teamFieldGoalPercent = teamStat{
		name: "FG%",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var attp, made int
			for _, sl := range teamPlayersStats {
				attp += sl.get(playerFieldGoalAttempt).(int)
				made += sl.get(playerFieldGoalMade).(int)
			}
			return float64ToPercent(float64(made) / float64(attp))
		},
		req: []playerStat{playerFieldGoalMade, playerFieldGoalAttempt, playerFieldGoalPercent},
	}
	teamReboundSimple = teamStat{
		name: "Rebs",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var rebs int
			for _, sl := range teamPlayersStats {
				rebs += sl.get(playerReboundSimple).(int)
			}
			return rebs
		},
		req: []playerStat{playerReboundSimple},
	}
	teamReboundCompound = teamStat{
		name: "Rebs",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var rebs int
			for _, sl := range teamPlayersStats {
				rebs += sl.get(playerReboundCompound).(int)
			}
			return rebs
		},
		req: []playerStat{playerReboundCompound},
	}
	teamOffensiveRebound = teamStat{
		name: "ORebs",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var oRebs int
			for _, sl := range teamPlayersStats {
				oRebs += sl.get(playerOffensiveRebound).(int)
			}
			return oRebs
		},
		req: []playerStat{playerOffensiveRebound},
	}
	teamDefensiveRebound = teamStat{
		name: "DRebs",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var dRebs int
			for _, sl := range teamPlayersStats {
				dRebs += sl.get(playerDefensiveRebound).(int)
			}
			return dRebs
		},
		req: []playerStat{playerDefensiveRebound},
	}
	teamAssist = teamStat{
		name: "Ast",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var ast int
			for _, sl := range teamPlayersStats {
				ast += sl.get(playerAssist).(int)
			}
			return ast
		},
		req: []playerStat{playerAssist},
	}
	teamSteal = teamStat{
		name: "Stl",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var stl int
			for _, sl := range teamPlayersStats {
				stl += sl.get(playerSteal).(int)
			}
			return stl
		},
		req: []playerStat{playerSteal},
	}
	teamBlock = teamStat{
		name: "Blk",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var blk int
			for _, sl := range teamPlayersStats {
				blk += sl.get(playerBlock).(int)
			}
			return blk
		},
		req: []playerStat{playerBlock},
	}
	teamTurnover = teamStat{
		name: "To",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var to int
			for _, sl := range teamPlayersStats {
				to += sl.get(playerTurnover).(int)
			}
			return to
		},
		req: []playerStat{playerTurnover},
	}
	teamFoulSimple = teamStat{
		name: "Fls",
		getFunc: func(teamPlayersStats teamPlayersStatline) any {
			var fls int
			for _, sl := range teamPlayersStats {
				fls += sl.get(playerFoulSimple).(int)
			}
			return fls
		},
		req: []playerStat{playerFoulSimple},
	}
)

// GAME STATS
var (
	PointsSimple = GameStat{
		name: "Pts",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var points int
			points += gameTeamsStats.home.get(teamPointSimple).(int)
			points += gameTeamsStats.away.get(teamPointSimple).(int)
			return points
		},
		req: []teamStat{teamPointSimple},
	}
	PointsCompound = GameStat{
		name: "Pts",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var points int
			points += gameTeamsStats.home.get(teamPointCompound).(int)
			points += gameTeamsStats.away.get(teamPointCompound).(int)
			return points
		},
		req: []teamStat{teamPointCompound},
	}
	TwosAttempted = GameStat{
		name: "2PtA",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var attp int
			attp += gameTeamsStats.home.get(teamTwoPointAttempt).(int)
			attp += gameTeamsStats.away.get(teamTwoPointAttempt).(int)
			return attp
		},
		req: []teamStat{teamTwoPointAttempt},
	}
	TwosMade = GameStat{
		name: "2PtM",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var makes int
			makes += gameTeamsStats.home.get(teamTwoPointMake).(int)
			makes += gameTeamsStats.away.get(teamTwoPointMake).(int)
			return makes
		},
		req: []teamStat{teamTwoPointMake},
	}
	TwoPointPercent = GameStat{
		name: "2Pt%",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var makes int
			makes += gameTeamsStats.home.get(teamTwoPointMake).(int)
			makes += gameTeamsStats.away.get(teamTwoPointMake).(int)

			var attp int
			attp += gameTeamsStats.home.get(teamTwoPointAttempt).(int)
			attp += gameTeamsStats.away.get(teamTwoPointAttempt).(int)

			percent := float64ToPercent(float64(makes) / float64(attp))
			return percent
		},
		req: []teamStat{teamTwoPointAttempt, teamTwoPointMake, teamTwoPointPercent},
	}
	ThreesAttempted = GameStat{
		name: "3PtA",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var threeAt int
			threeAt += gameTeamsStats.home.get(teamThreePointAttempt).(int)
			threeAt += gameTeamsStats.away.get(teamThreePointAttempt).(int)
			return threeAt
		},
		req: []teamStat{teamThreePointAttempt},
	}
	ThreesMade = GameStat{
		name: "3PtM",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var fgm int
			fgm += gameTeamsStats.home.get(teamThreePointMake).(int)
			fgm += gameTeamsStats.away.get(teamThreePointMake).(int)
			return fgm
		},
		req: []teamStat{teamThreePointMake},
	}
	ThreePointPercent = GameStat{
		name: "3Pt%",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var makes int
			makes += gameTeamsStats.home.get(teamThreePointMake).(int)
			makes += gameTeamsStats.away.get(teamThreePointMake).(int)

			var attp int
			attp += gameTeamsStats.home.get(teamThreePointAttempt).(int)
			attp += gameTeamsStats.away.get(teamThreePointAttempt).(int)

			percent := float64ToPercent(float64(makes) / float64(attp))
			return percent
		},
		req: []teamStat{teamThreePointAttempt, teamThreePointMake, teamThreePointPercent},
	}
	FreeThrowsAttempted = GameStat{
		name: "FTA",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var fta int
			fta += gameTeamsStats.home.get(teamFreeThrowAttempt).(int)
			fta += gameTeamsStats.away.get(teamFreeThrowAttempt).(int)
			return fta
		},
		req: []teamStat{teamFreeThrowAttempt},
	}
	FreeThrowsMade = GameStat{
		name: "FTM",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var ftm int
			ftm += gameTeamsStats.home.get(teamFreeThrowMake).(int)
			ftm += gameTeamsStats.away.get(teamFreeThrowMake).(int)
			return ftm
		},
		req: []teamStat{teamFreeThrowMake},
	}
	FreeThrowPercent = GameStat{
		name: "FT%",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var makes int
			makes += gameTeamsStats.home.get(teamFreeThrowMake).(int)
			makes += gameTeamsStats.away.get(teamFreeThrowMake).(int)

			var attp int
			attp += gameTeamsStats.home.get(teamFreeThrowAttempt).(int)
			attp += gameTeamsStats.away.get(teamFreeThrowAttempt).(int)

			percent := float64ToPercent(float64(makes) / float64(attp))
			return percent
		},
		req: []teamStat{teamFreeThrowAttempt, teamFreeThrowMake, teamFreeThrowPercent},
	}
	FieldGoalsAttempted = GameStat{
		name: "FGA",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var fga int
			fga += gameTeamsStats.home.get(teamFieldGoalAttempt).(int)
			fga += gameTeamsStats.away.get(teamFieldGoalAttempt).(int)
			return fga
		},
		req: []teamStat{teamFieldGoalAttempt},
	}
	FieldGoalsMade = GameStat{
		name: "FGM",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var fgm int
			fgm += gameTeamsStats.home.get(teamFieldGoalMake).(int)
			fgm += gameTeamsStats.away.get(teamFieldGoalMake).(int)
			return fgm
		},
		req: []teamStat{teamFieldGoalMake},
	}
	FieldGoalPercent = GameStat{
		name: "FG%",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var fgm, fga int
			fgm += gameTeamsStats.home.get(teamFieldGoalMake).(int)
			fgm += gameTeamsStats.away.get(teamFieldGoalMake).(int)
			fga += gameTeamsStats.home.get(teamFieldGoalAttempt).(int)
			fga += gameTeamsStats.away.get(teamFieldGoalAttempt).(int)
			return float64ToPercent(float64(fgm) / float64(fga))
		},
		req: []teamStat{teamFieldGoalMake, teamFieldGoalAttempt, teamFieldGoalPercent},
	}
	ReboundsSimple = GameStat{
		name: "Rebs",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var rebs int
			rebs += gameTeamsStats.home.get(teamReboundSimple).(int)
			rebs += gameTeamsStats.away.get(teamReboundSimple).(int)
			return rebs
		},
		req: []teamStat{teamReboundSimple},
	}
	ReboundsCompound = GameStat{
		name: "Rebs",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var rebs int
			rebs += gameTeamsStats.home.get(teamReboundCompound).(int)
			rebs += gameTeamsStats.away.get(teamReboundCompound).(int)
			return rebs
		},
		req: []teamStat{teamReboundCompound},
	}
	OffensiveRebounds = GameStat{
		name: "ORebs",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var rebs int
			rebs += gameTeamsStats.home.get(teamOffensiveRebound).(int)
			rebs += gameTeamsStats.away.get(teamOffensiveRebound).(int)
			return rebs
		},
		req: []teamStat{teamOffensiveRebound},
	}
	DefensiveRebounds = GameStat{
		name: "DRebs",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var rebs int
			rebs += gameTeamsStats.home.get(teamDefensiveRebound).(int)
			rebs += gameTeamsStats.away.get(teamDefensiveRebound).(int)
			return rebs
		},
		req: []teamStat{teamDefensiveRebound},
	}
	Assists = GameStat{
		name: "Ast",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var ast int
			ast += gameTeamsStats.home.get(teamAssist).(int)
			ast += gameTeamsStats.away.get(teamAssist).(int)
			return ast
		},
		req: []teamStat{teamAssist},
	}
	Steals = GameStat{
		name: "Stl",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var stl int
			stl += gameTeamsStats.home.get(teamSteal).(int)
			stl += gameTeamsStats.away.get(teamSteal).(int)
			return stl
		},
		req: []teamStat{teamSteal},
	}
	Blocks = GameStat{
		name: "Blk",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var blk int
			blk += gameTeamsStats.home.get(teamBlock).(int)
			blk += gameTeamsStats.away.get(teamBlock).(int)
			return blk
		},
		req: []teamStat{teamBlock},
	}
	Turnovers = GameStat{
		name: "To",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var to int
			to += gameTeamsStats.home.get(teamTurnover).(int)
			to += gameTeamsStats.away.get(teamTurnover).(int)
			return to
		},
		req: []teamStat{teamTurnover},
	}
	FoulsSimple = GameStat{
		name: "Fls",
		getFunc: func(gameTeamsStats gameTeamsStatline) any {
			var fls int
			fls += gameTeamsStats.home.get(teamFoulSimple).(int)
			fls += gameTeamsStats.away.get(teamFoulSimple).(int)
			return fls
		},
		req: []teamStat{teamFoulSimple},
	}
)
