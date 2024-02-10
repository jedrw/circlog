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
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (j *jobsPane) watchJobs(cTui *CirclogTui) {
	jobsChan := make(chan []circleci.Job)
	nextPageTokenChan := make(chan string)

	for {
		go func() {
			jobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.tuiState.workflow.Id, cTui.workflows.numPages, "")
			jobsChan <- jobs
			nextPageTokenChan <- nextPageToken
		}()

		time.Sleep(refreshInterval)

		select {
		case <-cTui.jobs.watchCtx.Done():
			return

		default:
			jobs := <-jobsChan
			nextPageToken := <-nextPageTokenChan
			cTui.app.QueueUpdateDraw(func() {
				cTui.jobs.clear()
				cTui.jobs.addJobsToTable(jobs, 1, nextPageToken)
			})
		}
	}
}

func (j *jobsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	cTui.jobs.watchCancel()
	fn()
	cTui.jobs.watchCtx, cTui.jobs.watchCancel = context.WithCancel(context.TODO())
	go cTui.jobs.watchJobs(cTui)
}

func (j *jobsPane) populateTable(jobs []circleci.Job, nextPageToken string) {
	j.clear()
	j.addJobsToTable(jobs, j.table.GetRowCount(), nextPageToken)

}

func (j *jobsPane) addJobsToTable(jobs []circleci.Job, startRow int, nextPageToken string) {
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
