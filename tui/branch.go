package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newBranchSelect(config *config.CirclogConfig) *tview.InputField {
	branchSelect := tview.NewInputField().SetText(config.Branch).SetFieldWidth(30)
	branchSelect.SetLabelColor(tcell.ColorDefault)

	branchSelect.SetLabel("Branch: ").SetDoneFunc(func(key tcell.Key) {
		config.Branch = branchSelect.GetText()
		clearAll()
		updatePipelinesTable(config, pipelinesTable)
		app.SetFocus(pipelinesTable)
	})

	branchSelect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
		}

		return event
	})

	return branchSelect
}