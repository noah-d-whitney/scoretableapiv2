package main

import (
	"ScoreTableApi/internal/data"
	"fmt"
	"net/http"
	"time"
)

func (app *application) InsertGame(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DateTime     *time.Time `json:"date_time"`
		TeamSize     int64      `json:"team_size"`
		Type         string     `json:"type"`
		PeriodLength int64      `json:"period_length"`
		PeriodCount  int64      `json:"period_count"`
		ScoreTarget  int64      `json:"score_target"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	game := &data.Game{
		DateTime:     *input.DateTime,
		TeamSize:     input.TeamSize,
		Type:         input.Type,
		PeriodLength: input.PeriodLength,
		PeriodCount:  input.PeriodCount,
		ScoreTarget:  input.ScoreTarget,
	}

	app.logger.PrintInfo(fmt.Sprintf("%+d\n", &game), nil)
}
