package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newJobsTable(config config.CirclogConfig, project string) *tview.Table {
	jobsTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	jobsTable.SetTitle(" JOBS ").SetBorder(true)

	jobsTable.SetSelectedFunc(func(row int, col int) {
		cell := jobsTable.GetCell(row, 0)
		if cell.Text != "None" && cell.Text != "" {
			job := cell.GetReference().(circleci.Job)
			updateStepsTree(config, project, job)
		}
	})

	return jobsTable
}

func updateJobsTable(config config.CirclogConfig, project string, workflow circleci.Workflow, jobsTable *tview.Table) {
	jobs, _ := circleci.GetWorkflowJobs(config, workflow.Id)

	jobsTable.Clear()

	for column, header := range []string{"Name", "Duration", "Dependencies"} {
		jobsTable.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

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
			for column, attr := range []string{job.Name, job.StoppedAt.Sub(job.StartedAt).String(), dependenciesString} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(job.Status))
				cell.SetReference(job)
				jobsTable.SetCell(row+1, column, cell)
			}
		}
	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		jobsTable.SetCell(1, 0, cell)
	}

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

	app.SetFocus(jobsTable)
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
