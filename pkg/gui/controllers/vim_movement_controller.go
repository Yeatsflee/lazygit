package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type VimMovementController struct {
	baseController
	c                 *ControllerCommon
	context           types.Context
	prepareGotoBottom func() func(func())
}

func NewVimMovementController(c *ControllerCommon, context types.Context, prepareGotoBottom func() func(func())) *VimMovementController {
	return &VimMovementController{c: c, context: context, prepareGotoBottom: prepareGotoBottom}
}

func (self *VimMovementController) Context() types.Context {
	return self.context
}

func (self *VimMovementController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Tag:               "navigation",
			Keys:              []gocui.Key{gocui.NewKeyStrMod("u", gocui.ModCtrl)},
			Handler:           func() error { return self.halfPage(-1) },
			Description:       "Scroll up half a page",
			GetDisabledReason: self.disabledReason,
		},
		{
			Tag:               "navigation",
			Keys:              []gocui.Key{gocui.NewKeyStrMod("d", gocui.ModCtrl)},
			Handler:           func() error { return self.halfPage(1) },
			Description:       "Scroll down half a page",
			GetDisabledReason: self.disabledReason,
		},
		{
			Tag:               "navigation",
			Keys:              []gocui.Key{gocui.NewKeyRune('g')},
			Handler:           self.gotoTop,
			Description:       "Go to top (gg)",
			GetDisabledReason: self.disabledReason,
		},
		{
			Tag:               "navigation",
			Keys:              []gocui.Key{gocui.NewKeyRune('G')},
			Handler:           self.gotoBottom,
			Description:       self.c.Tr.GotoBottom,
			GetDisabledReason: self.disabledReason,
		},
	}
}

func (self *VimMovementController) disabledReason() *types.DisabledReason {
	if !context.SupportsVimNavigation(self.c.Context().Current(), self.c.Context().CurrentSide()) {
		return &types.DisabledReason{AllowFurtherDispatching: true}
	}
	return nil
}

func (self *VimMovementController) halfPage(direction int) error {
	delta := direction * max(1, self.context.GetView().InnerHeight()/2)
	switch ctx := self.context.(type) {
	case types.IListContext:
		list := ctx.GetList()
		if list.Len() == 0 {
			return nil
		}
		before := ctx.ModelIndexToViewIndex(list.GetSelectedLineIdx())
		after := ctx.ViewIndexToModelIndex(before + delta)
		list.MoveSelectedLine(after - list.GetSelectedLineIdx())
		self.scroll(ctx.ModelIndexToViewIndex(list.GetSelectedLineIdx()) - before)
		ctx.SetNeedRerenderVisibleLines()
		ctx.HandleFocus(types.OnFocusOpts{ScrollSelectionIntoView: true})
	case types.IPatchExplorerContext:
		controller := &PatchExplorerController{c: self.c, context: ctx}
		return controller.withRenderAndFocus(func() error {
			state := ctx.GetState()
			before := state.GetSelectedViewLineIdx()
			state.AdjustSelectedLineIdx(delta)
			self.scroll(state.GetSelectedViewLineIdx() - before)
			return nil
		})()
	default:
		controller := &ViewSelectionController{c: self.c, context: ctx}
		controller.handleLineChange(delta)
	}
	return nil
}

func (self *VimMovementController) scroll(delta int) {
	if delta < 0 {
		self.context.GetViewTrait().ScrollUp(-delta)
	} else {
		self.context.GetViewTrait().ScrollDown(delta)
	}
}

func (self *VimMovementController) gotoTop() error {
	switch ctx := self.context.(type) {
	case types.IListContext:
		controller := &ListController{c: self.c, context: ctx}
		return controller.HandleGotoTop()
	case types.IPatchExplorerContext:
		controller := &PatchExplorerController{c: self.c, context: ctx}
		return controller.withRenderAndFocus(controller.HandleGotoTop)()
	default:
		controller := &ViewSelectionController{c: self.c, context: ctx}
		return controller.handleGotoTop()
	}
}

func (self *VimMovementController) gotoBottom() error {
	switch ctx := self.context.(type) {
	case types.IListContext:
		controller := &ListController{c: self.c, context: ctx}
		return controller.HandleGotoBottom()
	case types.IPatchExplorerContext:
		controller := &PatchExplorerController{c: self.c, context: ctx}
		return controller.withRenderAndFocus(controller.HandleGotoBottom)()
	default:
		controller := &ViewSelectionController{c: self.c, context: ctx, prepareGotoBottom: self.prepareGotoBottom}
		return controller.handleGotoBottom()
	}
}
