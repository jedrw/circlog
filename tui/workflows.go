package tui

import (
	"context"
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

func (w *workflowsPane) watchWorkflows(cTui *CirclogTui) {
	workflowsChan := make(chan []circleci.Workflow)
	nextPageTokenChan := make(chan string)

	for {
		go func() {
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.tuiState.pipeline.Id, cTui.pipelines.numPages, "")
			workflowsChan <- workflows
			nextPageTokenChan <- nextPageToken
		}()

		time.Sleep(refreshInterval)

		select {
		case <-cTui.workflows.watchCtx.Done():
			return

		default:
			workflows := <-workflowsChan
			nextPageToken := <-nextPageTokenChan
			cTui.app.QueueUpdateDraw(func() {
				cTui.workflows.clear()
				cTui.workflows.addWorkflowsToTable(workflows, 1, nextPageToken)
			})
		}
	}
}

func (w *workflowsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	cTui.workflows.watchCancel()
	fn()
	cTui.workflows.watchCtx, cTui.workflows.watchCancel = context.WithCancel(context.TODO())
	go cTui.workflows.watchWorkflows(cTui)
}

func (w *workflowsPane) populateWorkflowsTable(workflows []circleci.Workflow, nextPageToken string) {
	w.clear()
	w.addWorkflowsToTable(workflows, 1, nextPageToken)
}

func (w *workflowsPane) addWorkflowsToTable(workflows []circleci.Workflow, startRow int, nextPageToken string) {
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
