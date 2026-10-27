package filetree

import "github.com/jesseduffield/lazygit/pkg/gui/types"

type FoldAction int

const (
	FoldToggle FoldAction = iota
	FoldClose
	FoldOpen
	FoldCloseAll
	FoldOpenAll
)

type foldingTree interface {
	parentDirectoryTree
	types.IListCursor
	CollapsedPaths() *CollapsedPaths
	GetIndexForPath(string) (int, bool)
	ExpandAll()
}

func (self *FileTreeViewModel) Fold(action FoldAction) bool {
	self.Lock()
	defer self.Unlock()
	return foldTree(self, self.GetRoot().Raw(), action)
}

func (self *CommitFileTreeViewModel) Fold(action FoldAction) bool {
	self.Lock()
	defer self.Unlock()
	return foldTree(self, self.GetRoot().Raw(), action)
}

// foldTree keeps the selection on the directory being closed. A file belongs
// to its nearest visible directory, which may combine several path components.
func foldTree[T any](tree foldingTree, root *Node[T], action FoldAction) bool {
	if !tree.InTreeMode() || tree.Len() == 0 {
		return false
	}
	index := tree.GetSelectedLineIdx()
	paths := tree.CollapsedPaths()
	node := root.GetNodeAtIndex(index+1, paths)

	switch action {
	case FoldOpenAll:
		tree.ExpandAll()
		tree.SetSelection(tree.GetSelectedLineIdx())
		return true
	case FoldCloseAll:
		// Remember the topmost visible ancestor before closing the whole tree.
		// Unlike splitting the path, this also handles compressed directories.
		for {
			parent, found := ParentDirectoryIndex(tree, index)
			if !found {
				break
			}
			index = parent
		}
		path := root.GetNodeAtIndex(index+1, paths).GetInternalPath()
		// Visit hidden descendants too, so reopening one level cannot reveal
		// directories that escaped an earlier "close all" operation.
		for _, child := range root.Children {
			closeDirectoryTree(child, paths)
		}
		if selected, found := tree.GetIndexForPath(path); found {
			tree.SetSelection(selected)
		}
		return true
	}

	if node.IsFile() {
		parent, found := ParentDirectoryIndex(tree, index)
		if !found {
			return false
		}
		index = parent
		node = root.GetNodeAtIndex(index+1, paths)
	}
	path := node.GetInternalPath()
	collapsed := paths.IsCollapsed(path)
	if action == FoldClose && collapsed || action == FoldOpen && !collapsed {
		return false
	}
	tree.SetSelection(index)
	paths.ToggleCollapsed(path)
	return true
}

func closeDirectoryTree[T any](node *Node[T], paths *CollapsedPaths) {
	if node.IsFile() {
		return
	}
	paths.Collapse(node.GetInternalPath())
	for _, child := range node.Children {
		closeDirectoryTree(child, paths)
	}
}
