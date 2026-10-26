package gui

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func TestVimTopRequiresConsecutiveKeys(t *testing.T) {
	files := context.NewSimpleContext(context.NewBaseContext(context.NewBaseContextOpts{
		Key: context.FILES_CONTEXT_KEY, Kind: types.SIDE_CONTEXT,
	}))
	gui := &Gui{
		g: &gocui.Gui{},
		State: &GuiRepoState{
			ContextMgr: &ContextMgr{ContextStack: []types.Context{files}},
		},
	}
	dispatched := []gocui.Key{}
	press := func(key gocui.Key) {
		assert.NoError(t, gui.handleVimNavigationKey(key, func() error {
			dispatched = append(dispatched, key)
			return nil
		}))
	}
	g := gocui.NewKeyRune('g')
	other := gocui.NewKeyRune('z')

	press(g)
	assert.Empty(t, dispatched)
	press(other)
	press(g)
	assert.Equal(t, []gocui.Key{other}, dispatched)
	press(g)
	assert.Equal(t, []gocui.Key{other, g}, dispatched)

	// Paste does not start or complete a navigation sequence.
	gui.g.IsPasting = true
	press(g)
	assert.False(t, gui.State.vimPendingTop)
	gui.g.IsPasting = false
	press(g)
	assert.True(t, gui.State.vimPendingTop)
}
