package commit

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ParentNavigation = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Return through commit files and their parent directories with p",
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.DisableCustomPatch = true
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("dir/nested/file-one", "content\n")
		shell.CreateFileAndAdd("other/file-two", "content\n")
		shell.Commit("add files")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().
			Focus().
			Press(config.Keybinding{"p"}).
			IsFocused().
			PressEnter()

		t.Views().CommitFiles().
			IsFocused().
			NavigateToLine(Contains("file-one")).
			PressEnter()

		t.Views().Main().
			IsFocused().
			Press(config.Keybinding{"p"})

		t.Views().CommitFiles().
			IsFocused().
			SelectedLine(Contains("file-one")).
			Press(config.Keybinding{"p"}).
			SelectedLine(Equals("▼ dir/nested")).
			Press(config.Keybinding{"p"})

		t.Views().Commits().
			IsFocused().
			SelectedLine(Contains("add files")).
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Press(config.Keybinding{"p"})

		t.Views().Commits().
			IsFocused().
			PressEnter()

		// With no directory rows, the parent is the commit list.
		t.Views().CommitFiles().
			IsFocused().
			Press(keys.Files.ToggleTreeView).
			Press(config.Keybinding{"p"})

		t.Views().Commits().IsFocused()
	},
})
