package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newJobsTable(config config.CirclogConfig, project string) *tview.Table {
	jobsTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	jobsTable.SetTitle(" JOBS ").SetBorder(true)

	return jobsTable
}

func updateJobsTable(config config.CirclogConfig, project string, workflow circleci.Workflow, jobsTable *tview.Table) {
	jobs, nextPageToken, _ := circleci.GetWorkflowJobs(config, workflow.Id, 1, "")

	jobsTable.Clear()

	for column, header := range []string{"Name", "Duration", "Dependencies"} {
		jobsTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	addJobsToTable(jobs, jobsTable.GetRowCount(), nextPageToken)

	jobsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyBackspace2 {
			jobsTable.Clear()
			app.SetFocus(workflowsTable)
		}

		if event.Rune() == 'd' {
			app.Stop()
			fmt.Printf("circlog jobs %s -w %s\n", project, workflow.Id)
		}

		return event
	})

	jobsTable.SetSelectedFunc(func(row int, col int) {
		cell := jobsTable.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {
		case circleci.Job:
			updateStepsTree(config, project, cellRef)
		case string:
			if cell.Text == "Next page..." {
				nextPageToken := cell.GetReference().(string)
				newJobs, nextPageToken, _ := circleci.GetWorkflowJobs(config, workflow.Id, 1, nextPageToken)
				addJobsToTable(newJobs, jobsTable.GetRowCount()-1, nextPageToken)
			}
		}
	})

	app.SetFocus(jobsTable)
}

func addJobsToTable(jobs []circleci.Job, startRow int, nextPageToken string) {
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
				jobsTable.SetCell(row+startRow, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("Next page...").SetStyle(tcell.StyleDefault)
			cell.SetReference(nextPageToken)
			jobsTable.SetCell(jobsTable.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		jobsTable.SetCell(1, 0, cell)
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
