package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

type tuiState struct {
	pipeline circleci.Pipeline
	workflow circleci.Workflow
	job      circleci.Job
	action   circleci.Action
}

type CirclogTui struct {
	app *tview.Application

	config   config.CirclogConfig
	tuiState tuiState

	layout   *tview.Flex
	heading  *tview.Flex
	upperNav *tview.Flex
	lowerNav *tview.Flex

	info           *tview.Flex
	projectSelect  *tview.InputField
	branchSelect   *tview.InputField
	configText     *tview.TextView
	paneControls   *tview.TextView
	globalControls *tview.TextView

	pipelines pipelinesPane
	workflows workflowsPane
	jobs      jobsPane
	steps     stepsPane
	logs      logsPane

	colourByStatus map[string]tcell.Color
}

const refreshInterval = 1 * time.Second

var (
	colourByStatus = map[string]tcell.Color{
		"success":      tcell.ColorDarkGreen,
		"running":      tcell.ColorLightGreen,
		"not_run":      tcell.ColorGray,
		"failed":       tcell.ColorDarkRed,
		"error":        tcell.ColorDarkRed,
		"failing":      tcell.ColorPink,
		"on_hold":      tcell.ColorYellow,
		"canceled":     tcell.ColorDarkRed,
		"unauthorized": tcell.ColorDarkRed,

		"created": tcell.ColorLightGreen,
	}
)

func NewCirclogTui(config config.CirclogConfig) CirclogTui {
	return CirclogTui{
		config:         config,
		colourByStatus: colourByStatus,
	}
}

func (cTui *CirclogTui) Run() error {
	cTui.app = tview.NewApplication()

	cTui.initNavLayout()

	cTui.pipelines = cTui.newPipelinesPane()
	cTui.upperNav.AddItem(cTui.pipelines.table, 0, 1, false)

	cTui.workflows = cTui.newWorkflowsPane()
	cTui.upperNav.AddItem(cTui.workflows.table, 0, 1, false)

	cTui.jobs = cTui.newJobsPane()
	cTui.upperNav.AddItem(cTui.jobs.table, 0, 1, false)

	cTui.steps = cTui.newStepsPane()
	cTui.lowerNav.AddItem(cTui.steps.tree, 0, 1, false)

	cTui.logs = cTui.newLogsPane()
	cTui.lowerNav.AddItem(cTui.logs.view, 0, 2, false)

	if cTui.config.Project != "" {
		pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, "")
		cTui.pipelines.populateTable(pipelines, nextPageToken)
		cTui.app.SetRoot(cTui.layout, true).SetFocus(cTui.pipelines.table)
	} else {
		cTui.app.SetRoot(cTui.layout, true).SetFocus(cTui.info)

	}

	return cTui.app.Run()
}

func (cTui *CirclogTui) initNavLayout() {
	if cTui.layout != nil {
		return
	}

	cTui.layout = tview.NewFlex().SetDirection(tview.FlexRow)
	cTui.layout.SetTitle(" circlog ").SetBorder(true).SetBorderPadding(1, 0, 1, 1)
	cTui.layout.SetBackgroundColor(tcell.ColorDefault)

	cTui.heading = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.heading, 5, 0, false)

	cTui.info = tview.NewFlex().SetDirection(tview.FlexRow)
	cTui.heading.AddItem(cTui.info, 0, 1, false)

	cTui.initProjectSelect()
	cTui.info.AddItem(cTui.projectSelect, 1, 1, true)

	cTui.initBranchSelect()
	cTui.info.AddItem(cTui.branchSelect, 1, 1, false)

	cTui.configText = tview.NewTextView().SetText(fmt.Sprintf("Organisation: %s", cTui.config.Org))
	cTui.info.AddItem(cTui.configText, 0, 1, false)

	cTui.paneControls = tview.NewTextView().SetTextAlign(tview.AlignRight)
	cTui.heading.AddItem(cTui.paneControls, 0, 4, false)

	cTui.globalControls = tview.NewTextView().SetTextAlign(tview.AlignRight)
	cTui.globalControls.SetText(`Move	           [Up/Down]
	Select               [Enter]
	Dump command             [D]
	Branch Select            [B]
	Back/Quit              [Esc]
	`)
	cTui.heading.AddItem(cTui.globalControls, 0, 1, false)

	cTui.upperNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.upperNav, 0, 2, false)

	cTui.lowerNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.lowerNav, 0, 3, false)
}

