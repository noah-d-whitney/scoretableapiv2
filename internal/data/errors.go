package data

import json2 "encoding/json"

type ModelValidationErr struct {
	Errors map[string]string
}

func (e ModelValidationErr) Error() string {
	json, _ := json2.MarshalIndent(e.Errors, "", "\t")
	return string(json)
}

func NewModelValidationErr(key string, value string) ModelValidationErr {
	return ModelValidationErr{
		Errors: map[string]string{key: value},
	}
}

func (e ModelValidationErr) AddError(key string, value string) {
	if _, exists := e.Errors[key]; !exists {
		e.Errors[key] = value
	}
}

func (e ModelValidationErr) Valid() bool {
	return len(e.Errors) == 0
}
