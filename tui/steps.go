package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newStepsTree() *tview.TreeView {
	stepsTree := tview.NewTreeView()
	stepsTree.SetTitle(" STEPS ").SetBorder(true)

	return stepsTree
}

func updateStepsTree(config config.CirclogConfig, project string, job circleci.Job) {
	jobNode := tview.NewTreeNode(job.Name)

	stepsTree.SetRoot(jobNode).
		SetCurrentNode(jobNode).
		SetGraphics(true).
		SetTopLevel(1)

	stepsTree.SetSelectedFunc(func(node *tview.TreeNode) {
		action := node.GetReference().(circleci.Action)
		updateLogsView(config, project, job, action, logsView)
	})

	steps, _ := circleci.GetJobSteps(config, job.JobNumber)

	if len(steps.Steps) != 0 {
		for i, step := range steps.Steps {
			stepNode := tview.NewTreeNode(step.Name).
				SetSelectable(false).
				SetReference(step)
			jobNode.AddChild(stepNode)
			for _, action := range step.Actions {
				var actionDuration string
				if i == len(steps.Steps)-1 {
					actionDuration = job.StoppedAt.Sub(action.StartTime).Round(time.Millisecond).String()
				} else {
					actionDuration = steps.Steps[i+1].Actions[0].StartTime.Sub(action.StartTime).Round(time.Millisecond).String()
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

	stepsTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyBackspace2 {
			stepsTree.GetRoot().ClearChildren()
			app.SetFocus(jobsTable)
		}

		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog steps %s -j %d\n", project, job.JobNumber)
		}

		return event
	})

	app.SetFocus(stepsTree)
}
