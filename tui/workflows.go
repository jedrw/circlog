package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type workflowsPane struct {
	table       *tview.Table
	numPages    int
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (cTui *CirclogTui) newWorkflowsPane() workflowsPane {
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
			cTui.state.workflow = cellRef
			jobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.state.workflow.Id, 1, "")
			cTui.jobs.populateTable(jobs, nextPageToken)
			cTui.app.SetFocus(cTui.jobs.table)

		case string:
			if cell.Text == "..." {
				cTui.workflows.restartWatcher(cTui, func() {
					nextPageToken := cell.GetReference().(string)
					newWorkflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.state.pipeline.Id, 1, nextPageToken)
					cTui.workflows.addWorkflowsToTable(newWorkflows, nextPageToken)
					cTui.workflows.numPages++
				})
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.workflows.watchCancel()
			cTui.workflows.clear()
			table.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.pipelines.table)

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
		cTui.workflows.restartWatcher(cTui, func() {
			table.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.Clear()
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	return workflowsPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (w *workflowsPane) watchWorkflows(ctx context.Context, cTui *CirclogTui) {
	workflowsChan := make(chan []circleci.Workflow)
	nextPageTokenChan := make(chan string)
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		go func() {
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.state.pipeline.Id, cTui.pipelines.numPages, "")
			workflowsChan <- workflows
			nextPageTokenChan <- nextPageToken
		}()

		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case workflows := <-workflowsChan:
			nextPageToken := <-nextPageTokenChan
			cTui.app.QueueUpdateDraw(func() {
				w.clear()
				w.addWorkflowsToTable(workflows, nextPageToken)
			})

			<-ticker.C
		}
	}
}

func (w *workflowsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	w.watchCancel()
	fn()
	w.watchCtx, w.watchCancel = context.WithCancel(context.TODO())
	go w.watchWorkflows(w.watchCtx, cTui)
}

func (w *workflowsPane) populateWorkflowsTable(workflows []circleci.Workflow, nextPageToken string) {
	w.clear()
	w.addWorkflowsToTable(workflows, nextPageToken)
}

func (w *workflowsPane) addWorkflowsToTable(workflows []circleci.Workflow, nextPageToken string) {
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
				w.table.SetCell(row+1, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...")
			cell.SetReference(nextPageToken)
			w.table.SetCell(w.table.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		w.table.SetCell(1, 0, cell)
	}
}

func (w *workflowsPane) clear() {
	row := 1
	for row < w.table.GetRowCount() {
		w.table.RemoveRow(row)
	}
}
