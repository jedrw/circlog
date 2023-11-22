package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

var (
	app            *tview.Application
	controls       *tview.TextView
	logsView       *tview.TextView
	stepsTree      *tview.TreeView
	jobsTable      *tview.Table
	workflowsTable *tview.Table
	pipelinesTable *tview.Table
)

func Run(config config.CirclogConfig) {
	app = tview.NewApplication()

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.SetTitle(" circlog ").SetBorder(true).SetBorderPadding(1, 1, 1, 1)

	heading := tview.NewFlex().SetDirection(tview.FlexColumn)

	branch := config.Branch
	if config.Branch == "" {
		branch = "ALL"
	}

	info := tview.NewFlex().SetDirection(tview.FlexRow)
	heading.AddItem(info, 0, 1, true)

	projectSelect := newProjectSelect(&config)
	info.AddItem(projectSelect, 1, 1, true)
	
	configText := tview.NewTextView().SetText(fmt.Sprintf("Branch: %s\nOrganisation: %s", branch, config.Org))
	info.AddItem(configText, 0, 1, false)

	controls = tview.NewTextView().SetTextAlign(tview.AlignRight)
	controls.SetText(controlBindings)
	heading.AddItem(controls, 0, 1, false)

	layout.AddItem(heading, 6, 0, false)

	upperNav := tview.NewFlex().SetDirection(tview.FlexColumn)
	layout.AddItem(upperNav, 0, 2, false)

	lowerNav := tview.NewFlex().SetDirection(tview.FlexColumn)
	layout.AddItem(lowerNav, 0, 3, false)

	logsView = newLogsView()
	stepsTree = newStepsTree()
	jobsTable = newJobsTable(config)
	workflowsTable = newWorkflowsTable(config)
	pipelinesTable = newPipelinesTable(config)

	upperNav.AddItem(pipelinesTable, 0, 1, false)
	upperNav.AddItem(workflowsTable, 0, 1, false)
	upperNav.AddItem(jobsTable, 0, 1, false)
	lowerNav.AddItem(stepsTree, 0, 1, false)
	lowerNav.AddItem(logsView, 0, 2, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
		}
		return event
	})

	if config.Project != "" {
		// enables the TUI to be drawn while waiting for the pipelines to be returned from CircleCi.
		go updatePipelinesTable(config, pipelinesTable)
	}

	err := app.SetRoot(layout, true).SetFocus(info).Run()
	if err != nil {
		panic(err)
	}
}
