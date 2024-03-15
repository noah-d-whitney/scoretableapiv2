package assert

import (
	json2 "encoding/json"
	"slices"
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want %v", actual, expected)
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubstring)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}

func Int64SliceEqual(t *testing.T, actual, expected []int64) {
	t.Helper()
	acStr, _ := json2.Marshal(actual)
	expStr, _ := json2.Marshal(expected)

	if slices.Compare(actual, expected) != 0 {
		t.Errorf("got %s, expected: %s", acStr, expStr)
	}
}
