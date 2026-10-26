package gui

import (
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/patch_exploring"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type vimListLocation struct {
	key      types.ContextKey
	item     string
	index    int
	path     string
	filter   string
	position navigationViewPosition
}

type vimCommitFilesLocation struct {
	ref         models.Ref
	refRange    *types.RefRange
	canRebase   bool
	windowName  string
	titlePrefix string
	parent      vimListLocation
}

type vimJumpLocation struct {
	key         types.ContextKey
	side        vimListLocation
	commitFiles *vimCommitFilesLocation
	position    navigationViewPosition
	patchState  *patch_exploring.State
}

func captureNavigationView(view *gocui.View) navigationViewPosition {
	ox, oy := view.Origin()
	cx, cy := view.Cursor()
	return navigationViewPosition{originX: ox, originY: oy, cursorX: cx, cursorY: cy}
}

func captureVimList(ctx types.IListContext) vimListLocation {
	if locker, ok := ctx.(interface {
		RLock()
		RUnlock()
	}); ok {
		locker.RLock()
		defer locker.RUnlock()
	}
	list := ctx.GetList()
	result := vimListLocation{key: ctx.GetKey(), item: ctx.GetSelectedItemId(), index: list.GetSelectedLineIdx(), position: captureNavigationView(ctx.GetView())}
	switch files := ctx.(type) {
	case *context.WorkingTreeContext:
		if node := files.GetSelected(); node != nil {
			result.path = node.GetInternalPath()
		}
	case *context.CommitFilesContext:
		if node := files.GetSelected(); node != nil {
			result.path = node.GetInternalPath()
		}
	}
	if filterable, ok := ctx.(types.IFilterableContext); ok {
		result.filter = filterable.GetFilter()
	}
	return result
}

func (gui *Gui) captureVimJump() (vimJumpLocation, bool) {
	// A quick second history key must use the requested destination, even if
	// its asynchronous diff has not replaced the previous view content yet.
	if gui.State.pendingVimJump != nil {
		return *gui.State.pendingVimJump, true
	}
	current := gui.State.ContextMgr.Current()
	side := gui.State.ContextMgr.CurrentSide()
	if !context.SupportsVimNavigation(current, side) {
		return vimJumpLocation{}, false
	}
	list, ok := side.(types.IListContext)
	if !ok {
		return vimJumpLocation{}, false
	}
	location := vimJumpLocation{key: current.GetKey(), side: captureVimList(list), position: captureNavigationView(current.GetView())}
	if side.GetKey() == context.COMMIT_FILES_CONTEXT_KEY {
		files := gui.State.Contexts.CommitFiles
		location.commitFiles = &vimCommitFilesLocation{
			ref: files.GetRef(), refRange: files.GetRefRange(), canRebase: files.GetCanRebase(),
			windowName: files.GetWindowName(), titlePrefix: files.GetView().TitlePrefix,
			parent: captureVimList(gui.State.Contexts.LocalCommits),
		}
	}
	if patchContext, ok := current.(types.IPatchExplorerContext); ok {
		patchContext.GetMutex().Lock()
		if state := patchContext.GetState(); state != nil {
			copy := *state
			location.patchState = &copy
		}
		patchContext.GetMutex().Unlock()
	}
	return location, true
}

func equalVimJumps(a, b vimJumpLocation) bool {
	return sameVimJumpTarget(a, b) && a.position == b.position
}

func sameVimJumpTarget(a, b vimJumpLocation) bool {
	if a.key != b.key || a.side.key != b.side.key || a.side.item != b.side.item || a.side.filter != b.side.filter {
		return false
	}
	if (a.commitFiles == nil) != (b.commitFiles == nil) {
		return false
	}
	if a.commitFiles != nil {
		if a.commitFiles.ref.RefName() != b.commitFiles.ref.RefName() {
			return false
		}
		ar, br := a.commitFiles.refRange, b.commitFiles.refRange
		if (ar == nil) != (br == nil) || ar != nil && (ar.From.RefName() != br.From.RefName() || ar.To.RefName() != br.To.RefName()) {
			return false
		}
	}
	return true
}

