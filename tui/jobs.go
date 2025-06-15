package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jedrw/circlog/circleci"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) newJobsPane() *tview.Table {
	table := tview.NewTable()
	table.SetTitle(" JOBS ")
	table.SetBackgroundColor(tcell.ColorDefault)
	table.SetBorder(true)
	table.SetBorderColor(tcell.ColorGrey)
	table.SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)

	for column, header := range []string{"Name", "Duration", "Depends on"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, _ int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {

		case circleci.Job:
			cTui.state.mu.Lock()
			cTui.state.job = cellRef
			jobDetails, _ := circleci.GetSteps(cTui.config, cTui.state.job.JobNumber)
			cTui.state.mu.Unlock()
			cTui.populateStepsTree(cTui.state.job, jobDetails)
			cTui.app.SetFocus(cTui.stepsPane)

		case string:
			if cell.Text == "..." {
				cTui.state.mu.Lock()
				cTui.state.jobPages++
				nextPageToken := cell.GetReference().(string)
				newJobs, nextPageToken, _ := circleci.GetJobs(cTui.config, cTui.state.workflow.Id, 1, nextPageToken)
				cTui.state.mu.Unlock()
				cTui.addJobsToTable(newJobs, table.GetRowCount()-1, nextPageToken)
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.clearJobsTable()
			table.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.workflowsPane)

			return event
		}

		switch event.Rune() {

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog jobs %s -w %s\n", cTui.config.Project, cTui.state.workflow.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		table.SetBorderColor(tcell.ColorDefault)
		cTui.paneControls.Clear()
	})

	return table
}

func (cTui *CirclogTui) populateJobsTable(jobs []circleci.Job, nextPageToken string) {
	cTui.clearJobsTable()
	cTui.addJobsToTable(jobs, cTui.jobsPane.GetRowCount(), nextPageToken)
}

func (cTui *CirclogTui) addJobsToTable(jobs []circleci.Job, startRow int, nextPageToken string) {
	if len(jobs) != 0 {
		for row, job := range jobs {
			dependencies := getNamedJobDependencies(job, jobs)
			var dependenciesString string
			if len(dependencies) == 0 {
				dependenciesString = "[]"
			} else {
				// tview has the concept of Regions which use "[string]" as identifiers,
				// to escape these we must add a "[" before the closing "]"
				dependenciesString = fmt.Sprintf("[%s[]", strings.Join(dependencies, ", "))
			}

			var jobDuration string
			if job.Status == circleci.RUNNING {
				jobDuration = time.Since(job.StartedAt).Round(time.Millisecond).String()
			} else {
				jobDuration = job.StoppedAt.Sub(job.StartedAt).Round(time.Millisecond).String()
			}

			for column, attr := range []string{job.Name, jobDuration, dependenciesString} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(job.Status))
				cell.SetReference(job)
				cTui.jobsPane.SetCell(row+startRow, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...").SetStyle(tcell.StyleDefault)
			cell.SetReference(nextPageToken)
			cTui.jobsPane.SetCell(cTui.jobsPane.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		cTui.jobsPane.SetCell(1, 0, cell)
	}
}

func getNamedJobDependencies(job circleci.Job, jobs []circleci.Job) []string {
	var namedDependencies []string
	for _, dependsOnJobId := range job.Dependencies {
		for _, requiredJob := range jobs {
			if requiredJob.Id == dependsOnJobId {
				namedDependencies = append(namedDependencies, requiredJob.Name)
			}
		}
	}

	return namedDependencies
}

func (cTui *CirclogTui) clearJobsTable() {
	row := 1
	for row < cTui.jobsPane.GetRowCount() {
		cTui.jobsPane.RemoveRow(row)
	}
}
