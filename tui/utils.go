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
	cTui.watchCancelAll()

	cTui.jobs.numPages = 1
	cTui.workflows.numPages = 1
	cTui.pipelines.numPages = 1

	cTui.logs.view.Clear()
	cTui.steps.clear()
	cTui.jobs.clear()
	cTui.workflows.clear()
	cTui.pipelines.clear()
}

func (cTui *CirclogTui) watchCancelAll() {
	cTui.logs.watchCancel()
	cTui.steps.watchCancel()
	cTui.jobs.watchCancel()
	cTui.workflows.watchCancel()
	cTui.pipelines.watchCancel()
}
