package gocui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeybindingMiddlewareCanConsumeAndDispatchKeys(t *testing.T) {
	g := &Gui{}
	key := NewKeyRune('g')
	dispatched := 0
	g.SetKeybinding("", key, func(*Gui, *View) error {
		dispatched++
		return nil
	})
	consume := true
	seen := []Key{}
	g.KeybindingMiddleware = func(key Key, dispatch func() error) error {
		seen = append(seen, key)
		if consume {
			return nil
		}
		return dispatch()
	}

	assert.NoError(t, g.onKey(&GocuiEvent{Type: eventKey, Key: key}))
	assert.Equal(t, 0, dispatched)
	consume = false
	assert.NoError(t, g.onKey(&GocuiEvent{Type: eventKey, Key: key}))
	assert.Equal(t, 1, dispatched)
	assert.NoError(t, g.onKey(&GocuiEvent{Type: eventKey, Key: NewKeyRune('z')}))
	assert.Equal(t, []Key{key, key, NewKeyRune('z')}, seen)
}

func TestKeybindingMiddlewarePreservesHandlerErrors(t *testing.T) {
	g := &Gui{}
	expected := errors.New("binding failed")
	key := NewKeyRune('g')
	g.SetKeybinding("", key, func(*Gui, *View) error { return expected })
	g.KeybindingMiddleware = func(_ Key, dispatch func() error) error { return dispatch() }

	assert.ErrorIs(t, g.onKey(&GocuiEvent{Type: eventKey, Key: key}), expected)
}
