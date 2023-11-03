package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newPipelinesTable(config config.CirclogConfig, project string) *tview.Table {
	pipelinesTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0)
	pipelinesTable.SetTitle(" PIPELINES ").SetBorder(true)

	for column, header := range []string{"Number", "Branch/Tag", "Start"} {
		pipelinesTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	pipelinesTable.Select(1, 1)
	pipelinesTable.SetSelectedFunc(func(row int, col int) {
		cell := pipelinesTable.GetCell(row, 0)
		if cell.Text != "None" {
			pipeline := cell.GetReference().(circleci.Pipeline)
			updateWorkflowsTable(config, project, pipeline, workflowsTable)
		}
	})

	updatePipelinesTable(config, project, pipelinesTable)

	return pipelinesTable
}

func updatePipelinesTable(config config.CirclogConfig, project string, pipelinesTable *tview.Table) {
	pipelines, _ := circleci.GetProjectPipelines(config)

	if len(pipelines) != 0 {
		for row, pipeline := range pipelines {
			for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt.Local().String()} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(pipeline.State))
				cell.SetReference(pipeline)
				pipelinesTable.SetCell(row+1, column, cell)
			}
		}
	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		pipelinesTable.SetCell(1, 0, cell)
	}
}
