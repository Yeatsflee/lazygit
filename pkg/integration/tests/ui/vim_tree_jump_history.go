package ui

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var VimTreeJumpHistory = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Vim jumps reopen the same file when its directory is collapsed and new files change the tree indices",
	SetupConfig: func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.EmptyCommit("initial")
		shell.CreateFile("first/file-one", "first content\n")
		shell.CreateFile("last/nested/file-two", "second content\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			Focus().
			NavigateToLine(Contains("file-two")).
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Content(Contains("+second content"))

		t.Views().Files().Focus()
		t.Shell().CreateFile("earlier/new-file", "new content\n")
		t.Views().Files().
			Press(keys.Universal.Refresh).
			ContainsLines(Contains("new-file")).
			Press(keys.Files.CollapseAll).
			Lines(Equals("▶ /").IsSelected()).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().Main().
			IsFocused().
			Content(Contains("+second content").DoesNotContain("+first content")).
			Press(config.Keybinding{"<ctrl+o>"})

		t.Views().Files().
			IsFocused().
			SelectedLine(Contains("file-two")).
			ContainsLines(
				Contains("▼ last/nested"),
				Contains("file-two").IsSelected(),
			).
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Content(Contains("+second content"))

		// A deleted destination must not open another file's diff in its place.
		t.Views().Files().Focus()
		t.Shell().DeleteFile("last/nested/file-two")
		t.Views().Files().
			Press(keys.Universal.Refresh).
			Content(DoesNotContain("file-two")).
			Press(config.Keybinding{"<ctrl+o>"}).
			IsFocused().
			SelectedLine(DoesNotContain("file-two"))
	},
})
