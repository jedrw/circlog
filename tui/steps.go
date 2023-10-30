package tui

import (
	"fmt"

	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func ShowSteps(config config.CirclogConfig, project string, job circleci.Job, app *tview.Application, layout *tview.Flex) {
	stepsArea := tview.NewFlex()
	stepsArea.SetTitle(fmt.Sprintf(" %s - STEPS ", job.Name)).SetBorder(true)

	jobNode := tview.NewTreeNode(job.Name)
	stepsTree := tview.NewTreeView().
		SetRoot(jobNode).
		SetCurrentNode(jobNode).
		SetGraphics((false)).
		SetTopLevel(1)

	steps, _ := circleci.GetJobSteps(config, project, job.JobNumber)

	for _, step := range steps.Steps {
		stepNode := tview.NewTreeNode(step.Name).
			SetSelectable(false).
			SetReference(step)
		jobNode.AddChild(stepNode)
		for _, action := range step.Actions {
			actionNode := tview.NewTreeNode(fmt.Sprint(action.Index)).
				SetSelectable(true).
				SetReference(action).
				SetColor(ColourByStatus[action.Status])
			stepNode.AddChild(actionNode)
		}
	}

	stepsTree.SetSelectedFunc(func(node *tview.TreeNode) {
		action := node.GetReference()
		layout.RemoveItem(stepsArea)
		ShowLogs(config, project, job.JobNumber, action.(circleci.Action), app, layout)
	})

	stepsArea.AddItem(stepsTree, 0, 1, false)
	layout.AddItem(stepsArea, 0, 1, false)

	app.SetFocus(stepsTree)
}
