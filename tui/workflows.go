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
	update      chan bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (cTui *CirclogTui) newWorkflowsPane() workflowsPane {
	table := tview.NewTable()
	table.SetTitle(" WORKFLOWS ")
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
			cTui.updateState <- cellRef
			cTui.jobs.update <- true
			cTui.app.SetFocus(cTui.jobs.table)

		case string:
			if cell.Text == "..." {
				cTui.workflows.numPages++
				cTui.workflows.update <- true
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

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog workflows %s -l %s\n", cTui.config.Project, cTui.state.pipeline.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.workflows.restartWatcher(func() {
			table.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.Clear()
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	workflowsPane := workflowsPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}

	workflowsPane.update = workflowsPane.updater(cTui)

	return workflowsPane
}

func (w *workflowsPane) watchWorkflows(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			w.update <- true
		}
	}
}

func (w *workflowsPane) restartWatcher(fn func()) {
	w.watchCancel()
	fn()
	w.watchCtx, w.watchCancel = context.WithCancel(context.TODO())
	go w.watchWorkflows(w.watchCtx)
}

func (w *workflowsPane) updater(cTui *CirclogTui) chan bool {
	updateChan := make(chan bool)

	go func() {
		for <-updateChan {
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.state.pipeline.Id, w.numPages, "")
			cTui.app.QueueUpdateDraw(func() {
				currentRow := 1
				if w.table.GetRowCount() != 1 {
					currentRow, _ = w.table.GetSelection()
				}

				w.clear()
				w.populateWorkflowsTable(workflows, nextPageToken)

				if currentRow == 1 {
					w.table.ScrollToBeginning()
				} else {
					w.table.Select(currentRow, 0)
				}
			})
		}
	}()

	return updateChan
}

func (w *workflowsPane) populateWorkflowsTable(workflows []circleci.Workflow, nextPageToken string) {
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
			cell := tview.NewTableCell("...").SetReference("...")
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
