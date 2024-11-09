package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) initProjectSelect() {
	cTui.projectSelect = tview.NewInputField().SetText(cTui.config.Project).SetFieldWidth(30)
	cTui.projectSelect.SetLabelStyle(
		tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
	)
	cTui.projectSelect.SetFieldBackgroundColor(tcell.ColorDefault)

	cTui.projectSelect.SetLabel("Project: ").SetDoneFunc(func(key tcell.Key) {
		cTui.config.Project = cTui.projectSelect.GetText()
		pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, "")
		cTui.pipelines.populateTable(pipelines, nextPageToken)
		cTui.pipelines.table.ScrollToBeginning()
		cTui.app.SetFocus(cTui.pipelines.table)
	})

	cTui.projectSelect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.app.Stop()
		}

		return event
	})
}
