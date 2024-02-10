package tui

import (
	"context"
	"time"

	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type logsPane struct {
	view        *tview.TextView
	autoScroll  bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
}

func (l logsPane) watchLogs(cTui *CirclogTui) {
	logsChan := make(chan string)

	for {
		go func() {
			logs, _ := circleci.GetStepLogs(
				cTui.config,
				cTui.tuiState.job.JobNumber,
				cTui.tuiState.action.Step,
				cTui.tuiState.action.Index,
				cTui.tuiState.action.AllocationId,
			)

			logsChan <- logs
		}()

		select {
		case <-cTui.logs.watchCtx.Done():
			return

		case logs := <-logsChan:
			cTui.app.QueueUpdateDraw(func() {
				row, col := cTui.logs.view.GetScrollOffset()
				cTui.logs.updateLogsView(logs)
				if cTui.logs.autoScroll {
					cTui.logs.view.ScrollToEnd()
				} else {
					cTui.logs.view.ScrollTo(row, col)
				}
			})
		}

		time.Sleep(refreshInterval)
	}
}

func (l logsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	cTui.logs.watchCancel()
	fn()
	cTui.logs.watchCtx, cTui.logs.watchCancel = context.WithCancel(context.TODO())
	go cTui.logs.watchLogs(cTui)
}

func (l logsPane) updateLogsView(logs string) {
	l.view.SetText(tview.TranslateANSI(logs))
}
