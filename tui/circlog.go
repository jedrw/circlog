package tui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

type StateComponent interface {
}

type tuiState struct {
	pipeline circleci.Pipeline
	workflow circleci.Workflow
	job      circleci.Job
	action   circleci.Action
}

func (s *tuiState) updater() chan StateComponent {
	updateChan := make(chan StateComponent)

	go func() {
		for update := range updateChan {
			switch component := update.(type) {

			case circleci.Pipeline:
				s.pipeline = component

			case circleci.Workflow:
				s.workflow = component

			case circleci.Job:
				s.job = component

			case circleci.Action:
				s.action = component

			default:
				log.Fatalf("unexpected state type: %T", component)
			}
		}
	}()

	return updateChan
}

type CirclogTui struct {
	app *tview.Application

	config      config.CirclogConfig
	state       tuiState
	updateState chan StateComponent

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

	cTui.updateState = cTui.state.updater()

	if cTui.config.Project != "" {
		cTui.pipelines.update <- true
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
	Back/Quit              [Esc]
	`)
	cTui.heading.AddItem(cTui.globalControls, 0, 1, false)

	cTui.upperNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.upperNav, 0, 2, false)

	cTui.lowerNav = tview.NewFlex().SetDirection(tview.FlexColumn)
	cTui.layout.AddItem(cTui.lowerNav, 0, 3, false)
}
