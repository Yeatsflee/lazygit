package controllers

import (
	"strings"
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

type vimListContext struct {
	types.IListContext
	view        *gocui.View
	list        types.IList
	itemHeight  int
	needsRender bool
}

func (self *vimListContext) GetView() *gocui.View { return self.view }

func (self *vimListContext) GetViewTrait() types.IViewTrait { return context.NewViewTrait(self.view) }

func (self *vimListContext) GetList() types.IList { return self.list }

func (self *vimListContext) ModelIndexToViewIndex(index int) int { return index * self.itemHeight }

func (self *vimListContext) ViewIndexToModelIndex(index int) int { return index / self.itemHeight }

func (self *vimListContext) SetNeedRerenderVisibleLines() { self.needsRender = true }

func (self *vimListContext) HandleFocus(opts types.OnFocusOpts) {
	self.view.FocusPoint(0, self.ModelIndexToViewIndex(self.list.GetSelectedLineIdx()), opts.ScrollSelectionIntoView)
}

func TestVimHalfPageMovesSelectionAndViewportTogether(t *testing.T) {
	for _, scenario := range []struct {
		name           string
		height         int
		itemHeight     int
		initialIndex   int
		initialOrigin  int
		direction      int
		expectedIndex  int
		expectedOrigin int
	}{
		{name: "down", height: 10, itemHeight: 1, initialIndex: 12, initialOrigin: 10, direction: 1, expectedIndex: 17, expectedOrigin: 15},
		{name: "up", height: 10, itemHeight: 1, initialIndex: 17, initialOrigin: 15, direction: -1, expectedIndex: 12, expectedOrigin: 10},
		{name: "top boundary", height: 10, itemHeight: 1, initialIndex: 2, direction: -1, expectedIndex: 0, expectedOrigin: 0},
		{name: "bottom boundary", height: 10, itemHeight: 1, initialIndex: 98, initialOrigin: 90, direction: 1, expectedIndex: 99, expectedOrigin: 90},
		{name: "wrapped items", height: 12, itemHeight: 2, initialIndex: 12, initialOrigin: 20, direction: 1, expectedIndex: 15, expectedOrigin: 26},
		{name: "single visible line", height: 1, itemHeight: 1, initialIndex: 10, initialOrigin: 10, direction: 1, expectedIndex: 11, expectedOrigin: 11},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			items := make([]*models.Commit, 100)
			list := context.NewListViewModel(func() []*models.Commit { return items })
			list.SetSelection(scenario.initialIndex)
			view := gocui.NewView("commits", 0, 0, 30, scenario.height+1, gocui.OutputNormal)
			view.SetContent(strings.TrimSuffix(strings.Repeat("item\n", len(items)*scenario.itemHeight), "\n"))
			view.SetOriginY(scenario.initialOrigin)
			view.SetCursorY(scenario.initialIndex*scenario.itemHeight - scenario.initialOrigin)
			ctx := &vimListContext{view: view, list: list, itemHeight: scenario.itemHeight}

			assert.NoError(t, NewVimMovementController(nil, ctx).halfPage(scenario.direction))

			assert.Equal(t, scenario.expectedIndex, list.GetSelectedLineIdx())
			assert.Equal(t, scenario.expectedOrigin, view.OriginY())
			assert.True(t, ctx.needsRender)
		})
	}
}
