package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type logsPane struct {
	view       *tview.TextView
	autoScroll bool
	update     chan bool
	syncCtx    context.Context
	syncCancel context.CancelFunc
}

func (cTui *CirclogTui) newLogsPane() logsPane {
	view := tview.NewTextView()
	view.SetTitle(" LOGS - Autoscroll Enabled ")
	view.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	view.SetBorderColor(tcell.ColorGrey)
	view.SetDynamicColors(true)
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyEsc:
			cTui.updateState <- circleci.Action{}
			if cTui.steps.follow {
				cTui.steps.follow = false
				cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
			}

			view.Clear()
			view.SetBorderColor(tcell.ColorGrey)
			cTui.app.SetFocus(cTui.steps.tree)

			return event

		case tcell.KeyUp:
			cTui.logs.autoScroll = false
			view.SetTitle(" LOGS - Autoscroll Disabled ")
			cTui.steps.follow = false
			cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")

			return event
		}

		switch event.Rune() {

		case 'f':
			cTui.steps.toggleFollow(cTui)

		case 'a':
			cTui.logs.autoScroll = !cTui.logs.autoScroll
			if cTui.logs.autoScroll {
				view.SetTitle(" LOGS - Autoscroll Enabled ")
				view.ScrollToEnd()
			} else {
				cTui.steps.follow = false
				cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
				view.SetTitle(" LOGS - Autoscroll Disabled ")
			}

		case 'd':
			cTui.app.Stop()
			fmt.Printf("circlog logs %s -j %d -s %d -i %d -a \"%s\"\n",
				cTui.config.Project,
				cTui.state.job.JobNumber,
				cTui.state.action.Step,
				cTui.state.action.Index,
				cTui.state.action.AllocationId,
			)
		}

		return event
	})

	view.SetFocusFunc(func() {
		cTui.logs.restartSync(func() {
			view.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.SetText("Toggle Autoscroll        [A]")
		})
	})

	syncCtx, syncCancel := context.WithCancel(context.Background())

	logsPane := logsPane{
		view:       view,
		autoScroll: true,
		syncCtx:    syncCtx,
		syncCancel: syncCancel,
	}

	logsPane.update = logsPane.updater(cTui)

	return logsPane
}

func (l *logsPane) syncLogs(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			l.update <- true
		}
	}
}

func (l *logsPane) restartSync(fn func()) {
	l.syncCancel()
	fn()
	l.syncCtx, l.syncCancel = context.WithCancel(context.TODO())
	go l.syncLogs(l.syncCtx)
}

func (l *logsPane) updater(cTui *CirclogTui) chan bool {
	updateChan := make(chan bool)

	go func() {
		for <-updateChan {
			logs, _ := circleci.GetStepLogs(
				cTui.config,
				cTui.state.job.JobNumber,
				cTui.state.action.Step,
				cTui.state.action.Index,
				cTui.state.action.AllocationId,
			)
			cTui.app.QueueUpdateDraw(func() {
				row, col := l.view.GetScrollOffset()
				l.setLogsViewText(logs)
				if l.autoScroll {
					l.view.ScrollToEnd()
				} else {
					l.view.ScrollTo(row, col)
				}
			})
		}
	}()

	return updateChan
}

func (l *logsPane) setLogsViewText(logs string) {
	l.view.SetText(tview.TranslateANSI(logs))
}
