package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) newWorkflowsPane() *tview.Table {
	table := tview.NewTable()
	table.SetTitle(" WORKFLOWS ")
	table.SetBackgroundColor(tcell.ColorDefault)
	table.SetBorder(true)
	table.SetBorderColor(tcell.ColorGrey)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Name", "Duration"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Workflow:
			cTui.state.mu.Lock()
			cTui.state.workflow = cellRef
			jobs, nextPageToken, _ := circleci.GetJobs(cTui.config, cTui.state.workflow.Id, 1, "")
			cTui.state.mu.Unlock()
			cTui.populateJobsTable(jobs, nextPageToken)
			cTui.app.SetFocus(cTui.jobsPane)

		case string:
			if cell.Text == "..." {
				cTui.state.mu.Lock()
				cTui.state.workflowPages++
				nextPageToken := cell.GetReference().(string)
				newWorkflows, nextPageToken, _ := circleci.GetWorkflows(cTui.config, cTui.state.pipeline.Id, 1, nextPageToken)
				cTui.state.mu.Unlock()
				cTui.addWorkflowsToTable(newWorkflows, nextPageToken)
			}

		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.clearWorkflowsTable()
			table.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.pipelinesPane)

			return event
		}

		switch event.Rune() {

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog workflows %s -l %s\n", cTui.config.Project, cTui.state.pipeline.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		table.SetBorderColor(tcell.ColorDefault)
		cTui.paneControls.Clear()
	})

	return table
}

func (cTui *CirclogTui) populateWorkflowsTable(workflows []circleci.Workflow, nextPageToken string) {
	cTui.clearWorkflowsTable()
	cTui.addWorkflowsToTable(workflows, nextPageToken)
}

func (cTui *CirclogTui) addWorkflowsToTable(workflows []circleci.Workflow, nextPageToken string) {
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
				cTui.workflowsPane.SetCell(row+1, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...")
			cell.SetReference(nextPageToken)
			cTui.workflowsPane.SetCell(cTui.workflowsPane.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		cTui.workflowsPane.SetCell(1, 0, cell)
	}
}

func (cTui *CirclogTui) clearWorkflowsTable() {
	row := 1
	for row < cTui.workflowsPane.GetRowCount() {
		cTui.workflowsPane.RemoveRow(row)
	}
}
