package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) newPipelinesPane() *tview.Table {
	table := tview.NewTable()
	table.SetTitle(" PIPELINES ")
	table.SetBackgroundColor(tcell.ColorDefault)
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
			cTui.state.mu.RLock()
			cTui.state.pipeline = cellRef
			cTui.state.mu.RUnlock()
			workflows, nextPageToken, _ := circleci.GetWorkflows(cTui.config, cTui.state.pipeline.Id, 1, "")
			cTui.populateWorkflowsTable(workflows, nextPageToken)
			cTui.app.SetFocus(cTui.workflowsPane)

		case string:
			if cell.Text == "..." {
				cTui.state.mu.Lock()
				cTui.state.pipelinePages++
				cTui.state.mu.Unlock()
				nextPageToken := cell.GetReference().(string)
				newPipelines, nextPageToken, _ := circleci.GetPipelines(cTui.config, 1, nextPageToken)
				cTui.addPipelinesToTable(newPipelines, table.GetRowCount()-1, nextPageToken)
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.clearAll()
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
					cTui.clearPipelinesTable()
					cTui.state.mu.Lock()
					cTui.state.pipelinePages = 1
					cTui.state.mu.Unlock()
					cTui.config.Branch = cellRef.Vcs.Branch
					cTui.branchSelect.SetText(cTui.config.Branch)
					pipelines, nextPageToken, _ := circleci.GetPipelines(cTui.config, 1, "")
					cTui.populatePipelinesTable(pipelines, nextPageToken)
					cTui.pipelinesPane.ScrollToBeginning()
				}
			}

		case 'b':
			cTui.clearAll()
			table.SetBorderColor(tcell.ColorGrey)
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog pipelines %s\n", cTui.config.Project)
		}

		return event
	})

	table.SetFocusFunc(func() {
		table.SetBorderColor(tcell.ColorDefault)
		cTui.paneControls.SetText("Filter by branch\t[V]")
	})

	return table
}

func (cTui *CirclogTui) populatePipelinesTable(pipelines []circleci.Pipeline, nextPageToken string) {
	cTui.clearPipelinesTable()
	cTui.addPipelinesToTable(pipelines, cTui.pipelinesPane.GetRowCount(), nextPageToken)
}

func (cTui *CirclogTui) addPipelinesToTable(pipelines []circleci.Pipeline, startRow int, nextPageToken string) {
	if len(pipelines) != 0 {
		for row, pipeline := range pipelines {
			for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt.Local().Format(time.RFC822Z), pipeline.Trigger.Type} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(pipeline.State))
				cell.SetReference(pipeline)
				cTui.pipelinesPane.SetCell(row+startRow, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...")
			cell.SetReference(nextPageToken)
			cTui.pipelinesPane.SetCell(cTui.pipelinesPane.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		cTui.pipelinesPane.SetCell(1, 0, cell)
	}
}

func (cTui *CirclogTui) clearPipelinesTable() {
	row := 1
	for row < cTui.pipelinesPane.GetRowCount() {
		cTui.pipelinesPane.RemoveRow(row)
	}
}
