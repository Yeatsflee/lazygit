package gui

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/controllers/helpers"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func foldingContextTrait(key types.ContextKey) *context.ListContextTrait {
	return &context.ListContextTrait{Context: context.NewSimpleContext(context.NewBaseContext(context.NewBaseContextOpts{
		Key: key, Kind: types.SIDE_CONTEXT,
	}))}
}

func TestVimFoldingScope(t *testing.T) {
	cmn := common.NewDummyCommon()
	for _, test := range []struct {
		name   string
		key    types.ContextKey
		parent types.ContextKey
		flat   bool
		want   bool
	}{
		{name: "working tree", key: context.FILES_CONTEXT_KEY, want: true},
		{name: "flat files", key: context.FILES_CONTEXT_KEY, flat: true},
		{name: "local commit files", key: context.COMMIT_FILES_CONTEXT_KEY, parent: context.LOCAL_COMMITS_CONTEXT_KEY, want: true},
		{name: "flat commit files", key: context.COMMIT_FILES_CONTEXT_KEY, parent: context.LOCAL_COMMITS_CONTEXT_KEY, flat: true},
		{name: "branch files", key: context.COMMIT_FILES_CONTEXT_KEY, parent: context.LOCAL_BRANCHES_CONTEXT_KEY},
		{name: "stash files", key: context.COMMIT_FILES_CONTEXT_KEY, parent: context.STASH_CONTEXT_KEY},
		{name: "no parent", key: context.COMMIT_FILES_CONTEXT_KEY},
		{name: "commit list", key: context.LOCAL_COMMITS_CONTEXT_KEY},
		{name: "main", key: context.NORMAL_MAIN_CONTEXT_KEY},
		{name: "staging", key: context.STAGING_MAIN_CONTEXT_KEY},
		{name: "prompt", key: context.PROMPT_CONTEXT_KEY},
	} {
		t.Run(test.name, func(t *testing.T) {
			trait := foldingContextTrait(test.key)
			var current types.Context = trait.Context
			switch test.key {
			case context.FILES_CONTEXT_KEY:
				current = &context.WorkingTreeContext{ListContextTrait: trait, FileTreeViewModel: filetree.NewFileTreeViewModel(nil, cmn, !test.flat)}
			case context.COMMIT_FILES_CONTEXT_KEY:
				current = &context.CommitFilesContext{ListContextTrait: trait, CommitFileTreeViewModel: filetree.NewCommitFileTreeViewModel(nil, cmn, !test.flat)}
				if test.parent != "" {
					current.SetParentContext(foldingContextTrait(test.parent).Context)
				}
			}
			gui := &Gui{State: &GuiRepoState{ContextMgr: &ContextMgr{ContextStack: []types.Context{current}}}}
			assert.Equal(t, test.want, gui.supportsVimFolding())
		})
	}
}

func TestVimFoldPrefixCancellation(t *testing.T) {
	cmn := common.NewDummyCommon()
	files := &context.WorkingTreeContext{
		ListContextTrait:  foldingContextTrait(context.FILES_CONTEXT_KEY),
		FileTreeViewModel: filetree.NewFileTreeViewModel(nil, cmn, true),
	}
	gui := &Gui{
		g:     &gocui.Gui{},
		c:     &helpers.HelperCommon{Common: cmn},
		State: &GuiRepoState{ContextMgr: &ContextMgr{ContextStack: []types.Context{files}}},
	}
	dispatched := []gocui.Key{}
	press := func(key gocui.Key) {
		assert.NoError(t, gui.handleVimNavigationKey(key, func() error {
			dispatched = append(dispatched, key)
			return nil
		}))
	}
	z := gocui.NewKeyRune('z')
	press(z)
	assert.True(t, gui.State.vimPendingFold)
	assert.Empty(t, dispatched)
	press(z)
	assert.False(t, gui.State.vimPendingFold)
	assert.Equal(t, []gocui.Key{z}, dispatched)

	for _, cancel := range []gocui.Key{gocui.NewKeyName(gocui.KeyF12), gocui.NewKeyName(gocui.MouseLeft)} {
		press(z)
		press(cancel)
		assert.False(t, gui.State.vimPendingFold)
		assert.Equal(t, cancel, dispatched[len(dispatched)-1])
	}
	press(z)
	count := len(dispatched)
	press(gocui.NewKeyName(gocui.KeyEsc))
	assert.False(t, gui.State.vimPendingFold)
	assert.Len(t, dispatched, count)

	// Pasted characters neither start nor finish a fold sequence.
	press(z)
	gui.g.IsPasting = true
	a := gocui.NewKeyRune('a')
	press(a)
	press(z)
	assert.False(t, gui.State.vimPendingFold)
	assert.Equal(t, []gocui.Key{a, z}, dispatched[count:])
}
