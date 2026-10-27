package gui

import (
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
)

func (gui *Gui) handleVimNavigationKey(key gocui.Key, dispatch func() error) error {
	if gui.State == nil {
		return dispatch()
	}
	gui.State.vimKeyGeneration++

	pendingTop := gui.State.vimPendingTop
	pendingFold := gui.State.vimPendingFold
	gui.State.vimPendingTop = false
	gui.State.vimPendingFold = false
	if gui.g.IsPasting || !context.SupportsVimNavigation(
		gui.State.ContextMgr.Current(), gui.State.ContextMgr.CurrentSide(),
	) {
		gui.cancelVimRestore()
		return dispatch()
	}

	if gui.supportsVimFolding() {
		if pendingFold && gui.handleVimFoldKey(key) {
			return nil
		}
		if key.Equals(gocui.NewKeyRune('z')) {
			if pendingFold {
				// zz preserves the original z binding (normally reflog undo).
				gui.cancelVimRestore()
				return dispatch()
			}
			gui.State.vimPendingFold = true
			return nil
		}
		if pendingFold && key.Equals(gocui.NewKeyName(gocui.KeyEsc)) {
			return nil
		}
	}

	if key.Equals(gocui.NewKeyRune('g')) && !pendingTop {
		gui.State.vimPendingTop = true
		return nil
	}
	if key.Equals(gocui.NewKeyStrMod("o", gocui.ModCtrl)) {
		return gui.navigateVimHistory(false)
	}
	if key.Equals(gocui.NewKeyStrMod("i", gocui.ModCtrl)) || key.Equals(gocui.NewKeyName(gocui.KeyTab)) {
		return gui.navigateVimHistory(true)
	}
	currentKey := gui.State.ContextMgr.Current().GetKey()
	if currentKey == context.NORMAL_MAIN_CONTEXT_KEY || currentKey == context.NORMAL_SECONDARY_CONTEXT_KEY {
		for _, bottom := range append([]gocui.Key{gocui.NewKeyRune('G')}, gui.keybindingOpts().GetKeys(gui.UserConfig().Keybinding.Universal.GotoBottom)...) {
			if key.Equals(bottom) {
				return dispatch()
			}
		}
	}
	var source vimJumpLocation
	record := false
	if gui.isVimJumpKey(key) {
		source, record = gui.captureVimJump()
	}
	state := gui.State
	gui.cancelVimRestore()
	if err := dispatch(); err != nil {
		return err
	}
	if record && gui.State == state {
		gui.recordVimJumpFrom(source)
	}
	return nil
}

func (gui *Gui) recordVimJumpFrom(source vimJumpLocation) {
	destination, ok := gui.captureVimJump()
	if !ok || !equalVimJumps(source, destination) {
		gui.vimJumpHistory().Record(source)
	}
}

// Reading a long diff to its end can finish after another navigation key. Only
// apply the scroll, and commit its history entry, while it is still current.
func (gui *Gui) prepareVimGotoBottom() func(func()) {
	source, ok := gui.captureVimJump()
	if !ok {
		return nil
	}
	state, generation := gui.State, gui.State.vimKeyGeneration
	gui.cancelVimRestore()
	return func(scroll func()) {
		if gui.State != state || gui.State.vimKeyGeneration != generation {
			return
		}
		current, currentOK := gui.captureVimJump()
		if !currentOK || !sameVimJumpTarget(source, current) {
			return
		}
		scroll()
		gui.recordVimJumpFrom(source)
	}
}
