package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type pipelinesTable struct {
	table         *tview.Table
	numPages      int
	refreshCtx    context.Context
	refreshCancel context.CancelFunc
}

func (cTui *CirclogTui) newPipelinesTable() pipelinesTable {
	table := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	table.SetTitle(" PIPELINES ").SetBorder(true)

	for column, header := range []string{"Number", "Branch/Tag", "Start", "Trigger"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {
		case circleci.Pipeline:
			cTui.tuiState.pipeline = cellRef
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.tuiState.pipeline.Id, 1, "")
			cTui.workflows.populateWorkflowsTable(workflows, nextPageToken)
			cTui.app.SetFocus(cTui.workflows.table)
		case string:
			if cell.Text == "..." {
				cTui.pipelines.refreshCancel()
				nextPageToken := cell.GetReference().(string)
				newPipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, nextPageToken)
				cTui.pipelines.addPipelinesToTable(newPipelines, table.GetRowCount()-1, nextPageToken)
				cTui.pipelines.numPages++
				cTui.pipelines.refreshCtx, cTui.pipelines.refreshCancel = context.WithCancel(context.TODO())
				go cTui.refreshPipelinesTable(cTui.pipelines.refreshCtx)
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
		case 'f':
			cell := table.GetCell(table.GetSelection())
			cellRef := cell.GetReference()
			switch cellRef := cellRef.(type) {
			case circleci.Pipeline:
				if cellRef.Vcs.Branch != "" {
					cTui.config.Branch = cellRef.Vcs.Branch
					pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, "")
					cTui.pipelines.clear()
					cTui.pipelines.populateTable(pipelines, nextPageToken)
					cTui.branchSelect.SetText(cTui.config.Branch)
				}
			}

		case 'b':
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.refreshCancelAll()
			cTui.app.Stop()
			fmt.Printf("circlog pipelines %s\n", cTui.config.Project)
		}

		return event
	})

	pipelinesControlBindings := `Move	           [Up/Down]
		Select               [Enter]
		Select branch            [B]
		Filter by branch         [F]
		Dump command             [D]
		Back/Quit              [Esc]
	`

	table.SetFocusFunc(func() {
		cTui.pipelines.refreshCancel()
		cTui.controls.SetText(pipelinesControlBindings)
		cTui.pipelines.refreshCtx, cTui.pipelines.refreshCancel = context.WithCancel(context.TODO())
		go cTui.refreshPipelinesTable(cTui.pipelines.refreshCtx)
	})

	refreshCtx, refreshCancel := context.WithCancel(context.TODO())

	return pipelinesTable{
		table:         table,
		refreshCtx:    refreshCtx,
		refreshCancel: refreshCancel,
	}
}

func (p pipelinesTable) populateTable(pipelines []circleci.Pipeline, nextPageToken string) {
	p.clear()
	p.addPipelinesToTable(pipelines, p.table.GetRowCount(), nextPageToken)
}

func (p pipelinesTable) addPipelinesToTable(pipelines []circleci.Pipeline, startRow int, nextPageToken string) {
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

func (p pipelinesTable) clear() {
	row := 1
	for row < p.table.GetRowCount() {
		p.table.RemoveRow(row)
	}
}
