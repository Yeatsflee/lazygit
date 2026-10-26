package gui

import "github.com/jesseduffield/lazygit/pkg/gocui"

type navigationViewPosition struct {
	originX int
	originY int
	cursorX int
	cursorY int
}

type navigationRestoreRequest struct {
	view             *gocui.View
	position         navigationViewPosition
	apply            func()
	state            *GuiRepoState
	renderGeneration uint64
}

func (gui *Gui) queueNavigationRestore(view *gocui.View, position navigationViewPosition, apply func()) {
	gui.State.pendingNavigationRestore = &navigationRestoreRequest{
		view:     view,
		position: position,
		apply:    apply,
		state:    gui.State,
	}
}

// Capture the request and render when creating the task. A later navigation or
// refresh must not let an old task restore a viewport that is still loading.
func (gui *Gui) navigationContentReady(view *gocui.View) func() {
	request := gui.State.pendingNavigationRestore
	if request == nil || request.view != view {
		return nil
	}
	request.renderGeneration++
	generation := request.renderGeneration

	return func() {
		gui.c.OnUIThread(func() error {
			gui.c.AfterLayout(func() error {
				gui.applyNavigationRestore(request, generation)
				return nil
			})
			return nil
		})
	}
}

func (gui *Gui) applyNavigationRestore(request *navigationRestoreRequest, generation uint64) {
	if gui.State != request.state ||
		gui.State.pendingNavigationRestore != request ||
		request.renderGeneration != generation ||
		gui.State.ContextMgr.Current().GetView() != request.view {
		return
	}

	if request.apply != nil {
		request.apply()
	}
	view := request.view
	position := request.position
	// A working-tree diff may have become shorter since the jump was recorded.
	// Use the same bounds as ordinary scrolling to avoid restoring a blank page.
	view.SetOrigin(position.originX, 0)
	view.ScrollDown(position.originY)
	cursorY := position.cursorY
	if view.OriginY() != position.originY {
		selectedLine := min(max(0, position.originY+cursorY), max(0, view.ViewLinesHeight()-1))
		cursorY = selectedLine - view.OriginY()
	}
	view.SetCursor(position.cursorX, cursorY)
	gui.State.pendingNavigationRestore = nil
}