func (cTui *CirclogTui) newPipelinesPane() pipelinesPane {
	table := tview.NewTable()
	table.SetTitle(" PIPELINES ")
	table.SetBorder(true)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Number", "Branch/Tag", "Start", "Trigger"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Pipeline:
			cTui.tuiState.pipeline = cellRef
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.tuiState.pipeline.Id, 1, "")
			cTui.workflows.populateWorkflowsTable(workflows, nextPageToken)
			cTui.app.SetFocus(cTui.workflows.table)

		case string:
			if cell.Text == "..." {
				cTui.pipelines.restartWatcher(cTui, func() {
					nextPageToken := cell.GetReference().(string)
					newPipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, nextPageToken)
					cTui.pipelines.addPipelinesToTable(newPipelines, table.GetRowCount()-1, nextPageToken)
					cTui.pipelines.numPages++
				})
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.clearAll()
			cTui.config.Project = ""
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.projectSelect)

			return event
		}

		switch event.Rune() {

		case 'v':
			cell := table.GetCell(table.GetSelection())
			cellRef := cell.GetReference()
			switch cellRef := cellRef.(type) {
			case circleci.Pipeline:
				if cellRef.Vcs.Branch != "" {
					cTui.pipelines.restartWatcher(cTui, func() {
						cTui.pipelines.clear()
						cTui.pipelines.numPages = 1
						cTui.config.Branch = cellRef.Vcs.Branch
						cTui.branchSelect.SetText(cTui.config.Branch)
						pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, 1, "")
						cTui.pipelines.populateTable(pipelines, nextPageToken)
						cTui.pipelines.table.ScrollToBeginning()
					})
				}
			}

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.watchCancelAll()
			cTui.app.Stop()
			fmt.Printf("circlog pipelines %s\n", cTui.config.Project)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.pipelines.restartWatcher(cTui, func() {
			cTui.paneControls.SetText(`Filter by branch         [V]`)
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.TODO())

	return pipelinesPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (cTui *CirclogTui) newWorkflowsPane() workflowsPane {
	table := tview.NewTable()
	table.SetTitle(" WORKFLOWS ")
	table.SetBorder(true)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Name", "Duration"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Workflow:
			cTui.tuiState.workflow = cellRef
			jobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.tuiState.workflow.Id, 1, "")
			cTui.jobs.populateTable(jobs, nextPageToken)
			cTui.app.SetFocus(cTui.jobs.table)

		case string:
			if cell.Text == "..." {
				cTui.workflows.restartWatcher(cTui, func() {
					nextPageToken := cell.GetReference().(string)
					newWorkflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.tuiState.pipeline.Id, 1, nextPageToken)
					cTui.workflows.addWorkflowsToTable(newWorkflows, table.GetRowCount()-1, nextPageToken)
					cTui.workflows.numPages++
				})
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.workflows.watchCancel()
			cTui.workflows.clear()
			cTui.app.SetFocus(cTui.pipelines.table)

			return event
		}

		switch event.Rune() {

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog workflows %s -l %s\n", cTui.config.Project, cTui.tuiState.pipeline.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.workflows.restartWatcher(cTui, func() {
			cTui.paneControls.Clear()
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.TODO())

	return workflowsPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (cTui *CirclogTui) newJobsPane() jobsPane {
	table := tview.NewTable()
	table.SetTitle(" JOBS ")
	table.SetBorder(true)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Name", "Duration", "Depends on"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Job:
			cTui.tuiState.job = cellRef
			jobDetails, _ := circleci.GetJobSteps(cTui.config, cTui.tuiState.job.JobNumber)
			cTui.steps.populateStepsTree(cTui.tuiState.job, jobDetails)
			cTui.app.SetFocus(cTui.steps.tree)

		case string:
			if cell.Text == "..." {
				cTui.jobs.restartWatcher(cTui, func() {
					nextPageToken := cell.GetReference().(string)
					newJobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.tuiState.workflow.Id, 1, nextPageToken)
					cTui.jobs.addJobsToTable(newJobs, table.GetRowCount()-1, nextPageToken)
					cTui.jobs.numPages++
				})
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.jobs.watchCancel()
			cTui.jobs.clear()
			cTui.app.SetFocus(cTui.workflows.table)

			return event
		}

		switch event.Rune() {

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog jobs %s -w %s\n", cTui.config.Project, cTui.tuiState.workflow.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.jobs.restartWatcher(cTui, func() {
			cTui.paneControls.Clear()
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.TODO())

	return jobsPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (cTui *CirclogTui) newStepsPane() stepsPane {
	tree := tview.NewTreeView()
	tree.SetTitle(" STEPS - Follow Enabled [F[] ")
	tree.SetBorder(true)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		cTui.logs.restartWatcher(cTui, func() {
			cTui.steps.restartWatcher(cTui, func() {
				cTui.tuiState.action = node.GetReference().(circleci.Action)
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
			fmt.Printf("circlog steps %s -j %d\n", cTui.config.Project, cTui.tuiState.job.JobNumber)
		}

		return event
	})

	tree.SetFocusFunc(func() {
		cTui.steps.restartWatcher(cTui, func() {
			cTui.paneControls.SetText(`Toggle Follow            [F]`)
			if cTui.steps.follow {
				steps := cTui.steps.tree.GetRoot().GetChildren()
				latestStepActions := steps[len(steps)-1].GetChildren()
				for n := len(latestStepActions) - 1; n >= 0; n-- {
					if n == 0 {
						cTui.steps.tree.SetCurrentNode(latestStepActions[n])
					} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
						cTui.steps.tree.SetCurrentNode(latestStepActions[n])
					}
				}
			}
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.TODO())

	return stepsPane{
		tree:        tree,
		follow:      true,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (cTui *CirclogTui) newLogsPane() logsPane {
	view := tview.NewTextView()
	view.SetTitle(" LOGS - Autoscroll Enabled ")
	view.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	view.SetDynamicColors(true)

	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyEsc:
			cTui.logs.watchCancel()
			cTui.logs.view.Clear()
			if cTui.steps.follow {
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = !cTui.steps.follow
					cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
				})
			}

			cTui.app.SetFocus(cTui.steps.tree)

			return event

		case tcell.KeyUp:
			cTui.logs.restartWatcher(cTui, func() {
				cTui.logs.autoScroll = false
				view.SetTitle(" LOGS - Autoscroll Disabled ")
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = false
					cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
				})
			})

			return event
		}

		switch event.Rune() {

		case 'f':
			cTui.logs.restartWatcher(cTui, func() {
				cTui.logs.autoScroll = true
				cTui.logs.view.SetTitle(" LOGS - Autoscroll Enabled ")
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = !cTui.steps.follow
					if cTui.steps.follow {
						cTui.steps.tree.SetTitle(" STEPS - Follow Enabled ")
						steps := cTui.steps.tree.GetRoot().GetChildren()
						latestStepActions := steps[len(steps)-1].GetChildren()
						for n := len(latestStepActions) - 1; n >= 0; n-- {
							if n == 0 {
								cTui.steps.tree.SetCurrentNode(latestStepActions[n])
								cTui.app.SetFocus(cTui.steps.tree)
								cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
								cTui.steps.tree.SetCurrentNode(latestStepActions[n])
								cTui.app.SetFocus(cTui.steps.tree)
								cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							}
						}
					} else {
						cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
					}
				})
			})

		case 'a':
			cTui.logs.restartWatcher(cTui, func() {
				cTui.logs.autoScroll = !cTui.logs.autoScroll
				cTui.steps.restartWatcher(cTui, func() {
					if cTui.logs.autoScroll {
						view.SetTitle(" LOGS - Autoscroll Enabled ")
						view.ScrollToEnd()
					} else {
						cTui.steps.follow = false
						cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
						view.SetTitle(" LOGS - Autoscroll Disabled ")
					}
				})
			})

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog logs %s -j %d -s %d -i %d -a \"%s\"\n",
				cTui.config.Project,
				cTui.tuiState.job.JobNumber,
				cTui.tuiState.action.Step,
				cTui.tuiState.action.Index,
				cTui.tuiState.action.AllocationId,
			)
		}

		return event
	})

	view.SetFocusFunc(func() {
		cTui.logs.restartWatcher(cTui, func() {
			cTui.paneControls.SetText(`Toggle Autoscroll        [A]`)
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.TODO())

	return logsPane{
		view:        view,
		autoScroll:  true,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}
