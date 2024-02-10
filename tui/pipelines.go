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

func (p *pipelinesPane) watchPipelines(cTui *CirclogTui) {
	pipelinesChan := make(chan []circleci.Pipeline)
	nextPageTokenChan := make(chan string)

	for {
		go func() {
			pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, cTui.pipelines.numPages, "")
			pipelinesChan <- pipelines
			nextPageTokenChan <- nextPageToken
		}()

		time.Sleep(refreshInterval)

		select {
		case <-cTui.pipelines.watchCtx.Done():
			return

		default:
			pipelines := <-pipelinesChan
			nextPageToken := <-nextPageTokenChan
			cTui.app.QueueUpdateDraw(func() {
				cTui.pipelines.clear()
				cTui.pipelines.addPipelinesToTable(pipelines, 1, nextPageToken)
			})
		}
	}
}

func (p *pipelinesPane) restartWatcher(cTui *CirclogTui, fn func()) {
	cTui.pipelines.watchCancel()
	fn()
	cTui.pipelines.watchCtx, cTui.pipelines.watchCancel = context.WithCancel(context.TODO())
	go cTui.pipelines.watchPipelines(cTui)
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
