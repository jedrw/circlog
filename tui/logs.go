package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func ShowLogs(config config.CirclogConfig, project string, jobNumber int64, action circleci.Action, app *tview.Application, layout *tview.Flex) {
	logsArea := tview.NewFlex()
	logsArea.SetTitle(fmt.Sprintf(" %s - LOGS ", action.Name)).SetBorder(true)

	logs, _ := circleci.GetStepLogs(config, project, jobNumber, action.Step, action.Index, action.AllocationId)
	logsOutput := tview.NewTextView().SetText(logs)

	logsOutput.SetDoneFunc(func(key tcell.Key) {
		app.Stop()
		fmt.Printf("circlog logs %s -j %d -s %d -i %d -a \"%s\"\n", project, jobNumber, action.Step, action.Index, action.AllocationId)
	})

	logsArea.AddItem(logsOutput, 0, 1, false)
	layout.AddItem(logsArea, 0, 1, false)

	app.SetFocus(logsOutput)
}
