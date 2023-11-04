package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newPipelinesTable(config config.CirclogConfig, project string) *tview.Table {
	pipelinesTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	pipelinesTable.SetTitle(" PIPELINES ").SetBorder(true)

	for column, header := range []string{"Number", "Branch/Tag", "Start", "Trigger"} {
		pipelinesTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	pipelinesTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog pipelines %s\n", project)
		}

		return event
	})

	pipelinesTable.Select(1, 1)
	pipelinesTable.SetSelectedFunc(func(row int, col int) {
		cell := pipelinesTable.GetCell(row, 0)
		if cell.Text != "None" && cell.Text != "" {
			pipeline := cell.GetReference().(circleci.Pipeline)
			updateWorkflowsTable(config, project, pipeline, workflowsTable)
		}
	})

	return pipelinesTable
}

func updatePipelinesTable(config config.CirclogConfig, project string, pipelinesTable *tview.Table) {
	pipelines, _ := circleci.GetProjectPipelines(config)

	if len(pipelines) != 0 {
		for row, pipeline := range pipelines {
			for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt.Local().Format(time.RFC822Z), pipeline.Trigger.Type} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(pipeline.State))
				cell.SetReference(pipeline)
				pipelinesTable.SetCell(row+1, column, cell)
			}
		}
	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		pipelinesTable.SetCell(1, 0, cell)
	}

	pipelinesTable.ScrollToBeginning().Select(0, 0)

	// This function is called as a go routine so we must tell the application focus and draw once done.
	app.SetFocus(pipelinesTable).Draw()
}
