package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
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

func (cTui *CirclogTui) clearAll() {
	cTui.logs.view.Clear()
	stepNodes := cTui.steps.tree.GetRowCount()
	if stepNodes > 0 {
		cTui.steps.tree.GetRoot().ClearChildren()
	}
	cTui.jobs.clear()
	cTui.workflows.clear()
	cTui.pipelines.clear()
}