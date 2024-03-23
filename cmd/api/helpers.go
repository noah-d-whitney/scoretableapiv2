package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	json2 "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope,
	headers http.Header) error {
	json, err := json2.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	json = append(json, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(json)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json2.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dest)
	if err != nil {
		var syntaxError *json2.SyntaxError
		var unmarshalTypeError *json2.UnmarshalTypeError
		var invalidUnmarshalError *json2.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q",
					unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)",
				unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readCSGameStatus(qs url.Values, defaultValue []data.GameStatus,
	v *validator.Validator) []data.GameStatus {
	key := "status"
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	statuses := make([]data.GameStatus, 0)
	split := strings.Split(csv, ",")
	for _, s := range split {
		var status data.GameStatus
		switch s {
		case "not-started":
			status = data.NOTSTARTED
		case "in-progress":
			status = data.INPROGRESS
		case "finished":
			status = data.FINISHED
		case "canceled":
			status = data.CANCELED
		default:
			v.AddError(key,
				`must be selected from the following: "not-started","in-progress","finished","canceled"`)
			return defaultValue
		}
		statuses = append(statuses, status)
	}

	return statuses
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) readDate(qs url.Values, key string, defaultValue time.Time,
	v *validator.Validator) time.Time {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		v.AddError(key, "must be a valid date (YYYY-MM-DD)")
		return defaultValue
	}

	return t
}

func (app *application) readCSInt(qs url.Values, key string, defaultValue []int64,
	v *validator.Validator) []int64 {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	ints := make([]int64, 0)
	split := strings.Split(csv, ",")
	for _, s := range split {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			v.AddError(key, fmt.Sprintf(`"%s" is not a valid integer`, s))
			return defaultValue
		}
		ints = append(ints, i)
	}

	return ints
}

func (app *application) backgroundTask(task func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		task()
	}()
}
