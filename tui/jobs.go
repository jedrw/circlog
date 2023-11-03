package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/lupinelab/circlog/config"
	"github.com/rivo/tview"
)

func newJobsTable(config config.CirclogConfig, project string) *tview.Table {
	jobsTable := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0)
	jobsTable.SetTitle(" JOBS ").SetBorder(true)

	jobsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyBackspace2 {
			jobsTable.Clear()
			app.SetFocus(workflowsTable)
		}
		return event
	})

	jobsTable.SetSelectedFunc(func(row int, col int) {
		cell := jobsTable.GetCell(row, 0)
		if cell.Text != "None" {
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
			var dependencies []string
			for _, dependsOnJobId := range job.Dependencies {
				for _, job := range jobs {
					if dependsOnJobId == job.Id {
						dependencies = append(dependencies, job.Name)
					}
				}
			}

			for column, attr := range []string{job.Name, job.StoppedAt.Sub(job.StartedAt).String(), fmt.Sprintf("%v", dependencies)} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(job.Status))
				cell.SetReference(job)
				jobsTable.SetCell(row+1, column, cell)
			}
		}
	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		jobsTable.SetCell(1, 0, cell)
	}

	app.SetFocus(jobsTable)
}
