package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimFolding = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim fold commands toggle directories, fold the nearest visible parent, and fold the complete working file tree",
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowRootItemInFileTree = true
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("dir/nested/file-one", "one\n")
		shell.CreateFile("dir/file-two", "two\n")
		shell.CreateFile("other/file-three", "three\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			NavigateToLine(Equals("  ▼ dir")).
			Press(config.Keybinding{"z"}).
			IsFocused().
			SelectedLine(Equals("  ▼ dir")).
			Press(config.Keybinding{"a"}).
			Lines(
				Equals("▼ /"),
				Equals("  ▶ dir").IsSelected(),
				Equals("  ▼ other"),
				Equals("    ?? file-three"),
			).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"c"}).
			SelectedLine(Equals("  ▶ dir")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			SelectedLine(Equals("  ▼ dir")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			SelectedLine(Equals("  ▼ dir"))

		// Opening an already visible file's parent must preserve the file selection.
		t.Views().Files().
			NavigateToLine(Contains("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			IsFocused().
			SelectedLine(Contains("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"c"}).
			SelectedLine(Equals("    ▶ nested")).
			Content(DoesNotContain("file-one")).
			ContainsLines(Contains("file-two")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"a"}).
			NavigateToLine(Contains("file-two")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"a"}).
			SelectedLine(Equals("  ▶ dir"))

		// zM also folds nested directories hidden below a previously folded parent.
		t.Views().Files().
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"M"}).
			Lines(Equals("▶ /").IsSelected()).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			Lines(
				Equals("▼ /").IsSelected(),
				Equals("  ▶ dir"),
				Equals("  ▶ other"),
			).
			NavigateToLine(Equals("  ▶ dir")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"o"}).
			Lines(
				Equals("▼ /"),
				Equals("  ▼ dir").IsSelected(),
				Equals("    ?? file-two"),
				Equals("    ▶ nested"),
				Equals("  ▶ other"),
			).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"R"}).
			Lines(
				Equals("▼ /"),
				Equals("  ▼ dir").IsSelected(),
				Equals("    ?? file-two"),
				Equals("    ▼ nested"),
				Equals("      ?? file-one"),
				Equals("  ▼ other"),
				Equals("    ?? file-three"),
			)
	},
})
