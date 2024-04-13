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
	update      chan bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (cTui *CirclogTui) newStepsPane() stepsPane {
	tree := tview.NewTreeView()
	tree.SetTitle(" STEPS - Follow Disabled ")
	tree.SetBorder(true)
	tree.SetBorderColor(tcell.ColorGrey)
	// tree.SetChangedFunc(func(node *tview.TreeNode) {
	// 	currentAction := node.GetReference().(circleci.Action)
	// 	if cTui.state.action != currentAction {
	// 		cTui.updateState <- currentAction
	// 	}

	// 	cTui.logs.update <- cTui.state.action
	// })

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		currentAction := node.GetReference().(circleci.Action)
		if cTui.state.action != currentAction {
			cTui.updateState <- currentAction
		}

		cTui.logs.update <- true
		cTui.app.SetFocus(cTui.logs.view)
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyUp:
			if cTui.steps.follow {
				cTui.steps.follow = false
				cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
			}

		case tcell.KeyEsc:
			cTui.steps.watchCancel()
			cTui.logs.view.Clear()
			cTui.steps.clear()
			tree.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.jobs.table)

			return event
		}

		switch event.Rune() {

		case 'f':
			cTui.steps.toggleFollow(cTui)
			cTui.app.SetFocus(cTui.logs.view)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog steps %s -j %d\n", cTui.config.Project, cTui.state.job.JobNumber)
		}

		return event
	})

	tree.SetFocusFunc(func() {
		cTui.steps.restartWatcher(func() {
			tree.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.SetText("Toggle Follow            [F]")
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	stepsPane := stepsPane{
		tree:        tree,
		follow:      false,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}

	stepsPane.update = stepsPane.updater(cTui)

	return stepsPane
}

func (s *stepsPane) watchSteps(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			s.update <- true
		}
	}
}

func (s *stepsPane) restartWatcher(fn func()) {
	s.watchCancel()
	fn()
	s.watchCtx, s.watchCancel = context.WithCancel(context.TODO())
	go s.watchSteps(s.watchCtx)
}

func (s *stepsPane) updater(cTui *CirclogTui) chan bool {
	updateChan := make(chan bool)

	go func() {
		for <-updateChan {
			jobDetails, _ := circleci.GetJobSteps(cTui.config, cTui.state.job.JobNumber)
			cTui.app.QueueUpdateDraw(func() {
				if len(cTui.steps.tree.GetRoot().GetChildren()) > 0 {
					if cTui.steps.tree.GetRoot().GetChildren()[0].GetText() != "None" {
						currentNode := cTui.steps.tree.GetCurrentNode().GetReference().(circleci.Action)
						cTui.steps.clear()
						cTui.steps.populateStepsTree(cTui.state.job, jobDetails)
						if cTui.steps.follow {
							steps := cTui.steps.tree.GetRoot().GetChildren()
							latestStepActions := steps[len(steps)-1].GetChildren()
							for n := len(latestStepActions) - 1; n >= 0; n-- {
								if n == 0 {
									cTui.steps.tree.SetCurrentNode(latestStepActions[n])
									currentAction := latestStepActions[n].GetReference().(circleci.Action)
									cTui.updateState <- currentAction
								} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
									cTui.steps.tree.SetCurrentNode(latestStepActions[n])
									currentAction := latestStepActions[n].GetReference().(circleci.Action)
									cTui.updateState <- currentAction
								}
							}
						} else {
							for _, step := range cTui.steps.tree.GetRoot().GetChildren() {
								for _, action := range step.GetChildren() {
									node := action.GetReference().(circleci.Action)
									if node.Step == currentNode.Step && node.Index == currentNode.Index {
										cTui.steps.tree.SetCurrentNode(action)
									}
								}
							}
						}
					}
				}
			})
		}
	}()

	return updateChan
}

func (s *stepsPane) populateStepsTree(job circleci.Job, jobDetails circleci.JobDetails) {
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

func (s *stepsPane) clear() {
	stepNodes := s.tree.GetRowCount()
	if stepNodes > 0 {
		s.tree.GetRoot().ClearChildren()
	}
}

func (s *stepsPane) toggleFollow(cTui *CirclogTui) {
	if !cTui.steps.follow {
		cTui.steps.follow = true
		cTui.logs.autoScroll = true
		cTui.steps.tree.SetTitle(" STEPS - Follow Enabled ")
		cTui.logs.view.SetTitle(" LOGS - Autoscroll Enabled ")
		steps := cTui.steps.tree.GetRoot().GetChildren()
		latestStepActions := steps[len(steps)-1].GetChildren()
		for n := len(latestStepActions) - 1; n >= 0; n-- {
			if n == 0 {
				cTui.steps.tree.SetCurrentNode(latestStepActions[n])
				currentAction := latestStepActions[n].GetReference().(circleci.Action)
				cTui.updateState <- currentAction
				cTui.logs.update <- true
			} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
				cTui.steps.tree.SetCurrentNode(latestStepActions[n])
				currentAction := latestStepActions[n].GetReference().(circleci.Action)
				cTui.updateState <- currentAction
				cTui.logs.update <- true
			}
		}
	} else {
		cTui.steps.follow = false
		cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
	}
}
