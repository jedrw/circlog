package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type stepsTree struct {
	tree          *tview.TreeView
	refreshCtx    context.Context
	refreshCancel context.CancelFunc
}

func (cTui *CirclogTui) newStepsTree() stepsTree {
	tree := tview.NewTreeView()
	tree.SetTitle(" STEPS ").SetBorder(true)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		cTui.tuiState.action = node.GetReference().(circleci.Action)
		logs, _ := circleci.GetStepLogs(
			cTui.config,
			cTui.tuiState.job.JobNumber,
			cTui.tuiState.action.Step,
			cTui.tuiState.action.Index,
			cTui.tuiState.action.AllocationId,
		)

		cTui.logs.updateLogsView(logs)
		if cTui.logs.autoScroll {
			cTui.logs.view.ScrollToEnd()
		}
		
		cTui.app.SetFocus(cTui.logs.view)
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.steps.refreshCancel()
			tree.GetRoot().ClearChildren()
			cTui.app.SetFocus(cTui.jobs.table)

			return event
		}

		switch event.Rune() {
		case 'b':
			cTui.app.SetFocus(cTui.branchSelect)


		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog steps %s -j %d\n", cTui.config.Project, cTui.tuiState.job.JobNumber)
		}
		
		return event
	})

	tree.SetFocusFunc(func() {
		cTui.steps.refreshCancel()
		cTui.controls.SetText(cTui.controlBindings)
		cTui.steps.refreshCtx, cTui.steps.refreshCancel = context.WithCancel(context.TODO())
		go cTui.refreshStepsTree(cTui.steps.refreshCtx)
	})

	refreshCtx, refreshCancel := context.WithCancel(context.TODO())

	return stepsTree{
		tree: tree,
		refreshCtx: refreshCtx,
		refreshCancel: refreshCancel,
	}
}

func (s stepsTree) populateStepsTree(job circleci.Job, jobDetails circleci.JobDetails) {
	jobNode := tview.NewTreeNode(job.Name)

	s.tree.SetRoot(jobNode).
		SetCurrentNode(jobNode).
		SetGraphics(true).
		SetTopLevel(1)

	if len(jobDetails.Steps) != 0 {
		for i, step := range jobDetails.Steps {
			stepNode := tview.NewTreeNode(step.Name).
				SetSelectable(false).
				SetReference(step)
			jobNode.AddChild(stepNode)
			for _, action := range step.Actions {
				var actionDuration string
				if action.Status == circleci.RUNNING {
					actionDuration = time.Since(action.StartTime).Round(time.Millisecond).String()
				} else {
					if i == len(jobDetails.Steps)-1 {
						actionDuration = job.StoppedAt.Sub(action.StartTime).Round(time.Millisecond).String()
					} else {
						actionDuration = jobDetails.Steps[i+1].Actions[0].StartTime.Sub(action.StartTime).Round(time.Millisecond).String()
					}
				}

				actionNode := tview.NewTreeNode(fmt.Sprintf(" %d (%s)", action.Index, actionDuration)).
					SetSelectable(true).
					SetReference(action).
					SetColor(colourByStatus[action.Status])

				stepNode.AddChild(actionNode)
			}
		}
	} else {
		noneNode := tview.NewTreeNode("None").SetSelectable(false)
		jobNode.AddChild(noneNode)
	}

}

func (s stepsTree) clear() {
	stepNodes := s.tree.GetRowCount()
	s.tree.GetRoot()
	if stepNodes > 0 {
		s.tree.GetRoot().ClearChildren()
	}
}
