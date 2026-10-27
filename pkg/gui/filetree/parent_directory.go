package filetree

type parentDirectoryTree interface {
	InTreeMode() bool
	GetVisualDepth(int) int
	Len() int
}

// ParentDirectoryIndex returns the nearest visible parent directory of a row.
// A rendered directory can combine several path components, so its parent is
// the nearest preceding row with a shallower visual depth.
func ParentDirectoryIndex(tree parentDirectoryTree, index int) (int, bool) {
	if !tree.InTreeMode() || index < 0 || index >= tree.Len() {
		return 0, false
	}

	depth := tree.GetVisualDepth(index)
	for parent := index - 1; parent >= 0; parent-- {
		if tree.GetVisualDepth(parent) < depth {
			return parent, true
		}
	}
	return 0, false
}
