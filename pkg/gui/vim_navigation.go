package gui

import (
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
)

func (gui *Gui) handleVimNavigationKey(key gocui.Key, dispatch func() error) error {
	if gui.State == nil {
		return dispatch()
	}

	pendingTop := gui.State.vimPendingTop
	gui.State.vimPendingTop = false
	if gui.g.IsPasting || !context.SupportsVimNavigation(
		gui.State.ContextMgr.Current(), gui.State.ContextMgr.CurrentSide(),
	) {
		return dispatch()
	}

	if key.Equals(gocui.NewKeyRune('g')) && !pendingTop {
		gui.State.vimPendingTop = true
		return nil
	}
	return dispatch()
}
