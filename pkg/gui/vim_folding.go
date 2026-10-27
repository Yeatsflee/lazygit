package gui

import (
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
)

// Only the directory trees in windows 2 and 4 use the z prefix. Flat lists,
// main views, and files reached from other sources keep their original keys.
func (gui *Gui) supportsVimFolding() bool {
	switch current := gui.State.ContextMgr.Current().(type) {
	case *context.WorkingTreeContext:
		return current.InTreeMode()
	case *context.CommitFilesContext:
		parent := current.GetParentContext()
		return parent != nil && parent.GetKey() == context.LOCAL_COMMITS_CONTEXT_KEY && current.InTreeMode()
	}
	return false
}

func (gui *Gui) handleVimFoldKey(key gocui.Key) bool {
	var action filetree.FoldAction
	switch {
	case key.Equals(gocui.NewKeyRune('a')):
		action = filetree.FoldToggle
	case key.Equals(gocui.NewKeyRune('c')):
		action = filetree.FoldClose
	case key.Equals(gocui.NewKeyRune('o')):
		action = filetree.FoldOpen
	case key.Equals(gocui.NewKeyRune('M')):
		action = filetree.FoldCloseAll
	case key.Equals(gocui.NewKeyRune('R')):
		action = filetree.FoldOpenAll
	default:
		return false
	}

	gui.cancelVimRestore()
	current := gui.State.ContextMgr.Current()
	changed := false
	switch current := current.(type) {
	case *context.WorkingTreeContext:
		changed = current.Fold(action)
	case *context.CommitFilesContext:
		changed = current.Fold(action)
	}
	if changed {
		gui.c.PostRefreshUpdate(current)
	}
	return true
}
