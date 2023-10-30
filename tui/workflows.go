package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func ShowWorkflows(config config.CirclogConfig, project string, pipeline circleci.Pipeline, app *tview.Application, layout *tview.Flex) {
	workflowsArea := tview.NewFlex()
	workflowsArea.SetTitle(fmt.Sprintf(" %d - WORKFLOWS ", pipeline.Number)).SetBorder(true)

	workflowsTable := tview.NewTable().SetBorders(true)
	workflowsTable.SetSelectable(true, false)

	for column, header := range []string{"Name", "Created At"} {
		workflowsTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)))
	}

	workflows, _ := circleci.GetPipelineWorkflows(config, project, pipeline.Id)

	for row, workflow := range workflows {
		for column, attr := range []string{workflow.Name, workflow.CreatedAt} {
			cell := tview.NewTableCell(attr).SetStyle(StyleForStatus(workflow.Status))
			cell.SetReference(workflow)
			workflowsTable.SetCell(row+1, column, cell)
		}
	}

	workflowsTable.Select(1, 1)
	workflowsTable.SetSelectedFunc(func(row int, col int) {
		cell := workflowsTable.GetCell(row, 0)
		workflow := cell.GetReference()
		layout.RemoveItem(workflowsArea)
		ShowJobs(config, project, workflow.(circleci.Workflow), app, layout)
	})

	workflowsArea.AddItem(workflowsTable, 0, 1, false)
	layout.AddItem(workflowsArea, 0, 1, false)

	app.SetFocus(workflowsTable)
}
