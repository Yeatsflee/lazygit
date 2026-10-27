package commit

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimFolding = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim folds navigate commit directories without a root row and leave top-level files unchanged",
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.DisableCustomPatch = true
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("dir/nested/file-one", "one\n")
		shell.CreateFileAndAdd("other/file-two", "two\n")
		shell.CreateFileAndAdd("top-file", "top\n")
		shell.Commit("add files")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().Focus().PressEnter()
		t.Views().CommitFiles().
			IsFocused().
			NavigateToLine(Contains("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"c"}).
			SelectedLine(Equals("▶ dir/nested")).
			Content(DoesNotContain("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"c"}).
			SelectedLine(Equals("▶ dir/nested")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			SelectedLine(Equals("▼ dir/nested")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			SelectedLine(Equals("▼ dir/nested")).
			NavigateToLine(Contains("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"a"}).
			SelectedLine(Equals("▶ dir/nested")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"a"}).
			SelectedLine(Equals("▼ dir/nested"))

		// Without a root row, a top-level file has no visible parent to fold.
		t.Views().CommitFiles().
			NavigateToLine(Contains("top-file")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"a"}).
			SelectedLine(Contains("top-file")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"c"}).
			SelectedLine(Contains("top-file")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			SelectedLine(Contains("top-file")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"M"}).
			Lines(
				Equals("▶ dir/nested"),
				Equals("▶ other"),
				Contains("top-file").IsSelected(),
			).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"R"}).
			Lines(
				Equals("▼ dir/nested"),
				Contains("file-one"),
				Equals("▼ other"),
				Contains("file-two"),
				Contains("top-file").IsSelected(),
			)

		// A plain o still opens the selected file, while zo stays in the tree.
		t.Views().CommitFiles().
			NavigateToLine(Contains("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			IsFocused().
			SelectedLine(Contains("file-one")).
			Press(config.Keybinding{"o"})
		t.Views().Main().IsFocused().Content(Contains("+one"))
	},
})
