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

	info          *tview.Flex
	projectSelect *tview.InputField
	branchSelect  *tview.InputField
	configText    *tview.TextView
	controls      *tview.TextView

	pipelines pipelinesTable
	workflows workflowsTable
	jobs      jobsTable

	steps stepsTree
	logs  logsView

	controlBindings string
	colourByStatus  map[string]tcell.Color
}

const refreshInterval = 1

var (
	controlBindings = `Move	           [Up/Down]
		Select               [Enter]
		Dump command             [D]
		Back/Quit              [Esc]
	`

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
		config:          config,
		controlBindings: controlBindings,
		colourByStatus:  colourByStatus,
	}
}

func (cTui *CirclogTui) Run() error {
	cTui.app = tview.NewApplication()

	cTui.initNavLayout()

	cTui.pipelines = cTui.newPipelinesTable()
	cTui.upperNav.AddItem(cTui.pipelines.table, 0, 1, false)

	cTui.workflows = cTui.newWorkflowsTable()
	cTui.upperNav.AddItem(cTui.workflows.table, 0, 1, false)

	cTui.jobs = cTui.newJobsTable()
	cTui.upperNav.AddItem(cTui.jobs.table, 0, 1, false)

	cTui.steps = cTui.newStepsTree()
	cTui.lowerNav.AddItem(cTui.steps.tree, 0, 1, false)

	cTui.logs = cTui.newLogsView()
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
	cTui.layout.SetTitle(" circlog ").SetBorder(true).SetBorderPadding(1, 1, 1, 1)

	cTui.heading = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.heading, 6, 0, false)

	cTui.info = tview.NewFlex().SetDirection(tview.FlexRow)
	cTui.heading.AddItem(cTui.info, 0, 1, false)

	cTui.initProjectSelect()
	cTui.info.AddItem(cTui.projectSelect, 1, 1, true)

	cTui.initBranchSelect()
	cTui.info.AddItem(cTui.branchSelect, 1, 1, false)

	cTui.configText = tview.NewTextView().SetText(fmt.Sprintf("Organisation: %s", cTui.config.Org))
	cTui.info.AddItem(cTui.configText, 0, 1, false)

	cTui.controls = tview.NewTextView().SetTextAlign(tview.AlignRight)
	cTui.controls.SetText(cTui.controlBindings)
	cTui.heading.AddItem(cTui.controls, 0, 1, false)

	cTui.upperNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.upperNav, 0, 2, false)

	cTui.lowerNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.lowerNav, 0, 3, false)
}

func (cTui *CirclogTui) refreshPipelinesTable(ctx context.Context) {
	cTui.pipelines.numPages = 1
	pipelinesChan := make(chan []circleci.Pipeline)
	nextPageTokenChan := make(chan string)

	for {
		go func() {
			pipelines, nextPageToken, _ := circleci.GetProjectPipelines(cTui.config, cTui.pipelines.numPages, "")
			pipelinesChan <- pipelines
			nextPageTokenChan <- nextPageToken
		}()

		time.Sleep(time.Second * refreshInterval)

		select {
		case <-ctx.Done():
			return

		default:
			cTui.app.QueueUpdateDraw(func() {
				pipelines := <- pipelinesChan
				nextPageToken := <-nextPageTokenChan
				cTui.pipelines.clear()
				cTui.pipelines.addPipelinesToTable(pipelines, ((cTui.pipelines.numPages-1)*20)+1, nextPageToken)
			})
		}
	}
}

func (cTui *CirclogTui) refreshWorkflowsTable(ctx context.Context) {
	cTui.workflows.numPages = 1
	workflowsChan := make(chan []circleci.Workflow)
	nextPageTokenChan := make(chan string)

	for {
		go func() {
			workflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.tuiState.pipeline.Id, cTui.pipelines.numPages, "")
			workflowsChan <- workflows
			nextPageTokenChan <- nextPageToken
		}()

		time.Sleep(time.Second * refreshInterval)
		
		select {
		case <-ctx.Done():
			return
		
		default:
			cTui.app.QueueUpdateDraw(func() {
				workflows := <- workflowsChan
				nextPageToken := <-nextPageTokenChan
				cTui.workflows.clear()
				cTui.workflows.addWorkflowsToTable(workflows, ((cTui.workflows.numPages-1)*20)+1, nextPageToken)
			})
		}
	}
}

func (cTui *CirclogTui) refreshJobsTable(ctx context.Context) {
	cTui.jobs.numPages = 1
	jobsChan := make(chan []circleci.Job)
	nextPageTokenChan := make(chan string)
	
	for {
		go func() {
			jobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.tuiState.workflow.Id, cTui.workflows.numPages, "")
			jobsChan <- jobs
			nextPageTokenChan <- nextPageToken
		}()
		
		time.Sleep(time.Second * refreshInterval)
		
		select {
		case <-ctx.Done():
			return
		
		default:
			cTui.app.QueueUpdateDraw(func() {	
				jobs := <- jobsChan
				nextPageToken := <-nextPageTokenChan
				cTui.jobs.clear()
				cTui.jobs.addJobsToTable(jobs, ((cTui.jobs.numPages-1)*20)+1, nextPageToken)
			})
		}
	}
}


func (cTui *CirclogTui) refreshStepsTree(ctx context.Context) {
	stepsChan := make(chan circleci.JobDetails)
	
	for {
		go func() {
			jobDetails, _ := circleci.GetJobSteps(cTui.config, cTui.tuiState.job.JobNumber)

			stepsChan <- jobDetails
		}()

		time.Sleep(time.Second * refreshInterval)
		
		select {
		case <-ctx.Done():
			return
		
		default:
			cTui.app.QueueUpdateDraw(func() {	
				jobDetails := <- stepsChan
				currentNodeText := cTui.steps.tree.GetCurrentNode().GetText()
				cTui.steps.clear()
				cTui.steps.populateStepsTree(cTui.tuiState.job, jobDetails)
				for _, step := range cTui.steps.tree.GetRoot().GetChildren() {
					for _, action := range step.GetChildren() {
						if action.GetText() == currentNodeText {
							cTui.steps.tree.SetCurrentNode(action)
						}
					}
				}
			})
		}
	}
}

func (cTui *CirclogTui) refreshLogsView(ctx context.Context) {
	logsChan := make(chan string)
	
	for {
		go func() {
			logs, _ := circleci.GetStepLogs(
			cTui.config,
			cTui.tuiState.job.JobNumber,
			cTui.tuiState.action.Step,
			cTui.tuiState.action.Index,
			cTui.tuiState.action.AllocationId,
			)

			logsChan <- logs
		}()
		
		time.Sleep(time.Second * refreshInterval)

		select {
		case <-ctx.Done():
			return
		
		default:
			logs := <- logsChan
			cTui.app.QueueUpdateDraw(func() {		
				row, col := cTui.logs.view.GetScrollOffset()
				cTui.logs.updateLogsView(logs)
				cTui.logs.view.ScrollTo(row, col)
			})
		}
	}
}
