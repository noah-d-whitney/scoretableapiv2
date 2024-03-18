package assert

import (
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

func StringSliceEqual(t *testing.T, actual, expected []string) {
	t.Helper()

	if slices.Compare(actual, expected) != 0 {
		t.Errorf("got [%s], expected: [%s]", strings.Join(actual, ", "), strings.Join(expected, ","+
			" "))
	}
}
