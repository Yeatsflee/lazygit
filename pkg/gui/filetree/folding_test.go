package filetree

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/stretchr/testify/assert"
)

type foldingTestTree interface {
	foldingTree
	Fold(FoldAction) bool
	SetTree()
	GetSelectedItemId() string
	ToggleShowTree()
}

func foldingTestTrees(t *testing.T, paths []string, run func(*testing.T, foldingTestTree, bool)) {
	t.Helper()
	for _, showRoot := range []bool{false, true} {
		cfg := config.GetDefaultConfig()
		cfg.Gui.ShowRootItemInFileTree = showRoot
		cmn := common.NewDummyCommonWithUserConfigAndAppState(cfg, nil)
		files := []*models.File{}
		commitFiles := []*models.CommitFile{}
		for _, path := range paths {
			files = append(files, &models.File{Path: path})
			commitFiles = append(commitFiles, &models.CommitFile{Path: path})
		}
		for name, tree := range map[string]foldingTestTree{
			"files":        NewFileTreeViewModel(func() []*models.File { return files }, cmn, true),
			"commit files": NewCommitFileTreeViewModel(func() []*models.CommitFile { return commitFiles }, cmn, true),
		} {
			if showRoot {
				name += " with root"
			}
			t.Run(name, func(t *testing.T) {
				tree.SetTree()
				run(t, tree, showRoot)
			})
		}
	}
}

func TestFoldDirectoryFromFile(t *testing.T) {
	foldingTestTrees(t, []string{"dir/nested/one", "dir/nested/two", "other/file"}, func(t *testing.T, tree foldingTestTree, root bool) {
		path := InternalTreePathForFilePath("dir/nested/two", root)
		dir := InternalTreePathForFilePath("dir/nested", root)
		index, found := tree.GetIndexForPath(path)
		assert.True(t, found)
		tree.SetSelection(index)
		assert.False(t, tree.Fold(FoldOpen))
		assert.Equal(t, index, tree.GetSelectedLineIdx())
		assert.True(t, tree.Fold(FoldClose))
		assert.True(t, tree.CollapsedPaths().IsCollapsed(dir))
		_, found = tree.GetIndexForPath(path)
		assert.False(t, found)
		index, found = tree.GetIndexForPath(dir)
		assert.True(t, found)
		assert.Equal(t, index, tree.GetSelectedLineIdx())
		assert.False(t, tree.Fold(FoldClose))
		assert.True(t, tree.Fold(FoldOpen))
		assert.False(t, tree.Fold(FoldOpen))
		assert.True(t, tree.Fold(FoldToggle))
		assert.True(t, tree.CollapsedPaths().IsCollapsed(dir))
		assert.True(t, tree.Fold(FoldToggle))
		assert.False(t, tree.CollapsedPaths().IsCollapsed(dir))
	})
}

func TestFoldAllIncludesHiddenDescendantsAndPreservesSelection(t *testing.T) {
	foldingTestTrees(t, []string{"first/file", "last/compressed/deep/file", "last/compressed/other"}, func(t *testing.T, tree foldingTestTree, root bool) {
		dir := InternalTreePathForFilePath("last/compressed", root)
		deep := InternalTreePathForFilePath("last/compressed/deep", root)
		index, found := tree.GetIndexForPath(dir)
		assert.True(t, found)
		tree.SetSelection(index)
		assert.True(t, tree.Fold(FoldClose))
		assert.False(t, tree.CollapsedPaths().IsCollapsed(deep))
		assert.True(t, tree.Fold(FoldCloseAll))
		assert.True(t, tree.CollapsedPaths().IsCollapsed(deep))
		selectedPath := dir
		if root {
			selectedPath = "."
		}
		index, found = tree.GetIndexForPath(selectedPath)
		assert.True(t, found)
		assert.Equal(t, index, tree.GetSelectedLineIdx())
		assert.True(t, tree.Fold(FoldOpen))
		assert.True(t, tree.CollapsedPaths().IsCollapsed(deep))
		assert.True(t, tree.Fold(FoldOpenAll))
		assert.False(t, tree.CollapsedPaths().IsCollapsed(deep))
		index, found = tree.GetIndexForPath(selectedPath)
		assert.True(t, found)
		assert.Equal(t, index, tree.GetSelectedLineIdx())
	})
}

func TestFoldingEmptyAndFlatTrees(t *testing.T) {
	foldingTestTrees(t, nil, func(t *testing.T, tree foldingTestTree, _ bool) {
		for _, action := range []FoldAction{FoldToggle, FoldClose, FoldOpen, FoldCloseAll, FoldOpenAll} {
			assert.False(t, tree.Fold(action))
		}
	})
	foldingTestTrees(t, []string{"root.txt", "dir/file"}, func(t *testing.T, tree foldingTestTree, root bool) {
		index, found := tree.GetIndexForPath(InternalTreePathForFilePath("root.txt", root))
		assert.True(t, found)
		tree.SetSelection(index)
		if !root {
			assert.False(t, tree.Fold(FoldToggle))
			assert.False(t, tree.Fold(FoldClose))
			assert.False(t, tree.Fold(FoldOpen))
		}
		tree.ToggleShowTree()
		selected := tree.GetSelectedItemId()
		for _, action := range []FoldAction{FoldToggle, FoldClose, FoldOpen, FoldCloseAll, FoldOpenAll} {
			assert.False(t, tree.Fold(action))
		}
		assert.Equal(t, selected, tree.GetSelectedItemId())
	})
}
