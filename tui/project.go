package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) initProjectSelect() {
	cTui.projectSelect = tview.NewInputField().SetText(cTui.config.Project).SetFieldWidth(30)
	cTui.projectSelect.SetLabelColor(tcell.ColorDefault)

	cTui.projectSelect.SetLabel("Project: ").SetDoneFunc(func(key tcell.Key) {
		cTui.config.Project = cTui.projectSelect.GetText()
		cTui.pipelines.update <- true
		cTui.app.SetFocus(cTui.pipelines.table)
	})

	cTui.projectSelect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.app.Stop()
		}

		return event
	})
}
