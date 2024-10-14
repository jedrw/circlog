package tui

import (
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
