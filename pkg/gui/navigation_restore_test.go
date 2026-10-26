package gui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func TestNavigationRestorePreloadsSavedViewport(t *testing.T) {
	gui := &Gui{State: &GuiRepoState{}}
	view := gocui.NewView("main", 0, 0, 80, 20, gocui.OutputNormal)
	otherView := gocui.NewView("secondary", 0, 0, 80, 20, gocui.OutputNormal)
	ordinaryRead := gui.linesToReadFromCmdTask(view)

	gui.queueNavigationRestore(view, navigationViewPosition{originY: 6000}, nil)
	assert.GreaterOrEqual(t, gui.linesToReadFromCmdTask(view).Total, 6000+view.InnerHeight()+10)
	assert.Equal(t, ordinaryRead, gui.linesToReadFromCmdTask(otherView))

	gui.queueNavigationRestore(view, navigationViewPosition{originY: 1}, nil)
	assert.Equal(t, ordinaryRead, gui.linesToReadFromCmdTask(view))
}

func navigationRestoreTestGui(view *gocui.View) *Gui {
	return &Gui{State: &GuiRepoState{ContextMgr: &ContextMgr{
		ContextStack: []types.Context{context.NewSimpleContext(context.NewBaseContext(context.NewBaseContextOpts{View: view}))},
	}}}
}

func TestNavigationRestoreWaitsForCurrentRequest(t *testing.T) {
	for _, scenario := range []string{"canceled", "replaced", "other repository", "other view"} {
		t.Run(scenario, func(t *testing.T) {
			view := gocui.NewView("main", 0, 0, 80, 20, gocui.OutputNormal)
			gui := navigationRestoreTestGui(view)
			applied := false
			gui.queueNavigationRestore(view, navigationViewPosition{originY: 100}, func() { applied = true })
			original := gui.State.pendingNavigationRestore
			switch scenario {
			case "canceled":
				gui.State.pendingNavigationRestore = nil
			case "replaced":
				gui.queueNavigationRestore(view, navigationViewPosition{originY: 200}, nil)
			case "other repository":
				gui.State = &GuiRepoState{}
			case "other view":
				otherView := gocui.NewView("secondary", 0, 0, 80, 20, gocui.OutputNormal)
				gui.State.ContextMgr = navigationRestoreTestGui(otherView).State.ContextMgr
			}
			pending := gui.State.pendingNavigationRestore
			gui.applyNavigationRestore(original, 0)
			assert.False(t, applied)
			assert.Zero(t, view.OriginY())
			assert.Equal(t, pending, gui.State.pendingNavigationRestore)
		})
	}
}

func TestNavigationRestorePreservesLoadedViewport(t *testing.T) {
	view := gocui.NewView("main", 0, 0, 80, 20, gocui.OutputNormal)
	view.SetContent(strings.Repeat("line\n", 200))
	gui := navigationRestoreTestGui(view)
	position := navigationViewPosition{originX: 3, originY: 100, cursorX: 2, cursorY: 4}
	applyCount := 0
	gui.queueNavigationRestore(view, position, func() {
		applyCount++
		view.SetOrigin(0, 0)
	})
	request := gui.State.pendingNavigationRestore
	gui.applyNavigationRestore(request, 0)
	assert.Equal(t, position, captureNavigationView(view))
	assert.Nil(t, gui.State.pendingNavigationRestore)
	gui.applyNavigationRestore(request, 0)
	assert.Equal(t, 1, applyCount)
}

func TestNavigationRestoreClampsShortenedDiff(t *testing.T) {
	view := gocui.NewView("main", 0, 0, 80, 20, gocui.OutputNormal)
	view.SetContent("one\ntwo\nthree")
	gui := navigationRestoreTestGui(view)
	gui.queueNavigationRestore(view, navigationViewPosition{originY: 100, cursorY: 4}, nil)
	gui.applyNavigationRestore(gui.State.pendingNavigationRestore, 0)
	assert.Zero(t, view.OriginY())
	assert.Equal(t, 2, view.CursorY())
}

func TestNavigationRestoreWaitsForLatestRenderOfSameRequest(t *testing.T) {
	view := gocui.NewView("main", 0, 0, 80, 20, gocui.OutputNormal)
	view.SetContent(strings.Repeat("line\n", 200))
	gui := navigationRestoreTestGui(view)
	position := navigationViewPosition{originY: 100, cursorY: 4}
	applyCount := 0
	gui.queueNavigationRestore(view, position, func() { applyCount++ })
	request := gui.State.pendingNavigationRestore
	assert.NotNil(t, gui.navigationContentReady(view))
	firstRender := request.renderGeneration
	// A background refresh creates a second render after the first task became
	// ready but before its queued UI callback gets a chance to restore the view.
	assert.NotNil(t, gui.navigationContentReady(view))
	secondRender := request.renderGeneration
	gui.applyNavigationRestore(request, firstRender)
	assert.Zero(t, applyCount)
	assert.Zero(t, view.OriginY())
	assert.Same(t, request, gui.State.pendingNavigationRestore)

	gui.applyNavigationRestore(request, secondRender)
	assert.Equal(t, 1, applyCount)
	assert.Equal(t, position, captureNavigationView(view))
	assert.Nil(t, gui.State.pendingNavigationRestore)
}
