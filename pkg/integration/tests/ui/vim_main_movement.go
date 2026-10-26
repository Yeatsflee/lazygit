package ui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimMainMovement = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim half-page and boundary movement scroll long diffs and move the staging selection",
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
		shell.CreateFile("long.txt", content.String())
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		lines := func(first, last int) []*TextMatcher {
			result := []*TextMatcher{}
			for i := first; i <= last; i++ {
				result = append(result, Equals(fmt.Sprintf("+line%02d", i)))
			}
			return result
		}

		t.Views().Files().Focus().Press(keys.Universal.FocusMainView)
		t.Views().Main().
			IsFocused().
			TopLines(Contains("diff --git")).
			Press(config.Keybinding{"<ctrl+d>"}).
			VisibleLines(lines(1, 12)...).
			Press(config.Keybinding{"<ctrl+d>"}).
			VisibleLines(lines(7, 18)...).
			Press(config.Keybinding{"<ctrl+u>"}).
			VisibleLines(lines(1, 12)...).
			Press(config.Keybinding{"G"}).
			VisibleLines(lines(69, 80)...).
			Press(config.Keybinding{"g"}).
			VisibleLines(lines(69, 80)...).
			Press(config.Keybinding{"g"}).
			VisibleLines(
				Contains("diff --git"),
				Contains("new file mode"),
				Contains("index"),
				Equals("--- /dev/null"),
				Equals("+++ b/long.txt"),
				Contains("@@"),
				Equals("+line01"),
				Equals("+line02"),
				Equals("+line03"),
				Equals("+line04"),
				Equals("+line05"),
				Equals("+line06"),
			).
			PressEscape()

		t.Views().Files().IsFocused().PressEnter()
		t.Views().Staging().
			IsFocused().
			SelectedLine(Equals("+line01")).
			Press(config.Keybinding{"<ctrl+d>"}).
			SelectedLine(Equals("+line07")).
			VisibleLines(lines(1, 12)...).
			Press(config.Keybinding{"<ctrl+u>"}).
			SelectedLine(Equals("+line01")).
			Press(config.Keybinding{"G"}).
			SelectedLine(Equals("+line80")).
			Press(config.Keybinding{"g"}).
			SelectedLine(Equals("+line80")).
			Press(config.Keybinding{"g"}).
			SelectedLine(Contains("diff --git"))
	},
})
