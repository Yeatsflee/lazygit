package ui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimJumpHistory = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim jump history restores list selections and diff scrolling without recording ordinary movement",
	Width:       100,
	Height:      14,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ScreenMode = "full"
		config.GetUserConfig().Gui.ShowCommandLog = false
		config.GetUserConfig().Gui.ShowBottomLine = false
		config.GetUserConfig().Gui.ScrollPastBottom = false
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
		config.GetUserConfig().Gui.UseHunkModeInStagingView = false
	},
	SetupRepo: func(shell *Shell) {
		shell.EmptyCommit("initial")
		var content strings.Builder
		for i := 1; i <= 80; i++ {
			fmt.Fprintf(&content, "line%02d\n", i)
		}
		for i := 1; i <= 40; i++ {
			shell.CreateFile(fmt.Sprintf("file%02d.txt", i), content.String())
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		lines := func(first, last int) []*TextMatcher {
			result := []*TextMatcher{}
			for i := first; i <= last; i++ {
				result = append(result, Equals(fmt.Sprintf("+line%02d", i)))
			}
			return result
		}

		t.Views().Files().
			Focus().
			NavigateToLine(Contains("file03.txt")).
			Press(config.Keybinding{"G"}).
			SelectedLine(Contains("file40.txt")).
			SelectPreviousItem().
			Press(config.Keybinding{"<ctrl+d>"}).
			Press(config.Keybinding{"<ctrl+u>"}).
			NavigateToLine(Contains("file39.txt")).
			Press(config.Keybinding{"<ctrl+o>"}).
			SelectedLine(Contains("file03.txt")).
			Press(config.Keybinding{"<ctrl+i>"}).
			SelectedLine(Contains("file39.txt")).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("file01.txt")).
			Press(config.Keybinding{"<ctrl+o>"}).
			SelectedLine(Contains("file39.txt")).
			Press(config.Keybinding{"<ctrl+i>"}).
			SelectedLine(Contains("file01.txt")).
			Press(config.Keybinding{"<ctrl+o>"}).
			SelectedLine(Contains("file39.txt")).
			Press(config.Keybinding{"o"})

		// Entering a file after going back discards the old forward destination.
		t.Views().Staging().
			IsFocused().
			SelectedLine(Equals("+line01")).
			Press(config.Keybinding{"<ctrl+i>"}).
			IsFocused().
			SelectedLine(Equals("+line01")).
			Press(config.Keybinding{"<ctrl+d>"}).
			SelectedLine(Equals("+line07")).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("file39.txt")).
			Press(config.Keybinding{"<ctrl+i>"})

		t.Views().Staging().
			IsFocused().
			SelectedLine(Equals("+line07")).
			VisibleLines(lines(1, 12)...).
			Press(config.Keybinding{"p"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("file39.txt")).
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			TopLines(Contains("diff --git")).
			Press(config.Keybinding{"<ctrl+d>"}).
			Press(config.Keybinding{"<ctrl+d>"}).
			VisibleLines(lines(7, 18)...).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("file39.txt")).
			Press(config.Keybinding{"<ctrl+i>"})

		t.Views().Main().
			IsFocused().
			VisibleLines(lines(7, 18)...).
			Press(config.Keybinding{"G"}).
			VisibleLines(lines(69, 80)...).
			Press(config.Keybinding{"<ctrl+o>"}).
			VisibleLines(lines(7, 18)...).
			Press(config.Keybinding{"<ctrl+i>"}).
			VisibleLines(lines(69, 80)...)

		// G at the existing bottom must preserve the destination ahead of us.
		t.Views().Main().
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"<ctrl+o>"}).
			VisibleLines(lines(69, 80)...).
			Press(config.Keybinding{"G"}).
			Press(config.Keybinding{"<ctrl+i>"}).
			VisibleLines(
				Contains("diff --git"),
				Contains("new file mode"),
				Contains("index"),
				Equals("--- /dev/null"),
				Equals("+++ b/file39.txt"),
				Contains("@@"),
				Equals("+line01"),
				Equals("+line02"),
				Equals("+line03"),
				Equals("+line04"),
				Equals("+line05"),
				Equals("+line06"),
			).
			PressEscape()

		// Neither gg at the first item nor p at the root creates a new jump.
		t.Views().Files().
			IsFocused().
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"G"}).
			Press(config.Keybinding{"<ctrl+o>"}).
			SelectedLine(Contains("file01.txt")).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"g"}).
			Press(config.Keybinding{"p"}).
			Press(config.Keybinding{"<ctrl+i>"}).
			SelectedLine(Contains("file40.txt"))
	},
})