func (gui *Gui) isVimJumpKey(key gocui.Key) bool {
	opts := gui.keybindingOpts()
	u := opts.Config.Universal
	bindings := []config.Keybinding{{"g", "G", "o", "p"}, u.GoInto, u.Return, u.FocusMainView, u.GotoTop, u.GotoBottom, u.PrevBlock, u.NextBlock, u.NextMatch, u.PrevMatch}
	bindings = append(bindings, u.JumpToBlock...)
	for _, binding := range bindings {
		for _, candidate := range opts.GetKeys(binding) {
			if key.Equals(candidate) {
				return true
			}
		}
	}
	return false
}

func (gui *Gui) vimJumpHistory() *navigationHistory[vimJumpLocation] {
	if gui.State.vimHistory == nil {
		gui.State.vimHistory = newNavigationHistory(equalVimJumps)
	}
	return gui.State.vimHistory
}

func (gui *Gui) navigateVimHistory(forward bool) error {
	current, ok := gui.captureVimJump()
	if !ok {
		return nil
	}
	history := gui.vimJumpHistory()
	previousEntries := append([]vimJumpLocation(nil), history.entries...)
	previousCursor := history.cursor
	var destination vimJumpLocation
	if forward {
		destination, ok = history.Forward(current)
	} else {
		destination, ok = history.Back(current)
	}
	if !ok {
		return nil
	}
	previousRestore, previousJump := gui.State.pendingNavigationRestore, gui.State.pendingVimJump
	gui.cancelVimRestore()
	if err := gui.restoreVimJump(destination); err != nil {
		history.entries, history.cursor = previousEntries, previousCursor
		gui.State.pendingNavigationRestore, gui.State.pendingVimJump = previousRestore, previousJump
		return err
	}
	return nil
}

func (gui *Gui) cancelVimRestore() {
	gui.State.pendingNavigationRestore = nil
	gui.State.pendingVimJump = nil
}

// Restore selection by identity, so a refresh or a collapsed directory does not
// turn a saved file into a different row at the same index.
func (gui *Gui) restoreVimList(location vimListLocation) (types.IListContext, bool) {
	ctx := gui.State.ContextMgr.ContextForKey(location.key).(types.IListContext)
	if locker, ok := ctx.(interface {
		Lock()
		Unlock()
	}); ok {
		locker.Lock()
		defer locker.Unlock()
	}
	if filterable, ok := ctx.(types.IFilterableContext); ok && filterable.GetFilter() != location.filter {
		filterable.SetFilter(location.filter, gui.UserConfig().Gui.UseFuzzySearch())
	}
	list := ctx.GetList()
	index := location.index
	foundItem := location.item == "" && list.Len() == 0
	if tree, ok := list.(interface {
		ExpandToPath(string)
		GetIndexForPath(string) (int, bool)
	}); ok && location.path != "" {
		tree.ExpandToPath(location.path)
		if found, ok := tree.GetIndexForPath(location.path); ok {
			index = found
			foundItem = true
		}
	} else if commits, ok := ctx.(*context.LocalCommitsContext); ok {
		for i, item := range commits.GetItems() {
			if item.ID() == location.item {
				index = i
				foundItem = true
				break
			}
		}
	}
	list.SetSelection(index)
	return ctx, foundItem
}

