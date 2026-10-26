package ui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimCommitJumpHistory = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim jumps restore the original commit and diff after browsing files in a different commit",
	Width:       100,
	Height:      14,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ScreenMode = "full"
		config.GetUserConfig().Gui.ShowCommandLog = false
		config.GetUserConfig().Gui.ShowBottomLine = false
		config.GetUserConfig().Gui.ScrollPastBottom = false
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
		config.GetUserConfig().Gui.DisableCustomPatch = true
	},
	SetupRepo: func(shell *Shell) {
		var oldContent, newContent strings.Builder
		for i := 1; i <= 80; i++ {
			fmt.Fprintf(&oldContent, "old%02d\n", i)
			fmt.Fprintf(&newContent, "new%02d\n", i)
		}
		shell.CreateFileAndAdd("history.txt", oldContent.String())
		shell.Commit("older content")
		shell.UpdateFileAndAdd("history.txt", newContent.String())
		shell.Commit("newer content")
		shell.CreateFile("working.txt", "working content\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		bottomLines := []*TextMatcher{}
		for i := 69; i <= 80; i++ {
			bottomLines = append(bottomLines, Equals(fmt.Sprintf("+new%02d", i)))
		}

		t.Views().Files().Focus().SelectedLine(Contains("working.txt"))
		t.Views().Commits().
			Focus().
			SelectedLine(Contains("newer content")).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("working.txt")).
			Press(config.Keybinding{"<ctrl+i>"})

		t.Views().Commits().
			IsFocused().
			SelectedLine(Contains("newer content")).
			PressEnter()

		t.Views().CommitFiles().
			IsFocused().
			SelectedLine(Contains("history.txt")).
			PressEnter()

		t.Views().Main().
			IsFocused().
			Content(Contains("+new01")).
			Press(config.Keybinding{"G"}).
			VisibleLines(bottomLines...)

		t.Views().Commits().
			Focus().
			SelectNextItem().
			SelectedLine(Contains("older content")).
			PressEnter()

		t.Views().CommitFiles().
			IsFocused().
			SelectedLine(Contains("history.txt")).
			PressEnter()

		t.Views().Main().
			IsFocused().
			Content(Contains("+old01").DoesNotContain("+new01")).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().CommitFiles().
			IsFocused().
			SelectedLine(Contains("history.txt")).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().Commits().
			IsFocused().
			SelectedLine(Contains("older content")).
			Press(config.Keybinding{"<ctrl+o>"})

		// The mutable commit-files view now refers to the older commit. Returning
		// to its previous diff must restore the newer commit before rendering it.
		t.Views().Main().
			IsFocused().
			Content(Contains("+new01").DoesNotContain("+old01")).
			VisibleLines(bottomLines...).
			Press(config.Keybinding{"<ctrl+i>"})

		t.Views().Commits().
			IsFocused().
			SelectedLine(Contains("older content")).
			Press(config.Keybinding{"<ctrl+i>"})

		t.Views().CommitFiles().
			IsFocused().
			SelectedLine(Contains("history.txt")).
			Press(config.Keybinding{"<ctrl+i>"})

		t.Views().Main().
			IsFocused().
			Content(Contains("+old01").DoesNotContain("+new01"))
	},
})
