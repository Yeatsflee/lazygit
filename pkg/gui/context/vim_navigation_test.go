package context

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func TestSupportsVimNavigation(t *testing.T) {
	newContext := func(key types.ContextKey, parent types.Context) types.Context {
		context := NewSimpleContext(NewBaseContext(NewBaseContextOpts{Key: key}))
		context.SetParentContext(parent)
		return context
	}
	files := newContext(FILES_CONTEXT_KEY, nil)
	commits := newContext(LOCAL_COMMITS_CONTEXT_KEY, nil)
	branches := newContext(LOCAL_BRANCHES_CONTEXT_KEY, nil)
	commitFiles := newContext(COMMIT_FILES_CONTEXT_KEY, commits)
	stashFiles := newContext(COMMIT_FILES_CONTEXT_KEY, newContext(STASH_CONTEXT_KEY, nil))
	main := newContext(NORMAL_MAIN_CONTEXT_KEY, nil)
	staging := newContext(STAGING_MAIN_CONTEXT_KEY, nil)

	for _, test := range []struct {
		name    string
		current types.Context
		side    types.Context
		want    bool
	}{
		{name: "working tree", current: files, side: files, want: true},
		{name: "local commits", current: commits, side: commits, want: true},
		{name: "commit files", current: commitFiles, side: commitFiles, want: true},
		{name: "working tree diff", current: staging, side: files, want: true},
		{name: "commit diff", current: main, side: commitFiles, want: true},
		{name: "branch diff", current: main, side: branches},
		{name: "stash files", current: stashFiles, side: stashFiles},
		{name: "stash diff", current: main, side: stashFiles},
		{name: "popup over files", current: newContext(PROMPT_CONTEXT_KEY, nil), side: files},
		{name: "main without source", current: main},
		{name: "no current context", side: files},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, SupportsVimNavigation(test.current, test.side))
		})
	}
}
