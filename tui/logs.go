package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newLogsView() *tview.TextView {
	logsView := tview.NewTextView()
	logsView.SetTitle(" LOGS ").SetBorder(true).SetBorderPadding(0, 0, 1, 1)

	return logsView
}

func updateLogsView(config config.CirclogConfig, project string, job circleci.Job, action circleci.Action, logsview *tview.TextView) {
	logs, _ := circleci.GetStepLogs(config, project, job.JobNumber, action.Step, action.Index, action.AllocationId)
	logsview.SetText(logs).ScrollToBeginning()

	logsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyBackspace2 {
			logsView.Clear()
			controls.SetText(controlBindings)
			app.SetFocus(stepsTree)
		}

		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog logs %s -j %d -s %d -i %d -a \"%s\"\n", project, job.JobNumber, action.Step, action.Index, action.AllocationId)
		}

		return event
	})

	controls.SetText(controlBindings + "Dump Logs Command        [D]")

	app.SetFocus(logsview)
}
