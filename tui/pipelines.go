package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func ShowPipelines(config config.CirclogConfig, project string, app *tview.Application, layout *tview.Flex) *tview.Table {
	pipelinesArea := tview.NewFlex()
	pipelinesArea.SetTitle(" PIPELINES ").SetBorder(true)

	pipelinesTable := tview.NewTable().SetBorders(true)
	pipelinesTable.SetSelectable(true, false)

	for column, header := range []string{"Number", "Branch/Tag", "Start"} {
		pipelinesTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)))
	}

	pipelines, _ := circleci.GetProjectPipelines(config, project)

	for row, pipeline := range pipelines {
		for column, attr := range []string{fmt.Sprint(pipeline.Number), branchOrTag(pipeline), pipeline.CreatedAt} {
			cell := tview.NewTableCell(attr).SetStyle(StyleForStatus(pipeline.State))
			cell.SetReference(pipeline)
			pipelinesTable.SetCell(row+1, column, cell)
		}

	}

	pipelinesTable.Select(1, 1)
	pipelinesTable.SetSelectedFunc(func(row int, col int) {
		cell := pipelinesTable.GetCell(row, 0)
		pipeline := cell.GetReference()
		layout.RemoveItem(pipelinesArea)
		ShowWorkflows(config, project, pipeline.(circleci.Pipeline), app, layout)
	})

	pipelinesArea.AddItem(pipelinesTable, 0, 1, false)
	layout.AddItem(pipelinesArea, 0, 1, false)

	app.SetFocus(pipelinesTable)

	return pipelinesTable
}
