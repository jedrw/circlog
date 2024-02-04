package tui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type logsView struct {
	view          *tview.TextView
	autoScroll    bool
	refreshCtx    context.Context
	refreshCancel context.CancelFunc
}

func (cTui *CirclogTui) newLogsView() logsView {
	view := tview.NewTextView()
	view.SetTitle(" LOGS - Autoscroll Enabled ")
	view.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	view.SetDynamicColors(true)

	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.logs.refreshCancel()
			view.Clear()
			cTui.app.SetFocus(cTui.steps.tree)

			return event
		}

		switch event.Rune() {
		
		case 'a':
			cTui.logs.refreshCancel()
			cTui.logs.autoScroll = !cTui.logs.autoScroll
			if cTui.logs.autoScroll {
				view.SetTitle(" LOGS - Autoscroll Enabled ")
				view.ScrollToEnd()
			} else {
				view.SetTitle(" LOGS - Autoscroll Disabled	 ")
			}
			cTui.logs.refreshCtx, cTui.logs.refreshCancel = context.WithCancel(context.TODO())
			go cTui.refreshLogsView(cTui.logs.refreshCtx)

		case 'b':
			cTui.app.SetFocus(cTui.branchSelect)

		case 'd':
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

	logsControlBindings := `Move	           [Up/Down]
		Select               [Enter]
		Toggle Autoscroll        [A]
		Select branch            [B]
		Dump command             [D]
		Back/Quit              [Esc]
	`

	view.SetFocusFunc(func() {
		cTui.logs.refreshCancel()
		cTui.controls.SetText(logsControlBindings)
		cTui.logs.refreshCtx, cTui.logs.refreshCancel = context.WithCancel(context.TODO())
		go cTui.refreshLogsView(cTui.logs.refreshCtx)
	})

	refreshCtx, refreshCancel := context.WithCancel(context.TODO())

	return logsView{
		view:          view,
		autoScroll:    true,
		refreshCtx:    refreshCtx,
		refreshCancel: refreshCancel,
	}
}

func (l logsView) updateLogsView(logs string) {
	l.view.SetText(tview.TranslateANSI(logs))
}
