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
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (cTui *CirclogTui) newPipelinesPane() pipelinesPane {
	table := tview.NewTable()
	table.SetTitle(" PIPELINES ")
	table.SetBorder(true)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Number", "Branch/Tag", "Start", "Trigger"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Pipeline:
			cTui.state.pipeline = cellRef
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.state.pipeline.Id, 1, "")
			cTui.workflows.populateWorkflowsTable(workflows, nextPageToken)
			cTui.app.SetFocus(cTui.workflows.table)

		case string:
			if cell.Text == "..." {
				cTui.pipelines.restartWatcher(cTui, func() {
					nextPageToken := cell.GetReference().(string)
					newPipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, nextPageToken)
					cTui.pipelines.addPipelinesToTable(newPipelines, table.GetRowCount()-1, nextPageToken)
					cTui.pipelines.numPages++
				})
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.clearAll()
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
					cTui.pipelines.restartWatcher(cTui, func() {
						cTui.pipelines.clear()
						cTui.pipelines.numPages = 1
						cTui.config.Branch = cellRef.Vcs.Branch
						cTui.branchSelect.SetText(cTui.config.Branch)
						pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, "")
						cTui.pipelines.populateTable(pipelines, nextPageToken)
						cTui.pipelines.table.ScrollToBeginning()
					})
				}
			}

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.watchCancelAll()
			cTui.app.Stop()
			fmt.Printf("circlog pipelines %s\n", cTui.config.Project)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.pipelines.restartWatcher(cTui, func() {
			cTui.paneControls.SetText(`Filter by branch         [V]`)
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	return pipelinesPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (p *pipelinesPane) watchPipelines(ctx context.Context, cTui *CirclogTui) {
	pipelinesChan := make(chan []circleci.Pipeline)
	nextPageTokenChan := make(chan string)

	for {
		go func() {
			pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, p.numPages, "")
			pipelinesChan <- pipelines
			nextPageTokenChan <- nextPageToken
		}()

		select {
		case <-ctx.Done():
			return

		default:
			time.Sleep(refreshInterval)
			pipelines := <-pipelinesChan
			nextPageToken := <-nextPageTokenChan
			cTui.app.QueueUpdateDraw(func() {
				p.clear()
				p.addPipelinesToTable(pipelines, 1, nextPageToken)
			})
		}
	}
}

func (p *pipelinesPane) restartWatcher(cTui *CirclogTui, fn func()) {
	p.watchCancel()
	fn()
	p.watchCtx, p.watchCancel = context.WithCancel(context.TODO())
	go p.watchPipelines(p.watchCtx, cTui)
}

func (p *pipelinesPane) populateTable(pipelines []circleci.Pipeline, nextPageToken string) {
	p.clear()
	p.addPipelinesToTable(pipelines, p.table.GetRowCount(), nextPageToken)
}

func (p *pipelinesPane) addPipelinesToTable(pipelines []circleci.Pipeline, startRow int, nextPageToken string) {
	if len(pipelines) != 0 {
		for row, pipeline := range pipelines {
			for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt.Local().Format(time.RFC822Z), pipeline.Trigger.Type} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(pipeline.State))
				cell.SetReference(pipeline)
				p.table.SetCell(row+startRow, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...")
			cell.SetReference(nextPageToken)
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
