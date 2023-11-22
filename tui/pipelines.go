package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newPipelinesTable(config config.CirclogConfig) *tview.Table {
	pipelinesTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	pipelinesTable.SetTitle(" PIPELINES ").SetBorder(true)

	for column, header := range []string{"Number", "Branch/Tag", "Start", "Trigger"} {
		pipelinesTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	return pipelinesTable
}

func updatePipelinesTable(config config.CirclogConfig, pipelinesTable *tview.Table) {
	pipelines, nextPageToken, _ := circleci.GetProjectPipelines(config, 1, "")

	addPipelinesToTable(pipelines, pipelinesTable.GetRowCount(), nextPageToken)

	pipelinesTable.SetSelectedFunc(func(row int, col int) {
		cell := pipelinesTable.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {
		case circleci.Pipeline:
			updateWorkflowsTable(config, cellRef)
		case string:
			if cell.Text == "Next page..." {
				nextPageToken := cell.GetReference().(string)
				newPipelines, nextPageToken, _ := circleci.GetProjectPipelines(config, 1, nextPageToken)
				addPipelinesToTable(newPipelines, pipelinesTable.GetRowCount()-1, nextPageToken)
			}
		}
	})

	pipelinesTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
			config.Project = ""
			Run(config)
		}
		
		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog pipelines %s\n", config.Project)
		}

		return event
	})

	pipelinesTable.ScrollToBeginning().Select(0, 0)

	// This function is called as a go routine so we must tell the application focus and draw once done.
	app.SetFocus(pipelinesTable)
}

func addPipelinesToTable(pipelines []circleci.Pipeline, startRow int, nextPageToken string) {
	if len(pipelines) != 0 {
		for row, pipeline := range pipelines {
			for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt.Local().Format(time.RFC822Z), pipeline.Trigger.Type} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(pipeline.State))
				cell.SetReference(pipeline)
				pipelinesTable.SetCell(row+startRow, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("Next page...")
			cell.SetReference(nextPageToken)
			pipelinesTable.SetCell(pipelinesTable.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		pipelinesTable.SetCell(1, 0, cell)
	}
}
