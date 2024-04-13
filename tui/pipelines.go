package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type pipelinesPane struct {
	table       *tview.Table
	numPages    int
	update      chan bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (cTui *CirclogTui) newPipelinesPane() pipelinesPane {
	table := tview.NewTable()
	table.SetTitle(" PIPELINES ")
	table.SetBorder(true)
	table.SetBorderColor(tcell.ColorGrey)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Number", "Branch/Tag", "Start", "Trigger"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Pipeline:
			cTui.updateState <- cellRef
			cTui.workflows.update <- true
			cTui.app.SetFocus(cTui.workflows.table)

		case string:
			if cell.Text == "..." {
				cTui.pipelines.numPages++
				cTui.pipelines.update <- true
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.watchCancelAll()
			table.SetBorderColor(tcell.ColorGrey)
			cTui.config.Project = ""
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.projectSelect)

			return event
		}

		switch event.Rune() {

		case 'v':
			cell := table.GetCell(table.GetSelection())
			cellRef := cell.GetReference()
			switch cellRef := cellRef.(type) {
			case circleci.Pipeline:
				if cellRef.Vcs.Branch != "" {
					cTui.pipelines.restartWatcher(func() {
						cTui.pipelines.clear()
						cTui.pipelines.numPages = 1
						cTui.config.Branch = cellRef.Vcs.Branch
						cTui.branchSelect.SetText(cTui.config.Branch)
						cTui.pipelines.update <- true
					})
				}
			}

		case 'b':
			cTui.pipelines.restartWatcher(func() {
				cTui.clearAll()
				table.SetBorderColor(tcell.ColorGrey)
				cTui.config.Branch = ""
				cTui.app.SetFocus(cTui.branchSelect)
			})

		case 'd':
			cTui.watchCancelAll()
			cTui.app.Stop()
			fmt.Printf("circlog pipelines %s\n", cTui.config.Project)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.pipelines.restartWatcher(func() {
			table.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.SetText(`Branch Select            [B]
			Filter by branch         [V]`)
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())
	pipelinesPane := pipelinesPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}

	pipelinesPane.update = pipelinesPane.updater(cTui)

	return pipelinesPane
}

func (p *pipelinesPane) watchPipelines(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			p.update <- true
		}
	}
}

func (p *pipelinesPane) restartWatcher(fn func()) {
	p.watchCancel()
	fn()
	p.watchCtx, p.watchCancel = context.WithCancel(context.TODO())
	go p.watchPipelines(p.watchCtx)
}

func (p *pipelinesPane) updater(cTui *CirclogTui) chan bool {
	updateChan := make(chan bool)

	go func() {
		for <-updateChan {
			pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, cTui.pipelines.numPages, "")
			cTui.app.QueueUpdateDraw(func() {
				cTui.pipelines.restartWatcher(func() {
					currentRow := 1
					if p.table.GetRowCount() != 1 {
						currentRow, _ = p.table.GetSelection()
					}

					p.clear()
					p.populatePipelinesTable(pipelines, nextPageToken)

					if currentRow == 1 {
						p.table.ScrollToBeginning()
					} else {
						p.table.Select(currentRow, 0)
					}
				})
			})
		}
	}()

	return updateChan
}

func (p *pipelinesPane) populatePipelinesTable(pipelines []circleci.Pipeline, nextPageToken string) {
	if len(pipelines) != 0 {
		for row, pipeline := range pipelines {
			for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt.Local().Format(time.RFC822Z), pipeline.Trigger.Type} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(pipeline.State))
				cell.SetReference(pipeline)
				p.table.SetCell(row+1, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...").SetReference("...")
			p.table.SetCell(p.table.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		p.table.SetCell(1, 0, cell)
	}
}

func (p *pipelinesPane) clear() {
	row := 1
	for row < p.table.GetRowCount() {
		p.table.RemoveRow(row)
	}
}
