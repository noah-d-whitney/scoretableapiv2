package gamehub

import (
	"errors"
	"fmt"
)

type keysWithTypes struct {
	key     string
	keyType int
	dest    any
}

var ErrNoValueForKey = errors.New("no value found for key")
var ErrValueNotAsserted = errors.New("value could not be asserted to specified type")

func checkAndAssertStringFromMap(src map[string]any, key string) (string, error) {
	data, ok := src[key]
	if !ok {
		return "", ErrNoValueForKey
	}
	value, ok := data.(string)
	if !ok {
		return "", ErrValueNotAsserted
	}

	return value, nil
}

func checkAndAssertIntFromMap(src map[string]any, key string) (int, error) {
	data, ok := src[key]
	if !ok {
		return 0, ErrNoValueForKey
	}
	fmt.Printf("data value: %v\n", data)

	value, ok := data.(float64)
	if !ok {
		return 0, ErrValueNotAsserted
	}

	return int(value), nil

}

func checkAndAssertBoolFromMap(src map[string]any, key string) (bool, error) {
	data, ok := src[key]
	if !ok {
		return false, ErrNoValueForKey
	}
	fmt.Printf("data value: %v\n", data)

	value, ok := data.(bool)
	if !ok {
		return false, ErrValueNotAsserted
	}

	return value, nil

}