func (gui *Gui) restoreVimJump(location vimJumpLocation) error {
	if saved := location.commitFiles; saved != nil {
		files := gui.State.Contexts.CommitFiles
		// Load first: a commit removed by a rebase must not leave the current
		// file list claiming to belong to a revision we could not read.
		from, to := saved.ref.ParentRefName(), saved.ref.RefName()
		if saved.refRange != nil {
			from, to = saved.refRange.From.ParentRefName(), saved.refRange.To.RefName()
		}
		from, reverse := gui.State.Modes.Diffing.GetFromAndReverseArgsForDiff(from)
		items, err := gui.git.Loaders.CommitFileLoader.GetFilesInDiff(from, to, reverse)
		if err != nil {
			return err
		}
		parent, parentFound := gui.restoreVimList(saved.parent)
		parent.HandleRender()
		files.ReInit(saved.ref, saved.refRange)
		files.SetCanRebase(saved.canRebase && parentFound)
		files.SetParentContext(gui.State.Contexts.LocalCommits)
		files.SetWindowName(saved.windowName)
		files.GetView().TitlePrefix = saved.titlePrefix
		gui.State.Model.CommitFiles = items
		files.SetTree()
	}
	side, foundItem := gui.restoreVimList(location.side)
	side.HandleRender()
	target := gui.State.ContextMgr.ContextForKey(location.key)
	if !foundItem {
		// If a file was removed or committed since visiting it, return to its
		// list instead of applying that file's diff position to a different file.
		target = side
		location.key = side.GetKey()
		location.side = captureVimList(side)
		location.position = location.side.position
		location.patchState = nil
	}
	state := gui.State
	gui.State.pendingVimJump = &location
	if target.GetKind() == types.MAIN_CONTEXT {
		gui.queueNavigationRestore(target.GetView(), location.position, func() {
			gui.State.pendingVimJump = nil
			if patchContext, ok := target.(types.IPatchExplorerContext); ok && location.patchState != nil {
				patchContext.GetMutex().Lock()
				defer patchContext.GetMutex().Unlock()
				if current := patchContext.GetState(); current != nil {
					if current.GetDiff() == location.patchState.GetDiff() {
						copy := *location.patchState
						copy.OnViewWidthChanged(target.GetView())
						patchContext.SetState(&copy)
						patchContext.RenderAndFocus()
					} else {
						patchContext.NavigateTo(location.patchState.GetSelectedViewLineIdx())
					}
				}
			}
		})
	}
	if gui.State.ContextMgr.Current() == side {
		side.HandleFocus(types.OnFocusOpts{ScrollSelectionIntoView: true})
	} else {
		gui.State.ContextMgr.Push(side, types.OnFocusOpts{ScrollSelectionIntoView: true})
	}
	if target != side {
		gui.State.ContextMgr.Push(target, types.OnFocusOpts{})
	}
	if gui.State.ContextMgr.Current() != target {
		// Staging or patch building can return to the list if its diff is no
		// longer available. Do not keep a pending destination that never opened.
		gui.cancelVimRestore()
		return nil
	}
	gui.c.AfterLayout(func() error {
		if gui.State != state || gui.State.ContextMgr.Current() != target || gui.State.pendingVimJump != &location {
			return nil
		}
		view := side.GetView()
		selectedLine := side.ModelIndexToViewIndex(side.GetList().GetSelectedLineIdx())
		originY := max(0, selectedLine-location.side.position.cursorY)
		view.SetOrigin(location.side.position.originX, originY)
		view.SetCursor(location.side.position.cursorX, selectedLine-originY)
		side.HandleRender()
		if target == side {
			gui.State.pendingVimJump = nil
		}
		return nil
	})
	return nil
}

func (gui *Gui) vimJumpKeybindings() []*types.Binding {
	bindings := []*types.Binding{}
	for _, key := range []types.ContextKey{context.FILES_CONTEXT_KEY, context.LOCAL_COMMITS_CONTEXT_KEY, context.COMMIT_FILES_CONTEXT_KEY, context.NORMAL_MAIN_CONTEXT_KEY, context.NORMAL_SECONDARY_CONTEXT_KEY, context.STAGING_MAIN_CONTEXT_KEY, context.STAGING_SECONDARY_CONTEXT_KEY, context.PATCH_BUILDING_MAIN_CONTEXT_KEY} {
		ctx := gui.State.ContextMgr.ContextForKey(key)
		for _, forward := range []bool{false, true} {
			keys := []gocui.Key{gocui.NewKeyStrMod("o", gocui.ModCtrl)}
			description := "Jump back to the previous location"
			if forward {
				keys = []gocui.Key{gocui.NewKeyStrMod("i", gocui.ModCtrl), gocui.NewKeyName(gocui.KeyTab)}
				description = "Jump forward to the next location"
			}
			bindings = append(bindings, &types.Binding{
				ViewName: ctx.GetViewName(), Keys: keys, Description: description, Tag: "navigation",
				Handler: func() error { return gui.navigateVimHistory(forward) },
				GetDisabledReason: func() *types.DisabledReason {
					if !context.SupportsVimNavigation(gui.State.ContextMgr.Current(), gui.State.ContextMgr.CurrentSide()) {
						return &types.DisabledReason{AllowFurtherDispatching: true}
					}
					return nil
				},
			})
		}
	}
	return bindings
}
