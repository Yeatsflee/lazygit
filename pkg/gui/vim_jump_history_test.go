package gui

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/stretchr/testify/assert"
)

func TestVimHistoryCapturesPendingDestinationDuringRapidNavigation(t *testing.T) {
	location := func(file string, origin int) vimJumpLocation {
		return vimJumpLocation{
			key:      context.NORMAL_MAIN_CONTEXT_KEY,
			side:     vimListLocation{key: context.FILES_CONTEXT_KEY, item: file},
			position: navigationViewPosition{originY: origin},
		}
	}
	a := location("file-a", 10)
	b := location("file-b", 20)
	c := location("file-c", 30)
	gui := &Gui{State: &GuiRepoState{pendingVimJump: &c}}
	history := gui.vimJumpHistory()
	history.Record(a)
	history.Record(b)

	for _, step := range []struct {
		forward bool
		want    vimJumpLocation
	}{
		{want: b},
		{want: a},
		{forward: true, want: b},
		{forward: true, want: c},
		{want: b},
	} {
		// The requested destination is authoritative while the old diff is
		// still on screen. None of these restores has completed yet.
		current, ok := gui.captureVimJump()
		assert.True(t, ok)
		var destination vimJumpLocation
		if step.forward {
			destination, ok = history.Forward(current)
		} else {
			destination, ok = history.Back(current)
		}
		assert.True(t, ok)
		assert.Equal(t, step.want, destination)
		gui.State.pendingVimJump = &destination
	}

	assert.Equal(t, []vimJumpLocation{a, b, c}, history.entries)
}

func TestVimNoopJumpPreservesForwardHistory(t *testing.T) {
	a := vimJumpLocation{key: context.NORMAL_MAIN_CONTEXT_KEY, side: vimListLocation{item: "file-a"}}
	b := a
	b.position.originY = 20
	gui := &Gui{State: &GuiRepoState{pendingVimJump: &a}}
	history := gui.vimJumpHistory()
	history.Record(a)
	_, ok := history.Back(b)
	assert.True(t, ok)

	finish := gui.prepareVimGotoBottom()
	gui.State.pendingVimJump = &a
	finish(func() {})

	destination, ok := history.Forward(a)
	assert.True(t, ok)
	assert.Equal(t, b, destination)
}

func TestVimBottomCompletionIgnoresSupersededNavigation(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*Gui)
	}{
		{name: "another key", change: func(gui *Gui) { gui.State.vimKeyGeneration++ }},
		{name: "different file", change: func(gui *Gui) { gui.State.pendingVimJump.side.item = "file-b" }},
		{name: "different repository", change: func(gui *Gui) { gui.State = &GuiRepoState{} }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			source := vimJumpLocation{key: context.NORMAL_MAIN_CONTEXT_KEY, side: vimListLocation{item: "file-a"}}
			gui := &Gui{State: &GuiRepoState{pendingVimJump: &source}}
			finish := gui.prepareVimGotoBottom()
			current := source
			gui.State.pendingVimJump = &current
			scenario.change(gui)
			scrolled := false
			finish(func() { scrolled = true })
			assert.False(t, scrolled)
			assert.Nil(t, gui.State.vimHistory)
		})
	}
}
