package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type stepsPane struct {
	tree        *tview.TreeView
	follow      bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (s stepsPane) watchSteps(cTui *CirclogTui) {
	stepsChan := make(chan circleci.JobDetails)

	for {
		go func() {
			jobDetails, _ := circleci.GetJobSteps(cTui.config, cTui.tuiState.job.JobNumber)
			stepsChan <- jobDetails
		}()

		time.Sleep(refreshInterval)

		select {
		case <-cTui.steps.watchCtx.Done():
			return

		default:
			jobDetails := <-stepsChan
			cTui.app.QueueUpdateDraw(func() {
				currentNode := cTui.steps.tree.GetCurrentNode().GetReference().(circleci.Action)
				cTui.steps.clear()
				cTui.steps.populateStepsTree(cTui.tuiState.job, jobDetails)
				if cTui.steps.follow {
					steps := cTui.steps.tree.GetRoot().GetChildren()
					latestStepActions := steps[len(steps)-1].GetChildren()
					for n := len(latestStepActions) - 1; n >= 0; n-- {
						if n == 0 {
							cTui.steps.tree.SetCurrentNode(latestStepActions[n])
							cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							return
						} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
							cTui.steps.tree.SetCurrentNode(latestStepActions[n])
							cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							return
						}
					}
				}

				for _, step := range cTui.steps.tree.GetRoot().GetChildren() {
					for _, action := range step.GetChildren() {
						node := action.GetReference().(circleci.Action)
						if node.Step == currentNode.Step && node.Index == currentNode.Index {
							cTui.steps.tree.SetCurrentNode(action)
						}
					}
				}
			})
		}
	}
}

func (s stepsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	cTui.steps.watchCancel()
	fn()
	cTui.steps.watchCtx, cTui.steps.watchCancel = context.WithCancel(context.TODO())
	go cTui.steps.watchSteps(cTui)
}

func (s stepsPane) populateStepsTree(job circleci.Job, jobDetails circleci.JobDetails) {
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

func (s stepsPane) clear() {
	stepNodes := s.tree.GetRowCount()
	if stepNodes > 0 {
		s.tree.GetRoot().ClearChildren()
	}
}
