package ui

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimMovement = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim movement keys navigate working files, commits, and commit files without treating a single g as a command",
	Width:       100,
	Height:      30,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateNCommits(20)
		for i := 1; i <= 40; i++ {
			shell.CreateFileAndAdd(fmt.Sprintf("browse%02d.txt", i), "original\n")
		}
		shell.Commit("browse files")
		for i := 1; i <= 40; i++ {
			shell.UpdateFile(fmt.Sprintf("browse%02d.txt", i), "changed\n")
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			Focus().
			SelectedLine(Contains("browse01.txt")).
			Press(config.Keybinding{"<ctrl+d>"}).
			SelectedLine(DoesNotContain("browse01.txt")).
			Press(config.Keybinding{"<ctrl+u>"}).
			SelectedLine(Contains("browse01.txt")).
			Press(config.Keybinding{"G"}).
			SelectedLine(Contains("browse40.txt")).
			Press(config.Keybinding{"g"}).
			IsFocused().
			SelectedLine(Contains("browse40.txt")).
			Press(config.Keybinding{"<f12>"}).
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("browse40.txt")).
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("browse01.txt"))

		t.Views().Commits().
			Focus().
			SelectedLine(Contains("browse files")).
			Press(config.Keybinding{"<ctrl+d>"}).
			SelectedLine(DoesNotContain("browse files")).
			Press(config.Keybinding{"<ctrl+u>"}).
			SelectedLine(Contains("browse files")).
			Press(config.Keybinding{"G"}).
			SelectedLine(Contains("commit 01"))

		// A pending first g must not survive a context change.
		t.Views().Files().Focus().Press(config.Keybinding{"g"})
		t.Views().Commits().
			Focus().
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("commit 01")).
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("browse files")).
			PressEnter()

		t.Views().CommitFiles().
			IsFocused().
			SelectedLine(Contains("browse01.txt")).
			Press(config.Keybinding{"<ctrl+d>"}).
			SelectedLine(DoesNotContain("browse01.txt")).
			Press(config.Keybinding{"<ctrl+u>"}).
			SelectedLine(Contains("browse01.txt")).
			Press(config.Keybinding{"G"}).
			SelectedLine(Contains("browse40.txt")).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("browse01.txt"))

		// Outside the browsing contexts, g still opens the original reset menu.
		t.Views().Branches().Focus().Press(config.Keybinding{"g"})
		t.ExpectPopup().Menu().Title(Contains("Reset to")).Cancel()

		// Editing a prompt must keep both g keystrokes as ordinary text.
		t.Views().Branches().Press(keys.Universal.New)
		t.ExpectPopup().Prompt().Title(Contains("New branch name")).Type("gg")
		t.Views().Prompt().Content(Equals("gg"))
		t.ExpectPopup().Prompt().Title(Contains("New branch name")).Cancel()
	},
})
