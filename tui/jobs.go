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

type jobsPane struct {
	table       *tview.Table
	numPages    int
	update      chan bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (cTui *CirclogTui) newJobsPane() jobsPane {
	table := tview.NewTable()
	table.SetTitle(" JOBS ")
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
			cTui.updateState <- cellRef
			jobDetails, _ := circleci.GetJobSteps(cTui.config, cellRef.JobNumber)
			cTui.steps.populateStepsTree(cellRef, jobDetails)
			cTui.app.SetFocus(cTui.steps.tree)

		case string:
			if cell.Text == "..." {
				cTui.jobs.numPages++
				cTui.jobs.update <- true
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.jobs.watchCancel()
			cTui.jobs.clear()
			table.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.workflows.table)

			return event
		}

		switch event.Rune() {

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog jobs %s -w %s\n", cTui.config.Project, cTui.state.workflow.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.jobs.restartWatcher(func() {
			table.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.Clear()
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	jobsPane := jobsPane{
		table:       table,
		numPages:    1,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}

	jobsPane.update = jobsPane.updater(cTui)

	return jobsPane
}

func (j *jobsPane) watchJobs(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			j.update <- true
		}
	}
}

func (j *jobsPane) restartWatcher(fn func()) {
	j.watchCancel()
	fn()
	j.watchCtx, j.watchCancel = context.WithCancel(context.TODO())
	go j.watchJobs(j.watchCtx)
}

func (j *jobsPane) updater(cTui *CirclogTui) chan bool {
	updateChan := make(chan bool)

	go func() {
		for <-updateChan {
			jobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.state.workflow.Id, j.numPages, "")
			cTui.app.QueueUpdateDraw(func() {
				currentRow := 1
				if j.table.GetRowCount() != 1 {
					currentRow, _ = j.table.GetSelection()
				}

				j.clear()
				j.populateJobsTable(jobs, j.table.GetRowCount(), nextPageToken)

				if currentRow == 1 {
					j.table.ScrollToBeginning()
				} else {
					j.table.Select(currentRow, 0)
				}
			})
		}
	}()

	return updateChan

}

func (j *jobsPane) populateJobsTable(jobs []circleci.Job, startRow int, nextPageToken string) {
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

func (j *jobsPane) clear() {
	row := 1
	for row < j.table.GetRowCount() {
		j.table.RemoveRow(row)
	}
}
