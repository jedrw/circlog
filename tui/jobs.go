package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func ShowJobs(config config.CirclogConfig, project string, workflow circleci.Workflow, app *tview.Application, layout *tview.Flex) {
	jobsArea := tview.NewFlex()
	jobsArea.SetTitle(fmt.Sprintf(" %s - JOBS ", workflow.Name)).SetBorder(true)

	jobsTable := tview.NewTable().SetBorders(true)
	jobsTable.SetSelectable(true, false)

	for column, header := range []string{"Name", "Created At"} {
		jobsTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)))
	}

	jobs, _ := circleci.GetWorkflowJobs(config, project, workflow.Id)

	for row, job := range jobs {
		for column, attr := range []string{job.Name, job.StartedAt} {
			cell := tview.NewTableCell(attr).SetStyle(StyleForStatus(job.Status))
			cell.SetReference(job)
			jobsTable.SetCell(row+1, column, cell)
		}
	}

	jobsTable.Select(1, 1)
	jobsTable.SetSelectedFunc(func(row int, col int) {
		cell := jobsTable.GetCell(row, 0)
		jobNumber := cell.GetReference()
		layout.RemoveItem(jobsArea)
		ShowSteps(config, project, jobNumber.(circleci.Job), app, layout)
	})

	jobsArea.AddItem(jobsTable, 0, 1, false)
	layout.AddItem(jobsArea, 0, 1, false)

	app.SetFocus(jobsTable)
}
