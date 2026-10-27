package filetree

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestParentDirectoryIndex(t *testing.T) {
	files := []*models.File{
		{Path: "a/b/c/one"},
		{Path: "a/b/c/two"},
		{Path: "a/d/three"},
		{Path: "root.txt"},
	}
	scenarios := []struct {
		name         string
		showRootItem bool
		flat         bool
		selectedPath string
		parentPath   string
	}{
		{
			name:         "file in compressed directory",
			selectedPath: "a/b/c/two",
			parentPath:   "a/b/c",
		},
		{
			name:         "compressed directory",
			selectedPath: "a/b/c",
			parentPath:   "a",
		},
		{
			name:         "visible root is the parent",
			showRootItem: true,
			selectedPath: "./a",
			parentPath:   ".",
		},
		{
			name:         "root has no parent",
			showRootItem: true,
			selectedPath: ".",
		},
		{
			name:         "top level directory has no visible parent",
			selectedPath: "a",
		},
		{
			name:         "top level file has no visible parent",
			selectedPath: "root.txt",
		},
		{
			name:         "flat list has no parent",
			flat:         true,
			selectedPath: "a/b/c/two",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := config.GetDefaultConfig()
			userConfig.Gui.ShowRootItemInFileTree = s.showRootItem
			cmn := common.NewDummyCommonWithUserConfigAndAppState(userConfig, nil)
			tree := NewFileTree(func() []*models.File { return files }, cmn, !s.flat)
			tree.SetTree()
			index, found := tree.GetIndexForPath(s.selectedPath)
			if !assert.True(t, found) {
				return
			}

			parent, found := ParentDirectoryIndex(tree, index)
			if s.parentPath == "" {
				assert.False(t, found)
			} else if assert.True(t, found) {
				assert.Equal(t, s.parentPath, tree.Get(parent).GetInternalPath())
			}
		})
	}
}

func TestParentDirectoryIndexWithoutRow(t *testing.T) {
	tree := NewFileTree(func() []*models.File { return nil }, common.NewDummyCommon(), true)
	tree.SetTree()
	for _, index := range []int{-1, 0, 1} {
		_, found := ParentDirectoryIndex(tree, index)
		assert.False(t, found)
	}
}
