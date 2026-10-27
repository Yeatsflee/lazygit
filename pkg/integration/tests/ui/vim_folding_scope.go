package ui

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimFoldingScope = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim fold prefixes cancel on unrelated keys and context changes while other views and flat lists keep undo",
	SetupConfig: func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("dir/file-one", "one\n")
		shell.Commit("one")
		shell.CreateFileAndAdd("other/file-two", "two\n")
		shell.Commit("two")
		shell.UpdateFile("dir/file-one", "changed one\n")
		shell.UpdateFile("other/file-two", "changed two\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		cancelUndo := func() {
			t.ExpectPopup().Confirmation().
				Title(Equals("Undo")).
				Content(Contains("soft reset")).
				Cancel()
		}

		// zz retains the original undo confirmation without performing the reset.
		t.Views().Files().IsFocused().Press(config.Keybinding{"z"}).IsFocused().
			Press(config.Keybinding{"z"}).Tap(cancelUndo).IsFocused()
		t.Views().Commits().ContainsLines(Contains("two"), Contains("one"))

		// An unbound key clears the prefix; the next o opens the diff normally.
		t.Views().Files().NavigateToLine(Contains("file-one")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"<f12>"}).
			Press(config.Keybinding{"o"})
		t.Views().Staging().IsFocused().Press(config.Keybinding{"p"})

		// Escape cancels the prefix without leaving the file tree.
		t.Views().Files().IsFocused().Press(config.Keybinding{"z"}).PressEscape().
			IsFocused().SelectedLine(Contains("file-one"))

		// Leaving and re-entering the tree also clears an incomplete prefix.
		t.Views().Files().IsFocused().Press(config.Keybinding{"z"})
		t.Views().Commits().Focus()
		t.Views().Files().Focus().Press(config.Keybinding{"o"})
		t.Views().Staging().IsFocused().Press(config.Keybinding{"p"})

		// g and z are mutually exclusive prefixes rather than independently pending keys.
		t.Views().Files().NavigateToLine(Contains("file-two")).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"R"}).
			Press(config.Keybinding{"g"}).SelectedLine(Contains("file-two")).
			Press(config.Keybinding{"g"}).SelectedLine(Equals("▼ /")).
			NavigateToLine(Contains("file-two")).
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"g"}).
			SelectedLine(Contains("file-two")).
			Press(config.Keybinding{"g"}).SelectedLine(Equals("▼ /"))

		// Flat files, commits, branches, and the main diff keep their single-z undo.
		t.Views().Files().Press(keys.Files.ToggleTreeView).
			Press(config.Keybinding{"z"}).Tap(cancelUndo).
			Press(keys.Files.ToggleTreeView)
		t.Views().Commits().Focus().Press(config.Keybinding{"z"}).Tap(cancelUndo)
		t.Views().Commits().PressEnter()
		t.Views().CommitFiles().IsFocused().
			Press(config.Keybinding{"z"}).Press(config.Keybinding{"z"}).Tap(cancelUndo).
			Press(keys.Files.ToggleTreeView).
			Press(config.Keybinding{"z"}).Tap(cancelUndo).
			Press(keys.Files.ToggleTreeView)
		t.Views().Branches().Focus().Press(config.Keybinding{"z"}).Tap(cancelUndo)
		t.Views().Files().Focus().NavigateToLine(Contains("file-one")).
			Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().Press(config.Keybinding{"z"}).Tap(cancelUndo)

		// Text input receives both characters of every fold command unchanged.
		t.Views().Branches().Focus().Press(keys.Universal.New)
		t.ExpectPopup().Prompt().Title(Contains("New branch name")).Type("zazczozMzR")
		t.Views().Prompt().Content(Equals("zazczozMzR"))
		t.ExpectPopup().Prompt().Title(Contains("New branch name")).Cancel()
	},
})
