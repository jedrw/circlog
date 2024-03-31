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

func (cTui *CirclogTui) newStepsPane() stepsPane {
	tree := tview.NewTreeView()
	tree.SetTitle(" STEPS - Follow Disabled ")
	tree.SetBorder(true)
	tree.SetBorderColor(tcell.ColorGrey)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		cTui.logs.restartWatcher(cTui, func() {
			cTui.steps.restartWatcher(cTui, func() {
				cTui.state.action = node.GetReference().(circleci.Action)
			})
		})

		if cTui.logs.autoScroll {
			cTui.logs.view.ScrollToEnd()
		}

		cTui.app.SetFocus(cTui.logs.view)
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyUp:
			if cTui.steps.follow {
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = false
					cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
				})
			}

		case tcell.KeyEsc:
			cTui.logs.watchCancel()
			cTui.steps.watchCancel()
			cTui.logs.view.Clear()
			cTui.steps.clear()
			tree.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.jobs.table)

			return event

		}

		switch event.Rune() {

		case 'f':
			cTui.logs.restartWatcher(cTui, func() {
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = !cTui.steps.follow
					if cTui.steps.follow {
						cTui.logs.autoScroll = true
						cTui.steps.tree.SetTitle(" STEPS - Follow Enabled ")
						cTui.logs.view.SetTitle(" LOGS - Autoscroll Enabled ")
						steps := tree.GetRoot().GetChildren()
						latestStepActions := steps[len(steps)-1].GetChildren()
						for n := len(latestStepActions) - 1; n >= 0; n-- {
							if n == 0 {
								cTui.steps.tree.SetCurrentNode(latestStepActions[n])
								cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
								cTui.steps.tree.SetCurrentNode(latestStepActions[n])
								cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							}
						}
					} else {
						cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
					}
				})
			})

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
		cTui.steps.restartWatcher(cTui, func() {
			tree.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.SetText(`Toggle Follow            [F]`)
			if cTui.steps.follow {
				steps := cTui.steps.tree.GetRoot().GetChildren()
				latestStepActions := steps[len(steps)-1].GetChildren()
				for n := len(latestStepActions) - 1; n >= 0; n-- {
					if n == 0 {
						cTui.steps.tree.SetCurrentNode(latestStepActions[n])
						cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
					} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
						cTui.steps.tree.SetCurrentNode(latestStepActions[n])
						cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
					}
				}
			}
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	return stepsPane{
		tree:        tree,
		follow:      false,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (s *stepsPane) watchSteps(ctx context.Context, cTui *CirclogTui) {
	stepsChan := make(chan circleci.JobDetails)
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		go func() {
			jobDetails, _ := circleci.GetJobSteps(cTui.config, cTui.state.job.JobNumber)
			stepsChan <- jobDetails
		}()

		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			jobDetails := <-stepsChan
			cTui.app.QueueUpdateDraw(func() {
				if s.tree.GetRoot().GetChildren()[0].GetText() != "None" {
					currentNode := s.tree.GetCurrentNode().GetReference().(circleci.Action)
					s.clear()
					s.populateStepsTree(cTui.state.job, jobDetails)
					if s.follow {
						steps := s.tree.GetRoot().GetChildren()
						latestStepActions := steps[len(steps)-1].GetChildren()
						for n := len(latestStepActions) - 1; n >= 0; n-- {
							if n == 0 {
								cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
								s.tree.SetCurrentNode(latestStepActions[n])
								cTui.logs.restartWatcher(cTui, func() {})
								return
							} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
								cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
								s.tree.SetCurrentNode(latestStepActions[n])
								cTui.logs.restartWatcher(cTui, func() {})
								return
							}
						}
					}

					for _, step := range s.tree.GetRoot().GetChildren() {
						for _, action := range step.GetChildren() {
							node := action.GetReference().(circleci.Action)
							if node.Step == currentNode.Step && node.Index == currentNode.Index {
								s.tree.SetCurrentNode(action)
							}
						}
					}
				}
			})
		}
	}
}

func (s *stepsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	s.watchCancel()
	fn()
	s.watchCtx, s.watchCancel = context.WithCancel(context.TODO())
	go s.watchSteps(s.watchCtx, cTui)
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
