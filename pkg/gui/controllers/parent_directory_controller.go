package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

// ParentDirectoryController navigates upwards without changing the repository.
type ParentDirectoryController struct {
	baseController
	c       *ControllerCommon
	context types.Context
}

var _ types.IController = &ParentDirectoryController{}

func NewParentDirectoryController(c *ControllerCommon, context types.Context) *ParentDirectoryController {
	return &ParentDirectoryController{c: c, context: context}
}

func (self *ParentDirectoryController) Context() types.Context {
	return self.context
}

func (self *ParentDirectoryController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Keys:              opts.GetKeys(config.Keybinding{"p"}),
			Handler:           self.goToParent,
			GetDisabledReason: self.disabledReason,
			Description:       "Go to parent directory or file list",
			Tag:               "navigation",
		},
	}
}

func (self *ParentDirectoryController) disabledReason() *types.DisabledReason {
	if context.SupportsVimNavigation(self.c.Context().Current(), self.c.Context().CurrentSide()) {
		return nil
	}
	return &types.DisabledReason{AllowFurtherDispatching: true}
}

func (self *ParentDirectoryController) goToParent() error {
	switch self.context.GetKey() {
	case context.FILES_CONTEXT_KEY:
		files := self.c.Contexts().Files
		files.Lock()
		selectedParent := selectParentDirectory(files.FileTreeViewModel)
		files.Unlock()
		if selectedParent {
			self.c.PostRefreshUpdate(files)
		}
	case context.COMMIT_FILES_CONTEXT_KEY:
		files := self.c.Contexts().CommitFiles
		files.Lock()
		selectedParent := selectParentDirectory(files.CommitFileTreeViewModel)
		files.Unlock()
		if selectedParent {
			self.c.PostRefreshUpdate(files)
		} else {
			self.c.Context().Push(files.GetParentContext(), types.OnFocusOpts{})
		}
	case context.LOCAL_COMMITS_CONTEXT_KEY:
		// The commit list is already at the top of its hierarchy.
	default:
		self.c.Context().Pop()
	}
	return nil
}

type parentDirectoryTree interface {
	types.IListCursor
	InTreeMode() bool
	GetVisualDepth(int) int
	Len() int
}

func selectParentDirectory(tree parentDirectoryTree) bool {
	if !tree.InTreeMode() || tree.Len() == 0 {
		return false
	}
	selectedIndex := tree.GetSelectedLineIdx()
	selectedDepth := tree.GetVisualDepth(selectedIndex)
	// A rendered directory can combine several path components. Its parent is
	// the nearest preceding row with a shallower visual depth.
	for index := selectedIndex - 1; index >= 0; index-- {
		if tree.GetVisualDepth(index) < selectedDepth {
			tree.SetSelection(index)
			return true
		}
	}
	return false
}
