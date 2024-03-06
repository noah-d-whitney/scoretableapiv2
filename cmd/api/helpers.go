package main

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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
