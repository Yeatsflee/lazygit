package context

import "github.com/jesseduffield/lazygit/pkg/gui/types"

// SupportsVimNavigation confines the extra navigation keys to the working tree,
// local commits, and the main views reached from those lists.
func SupportsVimNavigation(current types.Context, side types.Context) bool {
	if current == nil {
		return false
	}
	switch current.GetKey() {
	case FILES_CONTEXT_KEY, LOCAL_COMMITS_CONTEXT_KEY, COMMIT_FILES_CONTEXT_KEY:
		return supportsVimSideNavigation(current)
	case NORMAL_MAIN_CONTEXT_KEY, NORMAL_SECONDARY_CONTEXT_KEY,
		STAGING_MAIN_CONTEXT_KEY, STAGING_SECONDARY_CONTEXT_KEY,
		PATCH_BUILDING_MAIN_CONTEXT_KEY:
		return supportsVimSideNavigation(side)
	}
	return false
}

func supportsVimSideNavigation(side types.Context) bool {
	if side == nil {
		return false
	}
	switch side.GetKey() {
	case FILES_CONTEXT_KEY, LOCAL_COMMITS_CONTEXT_KEY:
		return true
	case COMMIT_FILES_CONTEXT_KEY:
		parent := side.GetParentContext()
		return parent != nil && parent.GetKey() == LOCAL_COMMITS_CONTEXT_KEY
	}
	return false
}
