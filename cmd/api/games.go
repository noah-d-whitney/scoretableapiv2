package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

func (app *application) InsertGame(w http.ResponseWriter, r *http.Request) {
	var input data.GameDto
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userID := app.contextGetUser(r).ID
	v := validator.New()

	game := input.Convert(v)
	game.UserID = userID

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Games.Insert(game)
	if err != nil {
		var modelValidationErr data.ModelValidationErr
		switch {
		case errors.As(err, &modelValidationErr):
			app.failedValidationResponse(w, r, modelValidationErr.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/game/%s", game.PinID.Pin))
	err = app.writeJSON(w, http.StatusCreated, envelope{"game": game}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetGame(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	game, err := app.models.Games.Get(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"game": game}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

func (app *application) GetAllGames(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Filters  data.GamesFilter
		Includes struct {
			Values   []string
			SafeList []string
		}
	}

	qs := r.URL.Query()
	v := validator.New()
	userID := app.contextGetUser(r).ID

	input.Includes.Values = app.readCSV(qs, "includes", make([]string, 0))
	input.Includes.SafeList = []string{"players"}
	for _, str := range input.Includes.Values {
		if !slices.Contains(input.Includes.SafeList, str) {

			v.AddError("includes", fmt.Sprintf(`Invalid includes value.
Possible include values for teams are: "%s"`, strings.Join(input.Includes.SafeList, `", "`)))
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
	}

	input.Filters.DateRange.AfterDate = app.readDate(qs, "after_date", nil, v)
	input.Filters.DateRange.BeforeDate = app.readDate(qs, "before_date", nil, v)
	if input.Filters.DateRange.BeforeDate != nil {
		timePlusDay := *input.Filters.DateRange.BeforeDate
		timePlusDay = timePlusDay.Add(3 * time.Hour)
		input.Filters.DateRange.BeforeDate = &timePlusDay
	}
	input.Filters.TeamPins = app.readCSV(qs, "team_pins", nil)
	input.Filters.PlayerPins = app.readCSV(qs, "player_pins", nil)
	input.Filters.Type = data.GameType(app.readString(qs, "type", ""))
	input.Filters.TeamSize = app.readCSInt(qs, "team_size", nil, v)
	input.Filters.Status = app.readCSGameStatus(qs, nil, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 5, v)
	input.Filters.Sort = app.readString(qs, "sort", "date_time")
	input.Filters.SortSafeList = []string{"date_time", "-date_time"}

	if data.ValidateGamesFilter(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	games, metadata, err := app.models.Games.GetAll(userID, input.Filters, input.Includes.Values)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "games": games}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

func (app *application) UpdateGame(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	game, err := app.models.Games.Get(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input data.GameDto
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	input.Merge(v, game)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Games.Update(game)
	if err != nil {
		var modelValidationErr data.ModelValidationErr
		switch {
		case errors.As(err, &modelValidationErr):
			app.failedValidationResponse(w, r, modelValidationErr.Errors)
		default:
			app.badRequestResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"game": game}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

func (app *application) DeleteGame(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	err := app.models.Games.Delete(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"message": fmt.Sprintf("game (%s) successfully deleted", pin)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (app *application) StartGame(w http.ResponseWriter, r *http.Request) {
	//userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	g, err := app.models.Games.Get(7, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	h, err := app.gameHubs.StartGame(g)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// Get game from db and send
	// create game stat object and hub
	// Accept message
	// Update game stat in hub
	// Send updates to all clients
	// Send event to DB concurrently
	err = h.JoinKeeper(7, w, r)
	if err != nil {
		h.Errors <- err
	}
}

func (app *application) WatchGame(w http.ResponseWriter, r *http.Request) {
	//userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	_, err := app.gameHubs.WatcherJoinGame(pin, w, r)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
