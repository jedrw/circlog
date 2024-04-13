package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) initBranchSelect() {
	cTui.branchSelect = tview.NewInputField().SetText(cTui.config.Branch).SetFieldWidth(30)
	cTui.branchSelect.SetLabelColor(tcell.ColorDefault)

	cTui.branchSelect.SetLabel("Branch: ").SetDoneFunc(func(key tcell.Key) {
		cTui.config.Branch = cTui.branchSelect.GetText()
		cTui.pipelines.update <- true

		cTui.app.SetFocus(cTui.pipelines.table)
	})

	cTui.branchSelect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.app.Stop()
		}

		return event
	})
}
