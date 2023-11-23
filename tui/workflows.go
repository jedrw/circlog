package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newWorkflowsTable(config config.CirclogConfig) *tview.Table {
	workflowsTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	workflowsTable.SetTitle(" WORKFLOWS ").SetBorder(true)

	return workflowsTable
}

func updateWorkflowsTable(config config.CirclogConfig, pipeline circleci.Pipeline) {
	workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(config, pipeline.Id, 1, "")

	workflowsTable.Clear()

	for column, header := range []string{"Name", "Duration"} {
		workflowsTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	addWorkflowsToTable(workflows, workflowsTable.GetRowCount(), nextPageToken)

	workflowsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			workflowsTable.Clear()
			app.SetFocus(pipelinesTable)
		}

		if event.Rune() == 'b' {
			app.SetFocus(branchSelect)
		}

		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog workflows %s -l %s\n", config.Project, pipeline.Id)
		}

		return event
	})

	workflowsTable.SetSelectedFunc(func(row int, col int) {
		cell := workflowsTable.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {
		case circleci.Workflow:
			updateJobsTable(config, config.Project, cellRef, jobsTable)
		case string:
			if cell.Text == "..." {
				nextPageToken := cell.GetReference().(string)
				newWorkflows, nextPageToken, _ := circleci.GetPipelineWorkflows(config, pipeline.Id, 1, nextPageToken)
				addWorkflowsToTable(newWorkflows, workflowsTable.GetRowCount(), nextPageToken)
			}
		}
	})

	app.SetFocus(workflowsTable)
}

func addWorkflowsToTable(workflows []circleci.Workflow, startRow int, nextPageToken string) {
	if len(workflows) != 0 {
		for row, workflow := range workflows {
			var workflowDuration string
			if workflow.Status == circleci.RUNNING {
				workflowDuration = time.Since(workflow.CreatedAt).Round(time.Millisecond).String()
			} else {
				workflowDuration = workflow.StoppedAt.Sub(workflow.CreatedAt).Round(time.Millisecond).String()
			}

			for column, attr := range []string{workflow.Name, workflowDuration} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(workflow.Status))
				cell.SetReference(workflow)
				workflowsTable.SetCell(row+1, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...")
			cell.SetReference(nextPageToken)
			workflowsTable.SetCell(workflowsTable.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		workflowsTable.SetCell(1, 0, cell)
	}
}
