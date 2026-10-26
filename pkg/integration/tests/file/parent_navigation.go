package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ParentNavigation = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Return from a working tree diff and select parent directories with p",
	SetupConfig: func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("dir/nested/file-one", "content\n")
		shell.CreateFile("other/file-two", "content\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			NavigateToLine(Contains("file-one")).
			PressEnter()

		t.Views().Staging().
			IsFocused().
			Press(config.Keybinding{"p"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("file-one")).
			Press(config.Keybinding{"p"}).
			SelectedLine(Equals("  ▼ dir/nested")).
			Press(config.Keybinding{"p"}).
			SelectedLine(Equals("▼ /")).
			Press(config.Keybinding{"p"}).
			IsFocused().
			SelectedLine(Equals("▼ /"))

		// A flat file list has no directory rows to select.
		t.Views().Files().
			NavigateToLine(Contains("file-one")).
			Press(keys.Files.ToggleTreeView).
			Press(config.Keybinding{"p"}).
			IsFocused().
			SelectedLine(Contains("dir/nested/file-one")).
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Press(config.Keybinding{"p"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("dir/nested/file-one"))
	},
})
