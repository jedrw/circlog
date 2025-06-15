package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) newStepsPane() *tview.TreeView {
	tree := tview.NewTreeView()
	tree.SetTitle(" STEPS - Follow Disabled ")
	tree.SetBackgroundColor(tcell.ColorDefault)
	tree.SetBorder(true)
	tree.SetBorderColor(tcell.ColorGrey)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		cTui.state.mu.Lock()
		cTui.state.action = node.GetReference().(circleci.Action)
		cTui.state.mu.Unlock()

		cTui.app.SetFocus(cTui.logsPane)
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyUp:
			if cTui.state.follow {
				cTui.state.mu.Lock()
				cTui.state.follow = false
				cTui.state.mu.Unlock()
				cTui.stepsPane.SetTitle(" STEPS - Follow Disabled ")
			}

		case tcell.KeyEsc:
			cTui.logsPane.Clear()
			cTui.clearStepsTree()
			tree.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.jobsPane)

			return event

		}

		switch event.Rune() {

		case 'f':
			toggleFollow(cTui)
			cTui.app.SetFocus(cTui.logsPane)

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog steps %s -j %d\n", cTui.config.Project, cTui.state.job.JobNumber)
		}

		return event
	})

	tree.SetFocusFunc(func() {
		tree.SetBorderColor(tcell.ColorDefault)
		cTui.paneControls.SetText("Toggle Follow\t[F]")
	})

	return tree
}

// TODO

func (cTui *CirclogTui) upateStepsTree(jobDetails circleci.JobDetails) {
	if cTui.stepsPane.GetRoot() != nil {
		if len(cTui.stepsPane.GetRoot().GetChildren()) > 0 {
			if cTui.stepsPane.GetRoot().GetChildren()[0].GetText() != "None" {
				currentNode := cTui.stepsPane.GetCurrentNode().GetReference().(circleci.Action)
				cTui.clearStepsTree()
				cTui.populateStepsTree(cTui.state.job, jobDetails)
				cTui.state.mu.Lock()
				if cTui.state.follow {
					steps := cTui.stepsPane.GetRoot().GetChildren()
					latestStepActions := steps[len(steps)-1].GetChildren()
					for n := len(latestStepActions) - 1; n >= 0; n-- {
						if n == 0 {
							cTui.stepsPane.SetCurrentNode(latestStepActions[n])
							cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
						} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
							cTui.stepsPane.SetCurrentNode(latestStepActions[n])
							cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
						}
					}
				} else {
					for _, step := range cTui.stepsPane.GetRoot().GetChildren() {
						for _, action := range step.GetChildren() {
							node := action.GetReference().(circleci.Action)
							if node.Step == currentNode.Step && node.Index == currentNode.Index {
								cTui.stepsPane.SetCurrentNode(action)
							}
						}
					}
				}
				cTui.state.mu.Unlock()
			}
		}
	}
}

func (cTui *CirclogTui) populateStepsTree(job circleci.Job, jobDetails circleci.JobDetails) {
	jobNode := tview.NewTreeNode(job.Name)

	cTui.stepsPane.SetRoot(jobNode).
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

func (cTui *CirclogTui) clearStepsTree() {
	stepNodes := cTui.stepsPane.GetRowCount()
	if stepNodes > 0 {
		cTui.stepsPane.GetRoot().ClearChildren()
	}
}
