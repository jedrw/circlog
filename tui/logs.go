package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type logsView struct {
	view *tview.TextView
}

func (cTui *CirclogTui) newLogsView() logsView {
	view := tview.NewTextView()
	view.SetTitle(" LOGS ").SetBorder(true).SetBorderPadding(0, 0, 1, 1)

	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			view.Clear()
			cTui.app.SetFocus(cTui.steps.tree)
		}

		if event.Rune() == 'b' {
			cTui.app.SetFocus(cTui.branchSelect)
		}

		if event.Rune() == 'd' {
			cTui.app.Stop()
			fmt.Printf("circlog logs %s -j %d -s %d -i %d -a \"%s\"\n",
				cTui.config.Project,
				cTui.tuiState.job.JobNumber,
				cTui.tuiState.action.Step,
				cTui.tuiState.action.Index,
				cTui.tuiState.action.AllocationId,
			)
		}

		return event
	})

	view.SetFocusFunc(func() {
		cTui.controls.SetText(cTui.controlBindings)
	})

	return logsView{
		view: view,
	}
}

func (l logsView) updateLogsView(logs string) {
	l.view.SetText(logs).ScrollToBeginning()
}
