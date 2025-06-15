package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/jedrw/circlog/config"
	"github.com/rivo/tview"
)

type tuiState struct {
	mu            sync.RWMutex
	follow        bool
	autoScroll    bool
	pipeline      circleci.Pipeline
	pipelinePages int
	workflow      circleci.Workflow
	workflowPages int
	job           circleci.Job
	jobPages      int
	action        circleci.Action
}

type CirclogTui struct {
	app    *tview.Application
	config config.CirclogConfig
	state  tuiState

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

	pipelinesPane *tview.Table
	workflowsPane *tview.Table
	jobsPane      *tview.Table
	stepsPane     *tview.TreeView
	logsPane      *tview.TextView

	colourByStatus map[string]tcell.Color
}

var (
	colourByStatus = map[string]tcell.Color{
		"success":      tcell.ColorDarkGreen,
		"running":      tcell.ColorLightGreen,
		"not_run":      tcell.ColorGray,
		"not_running":  tcell.ColorGray,
		"blocked":      tcell.ColorGray,
		"failed":       tcell.ColorDarkRed,
		"error":        tcell.ColorDarkRed,
		"failing":      tcell.ColorPink,
		"on_hold":      tcell.ColorYellow,
		"canceled":     tcell.ColorDarkRed,
		"unauthorized": tcell.ColorDarkRed,

		"created": tcell.ColorDarkGreen,
	}
)

func NewCirclogTui(config config.CirclogConfig) CirclogTui {
	return CirclogTui{
		config:         config,
		colourByStatus: colourByStatus,
		state: tuiState{
			follow:        false,
			autoScroll:    true,
			pipelinePages: 1,
			workflowPages: 1,
			jobPages:      1,
		},
	}
}

func (cTui *CirclogTui) initNavLayout() {
	if cTui.layout != nil {
		return
	}

	cTui.layout = tview.NewFlex().SetDirection(tview.FlexRow)
	cTui.layout.SetTitle(" circlog ").SetBorder(true).SetBorderPadding(1, 0, 1, 1)
	cTui.layout.SetBackgroundColor(tcell.ColorDefault)

	cTui.heading = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.heading.SetBackgroundColor(tcell.ColorDefault)
	cTui.layout.AddItem(cTui.heading, 5, 0, false)

	cTui.info = tview.NewFlex().SetDirection(tview.FlexRow)
	cTui.info.SetBackgroundColor(tcell.ColorDefault)
	cTui.heading.AddItem(cTui.info, 0, 1, false)

	cTui.initProjectSelect()
	cTui.info.AddItem(cTui.projectSelect, 1, 1, true)

	cTui.initBranchSelect()
	cTui.info.AddItem(cTui.branchSelect, 1, 1, false)

	cTui.configText = tview.NewTextView().SetText(fmt.Sprintf("Organisation: %s", cTui.config.Org))
	cTui.configText.SetBackgroundColor(tcell.ColorDefault)
	cTui.info.AddItem(cTui.configText, 0, 1, false)

	cTui.paneControls = tview.NewTextView().SetTextAlign(tview.AlignRight)
	cTui.paneControls.SetBackgroundColor(tcell.ColorDefault)
	cTui.heading.AddItem(cTui.paneControls, 0, 4, false)

	cTui.globalControls = tview.NewTextView().SetTextAlign(tview.AlignRight)
	cTui.globalControls.SetBackgroundColor(tcell.ColorDefault)
	cTui.globalControls.SetText("Move   \t[Up/Down]\nSelect   \t[Enter]\nDump command \t[D]\nBranch Select\t[B]\nBack/Quit  \t[Esc]")
	cTui.heading.AddItem(cTui.globalControls, 0, 1, false)

	cTui.upperNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.upperNav.SetBackgroundColor(tcell.ColorDefault)
	cTui.layout.AddItem(cTui.upperNav, 0, 2, false)

	cTui.lowerNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.lowerNav.SetBackgroundColor(tcell.ColorDefault)
	cTui.layout.AddItem(cTui.lowerNav, 0, 3, false)
}

func (cTui *CirclogTui) Refresh() {
	// TODO update everything
	cTui.state.mu.Lock()
	pipelines, pipelinesNextPageToken, _ := circleci.GetPipelines(cTui.config, cTui.state.pipelinePages, "")
	workflows, workflowsNextPageToken, _ := circleci.GetWorkflows(cTui.config, cTui.state.pipeline.Id, cTui.state.workflowPages, "")
	jobs, jobsNextPageToken, _ := circleci.GetJobs(cTui.config, cTui.state.workflow.Id, cTui.state.jobPages, "")
	jobDetails, _ := circleci.GetSteps(cTui.config, cTui.state.job.JobNumber)
	cTui.state.mu.Unlock()
	// drawWorkflowsTable := cTui.workflowsPane.refresh(cTui)
	// drawJobsTable := cTui.jobsPane.refresh(cTui)

	cTui.app.QueueUpdateDraw(func() {
		cTui.populatePipelinesTable(pipelines, pipelinesNextPageToken)
		cTui.populateWorkflowsTable(workflows, workflowsNextPageToken)
		cTui.populateJobsTable(jobs, jobsNextPageToken)
		cTui.state.mu.RLock()
		cTui.populateStepsTree(cTui.state.job, jobDetails)
		cTui.state.mu.RUnlock()
		cTui.upateStepsTree(jobDetails)
	})
}

func (cTui *CirclogTui) Run() error {
	cTui.app = tview.NewApplication()

	cTui.initNavLayout()

	cTui.pipelinesPane = cTui.newPipelinesPane()
	cTui.upperNav.AddItem(cTui.pipelinesPane, 0, 1, false)

	cTui.workflowsPane = cTui.newWorkflowsPane()
	cTui.upperNav.AddItem(cTui.workflowsPane, 0, 1, false)

	cTui.jobsPane = cTui.newJobsPane()
	cTui.upperNav.AddItem(cTui.jobsPane, 0, 1, false)

	cTui.stepsPane = cTui.newStepsPane()
	cTui.lowerNav.AddItem(cTui.stepsPane, 0, 1, false)

	cTui.logsPane = cTui.newLogsPane()
	cTui.lowerNav.AddItem(cTui.logsPane, 0, 2, false)

	if cTui.config.Project != "" {
		pipelines, nextPageToken, _ := circleci.GetPipelines(cTui.config, 1, "")
		cTui.populatePipelinesTable(pipelines, nextPageToken)
		cTui.app.SetRoot(cTui.layout, true).SetFocus(cTui.pipelinesPane)
	} else {
		cTui.app.SetRoot(cTui.layout, true).SetFocus(cTui.info)
	}

	ticker := time.NewTicker(cTui.config.RefreshInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cTui.Refresh()
			}
		}
	}()

	return cTui.app.Run()
}
