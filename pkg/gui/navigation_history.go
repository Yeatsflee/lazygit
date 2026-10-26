package gui

const navigationHistoryLimit = 100

// navigationHistory stores the origins of explicit jumps. Ordinary cursor
// movement is saved only when leaving a location, rather than becoming another
// entry in the history.
type navigationHistory[T any] struct {
	entries []T
	// A cursor past the last entry means that the current jump's destination
	// has not been saved yet.
	cursor int
	equal  func(T, T) bool
}

func newNavigationHistory[T any](equal func(T, T) bool) *navigationHistory[T] {
	return &navigationHistory[T]{equal: equal}
}

// Record saves the position before a jump, discarding any forward history.
func (history *navigationHistory[T]) Record(current T) {
	if history.cursor < len(history.entries) {
		history.entries = history.entries[:history.cursor]
	}
	if len(history.entries) == 0 || !history.equal(history.entries[len(history.entries)-1], current) {
		history.append(current)
	}
	history.cursor = len(history.entries)
}

func (history *navigationHistory[T]) Back(current T) (T, bool) {
	if len(history.entries) == 0 {
		var zero T
		return zero, false
	}

	if history.cursor == len(history.entries) {
		if !history.equal(history.entries[len(history.entries)-1], current) {
			history.append(current)
		}
		history.cursor = len(history.entries) - 1
	} else {
		history.entries[history.cursor] = current
	}

	if history.cursor == 0 {
		var zero T
		return zero, false
	}
	history.cursor--
	return history.entries[history.cursor], true
}

func (history *navigationHistory[T]) Forward(current T) (T, bool) {
	if history.cursor < len(history.entries) {
		history.entries[history.cursor] = current
	}
	if history.cursor+1 >= len(history.entries) {
		var zero T
		return zero, false
	}
	history.cursor++
	return history.entries[history.cursor], true
}

func (history *navigationHistory[T]) append(location T) {
	if len(history.entries) == navigationHistoryLimit {
		copy(history.entries, history.entries[1:])
		history.entries[len(history.entries)-1] = location
		return
	}
	history.entries = append(history.entries, location)
}
