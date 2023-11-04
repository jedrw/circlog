package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newWorkflowsTable(config config.CirclogConfig, project string) *tview.Table {
	workflowsTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	workflowsTable.SetTitle(" WORKFLOWS ").SetBorder(true)

	workflowsTable.SetSelectedFunc(func(row int, col int) {
		cell := workflowsTable.GetCell(row, 0)
		if cell.Text != "None" && cell.Text != "" {
			workflow := cell.GetReference().(circleci.Workflow)
			updateJobsTable(config, project, workflow, jobsTable)
		}
	})

	return workflowsTable
}

func updateWorkflowsTable(config config.CirclogConfig, project string, pipeline circleci.Pipeline, workflowsTable *tview.Table) {
	workflows, _ := circleci.GetPipelineWorkflows(config, project, pipeline.Id)

	workflowsTable.Clear()

	for column, header := range []string{"Name", "Duration"} {
		workflowsTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	if len(workflows) != 0 {
		for row, workflow := range workflows {
			for column, attr := range []string{workflow.Name, workflow.StoppedAt.Sub(workflow.CreatedAt).String()} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(workflow.Status))
				cell.SetReference(workflow)
				workflowsTable.SetCell(row+1, column, cell)
			}
		}
	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		workflowsTable.SetCell(1, 0, cell)
	}

	workflowsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyBackspace2 {
			workflowsTable.Clear()
			app.SetFocus(pipelinesTable)
		}

		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog workflows %s -l %s\n", project, pipeline.Id)
		}

		return event
	})

	app.SetFocus(workflowsTable)
}
