package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newProjectSelect(config *config.CirclogConfig) *tview.InputField {
	projectSelect := tview.NewInputField().SetText(config.Project).SetFieldWidth(30)
	projectSelect.SetLabelColor(tcell.ColorDefault)
	
	projectSelect.SetLabel("Project: ").SetDoneFunc(func(key tcell.Key) {
		config.Project = projectSelect.GetText()
		updatePipelinesTable(config, pipelinesTable)
		app.SetFocus(pipelinesTable)
	})

	projectSelect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
		}
		
		return event
	})

	return projectSelect
}
