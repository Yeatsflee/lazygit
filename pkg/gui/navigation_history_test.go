package gui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newStringNavigationHistory() *navigationHistory[string] {
	return newNavigationHistory(func(a, b string) bool { return a == b })
}

func assertNavigationPosition[T any](t *testing.T, want T, move func(T) (T, bool), current T) {
	t.Helper()
	position, ok := move(current)
	assert.True(t, ok)
	assert.Equal(t, want, position)
}

func assertNoNavigationPosition[T any](t *testing.T, move func(T) (T, bool), current T) {
	t.Helper()
	_, ok := move(current)
	assert.False(t, ok)
}

func TestNavigationHistoryBackAndForward(t *testing.T) {
	history := newStringNavigationHistory()
	history.Record("file A")

	assertNavigationPosition(t, "file A", history.Back, "file B")
	assertNavigationPosition(t, "file B", history.Forward, "file A")
	assertNavigationPosition(t, "file A", history.Back, "file B")
	assertNavigationPosition(t, "file B", history.Forward, "file A")
}

func TestNavigationHistoryMultipleJumps(t *testing.T) {
	history := newStringNavigationHistory()
	history.Record("middle")
	history.Record("bottom")

	assertNavigationPosition(t, "bottom", history.Back, "top")
	assertNavigationPosition(t, "middle", history.Back, "bottom")
	assertNavigationPosition(t, "bottom", history.Forward, "middle")
	assertNavigationPosition(t, "top", history.Forward, "bottom")
}

func TestNavigationHistoryBoundaries(t *testing.T) {
	history := newStringNavigationHistory()
	assertNoNavigationPosition(t, history.Back, "file A")
	assertNoNavigationPosition(t, history.Forward, "file A")

	history.Record("file A")
	assertNoNavigationPosition(t, history.Forward, "file B")
	assertNavigationPosition(t, "file A", history.Back, "file B")
	assertNoNavigationPosition(t, history.Back, "file A")
	assertNoNavigationPosition(t, history.Back, "file A")
	assertNavigationPosition(t, "file B", history.Forward, "file A")
	assertNoNavigationPosition(t, history.Forward, "file B")
	assertNoNavigationPosition(t, history.Forward, "file B")
	assertNavigationPosition(t, "file A", history.Back, "file B")
}

func TestNavigationHistoryIgnoresRepeatedOriginsAndUnchangedDestination(t *testing.T) {
	history := newStringNavigationHistory()
	history.Record("file A")
	history.Record("file A")
	assertNoNavigationPosition(t, history.Back, "file A")

	history.Record("file A")
	history.Record("file B")
	history.Record("file B")
	assertNavigationPosition(t, "file A", history.Back, "file B")
	assertNavigationPosition(t, "file B", history.Forward, "file A")
	assertNoNavigationPosition(t, history.Forward, "file B")
}

func TestNavigationHistorySavesMovementBeforeLeaving(t *testing.T) {
	history := newStringNavigationHistory()
	history.Record("file A")
	history.Record("file B")
	assertNavigationPosition(t, "file B", history.Back, "file C")
	assertNavigationPosition(t, "file A", history.Back, "file B after j/k")
	assertNavigationPosition(t, "file B after j/k", history.Forward, "file A after j/k")
	assertNavigationPosition(t, "file C", history.Forward, "file B after more j/k")
	assertNavigationPosition(t, "file B after more j/k", history.Back, "file C")
	assertNavigationPosition(t, "file A after j/k", history.Back, "file B after more j/k")
}

func TestNavigationHistoryNewJumpDiscardsForwardHistory(t *testing.T) {
	history := newStringNavigationHistory()
	history.Record("file A")
	history.Record("file B")
	assertNavigationPosition(t, "file B", history.Back, "file C")

	history.Record("file B after j/k")
	assertNoNavigationPosition(t, history.Forward, "file D")
	assertNavigationPosition(t, "file B after j/k", history.Back, "file D")
	assertNavigationPosition(t, "file A", history.Back, "file B after j/k")
	assertNavigationPosition(t, "file B after j/k", history.Forward, "file A")
	assertNavigationPosition(t, "file D", history.Forward, "file B after j/k")
	assertNoNavigationPosition(t, history.Forward, "file D")
}

func TestNavigationHistoryLimitsRetainedLocations(t *testing.T) {
	history := newNavigationHistory(func(a, b int) bool { return a == b })
	for i := range 110 {
		history.Record(i)
	}
	for i := 110; i > 11; i-- {
		assertNavigationPosition(t, i-1, history.Back, i)
	}
	assertNoNavigationPosition(t, history.Back, 11)
	for i := 11; i < 110; i++ {
		assertNavigationPosition(t, i+1, history.Forward, i)
	}
	assertNoNavigationPosition(t, history.Forward, 110)
}
