package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) initBranchSelect() {
	cTui.branchSelect = tview.NewInputField().SetText(cTui.config.Branch).SetFieldWidth(30)
	cTui.branchSelect.SetLabelStyle(
		tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
	)
	cTui.branchSelect.SetFieldBackgroundColor(tcell.ColorDefault)

	cTui.branchSelect.SetLabel("Branch: ").SetDoneFunc(func(key tcell.Key) {
		cTui.config.Branch = cTui.branchSelect.GetText()
		pipelines, nextPageToken, _ := circleci.GetPipelines(cTui.config, 1, "")
		cTui.populatePipelinesTable(pipelines, nextPageToken)
		cTui.pipelinesPane.ScrollToBeginning()
		cTui.app.SetFocus(cTui.pipelinesPane)
	})

	cTui.branchSelect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.app.Stop()
		}

		return event
	})
}
