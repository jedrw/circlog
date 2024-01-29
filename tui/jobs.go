package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type jobsTable struct {
	table         *tview.Table
	numPages      int
	refreshCtx    context.Context
	refreshCancel context.CancelFunc
}

func (cTui *CirclogTui) newJobsTable() jobsTable {
	table := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	table.SetTitle(" JOBS ").SetBorder(true)

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
				cTui.jobs.refreshCancel()
				nextPageToken := cell.GetReference().(string)
				newJobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.tuiState.workflow.Id, 1, nextPageToken)
				cTui.jobs.addJobsToTable(newJobs, table.GetRowCount()-1, nextPageToken)
				cTui.jobs.numPages++
				cTui.jobs.refreshCtx, cTui.jobs.refreshCancel = context.WithCancel(context.TODO())
				go cTui.refreshJobsTable(cTui.jobs.refreshCtx)
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.jobs.refreshCancel()
			cTui.jobs.clear()
			cTui.app.SetFocus(cTui.workflows.table)

			return event
		}

		switch event.Rune() {
		case 'b':
			cTui.jobs.numPages = 1
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog jobs %s -w %s\n", cTui.config.Project, cTui.tuiState.workflow.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.jobs.refreshCancel()
		cTui.controls.SetText(cTui.controlBindings)
		cTui.jobs.refreshCtx, cTui.jobs.refreshCancel = context.WithCancel(context.TODO())
		go cTui.refreshJobsTable(cTui.jobs.refreshCtx)
	})

	refreshCtx, refreshCancel := context.WithCancel(context.TODO())

	return jobsTable{
		table:         table,
		numPages:      1,
		refreshCtx:    refreshCtx,
		refreshCancel: refreshCancel,
	}
}

func (j jobsTable) populateTable(jobs []circleci.Job, nextPageToken string) {
	j.clear()
	j.addJobsToTable(jobs, j.table.GetRowCount(), nextPageToken)

}

func (j jobsTable) addJobsToTable(jobs []circleci.Job, startRow int, nextPageToken string) {
	if len(jobs) != 0 {
		for row, job := range jobs {
			dependencies := getNamedJobDependencies(job, jobs)
			var dependenciesString string
			if len(dependencies) == 0 {
				dependenciesString = "[]"
			} else {
				// tview has the concept of Regions which use "[string]" as identifiers
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
				j.table.SetCell(row+startRow, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...").SetStyle(tcell.StyleDefault)
			cell.SetReference(nextPageToken)
			j.table.SetCell(j.table.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		j.table.SetCell(1, 0, cell)
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

func (j jobsTable) clear() {
	row := 1
	for row < j.table.GetRowCount() {
		j.table.RemoveRow(row)
	}
}
