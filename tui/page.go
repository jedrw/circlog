package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
)

var (
	colourByStatus = map[string]tcell.Color{
		"success":      tcell.ColorDarkGreen,
		"running":      tcell.ColorLightGreen,
		"not_run":      tcell.ColorGray,
		"failed":       tcell.ColorDarkRed,
		"error":        tcell.ColorDarkRed,
		"failing":      tcell.ColorPink,
		"on_hold":      tcell.ColorYellow,
		"canceled":     tcell.ColorDarkRed,
		"unauthorized": tcell.ColorDarkRed,

		"created":       tcell.ColorLightGreen,
		"errored":       tcell.ColorDarkRed,
		"setup-pending": tcell.ColorGrey,
		"setup":         tcell.ColorLightGray,
		"pending":       tcell.ColorYellowGreen,
	}

	controlBindings = `Move	           [Up/Down]
		Select               [Enter]
		Select branch            [B]
		Filter by branch         [F]
		Dump command             [D]
		Back/Quit              [Esc]
	`
)

func styleForStatus(status string) tcell.Style {
	return tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(colourByStatus[status])
}

func branchOrTag(pipeline circleci.Pipeline) string {
	branchOrTag := pipeline.Vcs.Branch
	if branchOrTag == "" {
		branchOrTag = pipeline.Vcs.Tag
	}

	return branchOrTag
}


func clearAll() {
	logsView.Clear()
	stepNodes := stepsTree.GetRowCount()
	if stepNodes > 0 {
		stepsTree.GetRoot().ClearChildren()
	}
	jobsTable.Clear()
	workflowsTable.Clear()
	pipelinesTable.Clear()
}