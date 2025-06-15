package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
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
	cTui.state.mu.Lock()
	cTui.state.jobPages = 1
	cTui.state.workflowPages = 1
	cTui.state.pipelinePages = 1
	cTui.state.mu.Unlock()

	cTui.logsPane.Clear()
	cTui.clearStepsTree()
	cTui.clearJobsTable()
	cTui.clearWorkflowsTable()
	cTui.clearPipelinesTable()
}
