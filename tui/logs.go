package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (cTui *CirclogTui) newLogsPane() *tview.TextView {
	view := tview.NewTextView()
	view.SetTitle(" LOGS - Autoscroll Enabled ")
	view.SetBackgroundColor(tcell.ColorDefault)
	view.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	view.SetBorderColor(tcell.ColorGrey)
	view.SetDynamicColors(true)
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyEsc:
			cTui.state.mu.Lock()
			if cTui.state.follow {
				cTui.state.follow = false
				cTui.stepsPane.SetTitle(" STEPS - Follow Disabled ")
			}
			cTui.state.mu.Unlock()

			view.Clear()
			view.SetBorderColor(tcell.ColorGrey)

			cTui.app.SetFocus(cTui.stepsPane)

			return event

		case tcell.KeyUp:
			cTui.state.mu.Lock()
			cTui.state.autoScroll = false
			view.SetTitle(" LOGS - Autoscroll Disabled ")
			cTui.state.follow = false
			cTui.state.mu.Unlock()
			cTui.stepsPane.SetTitle(" STEPS - Follow Disabled ")

			return event
		}

		switch event.Rune() {

		case 'f':
			toggleFollow(cTui)

		case 'a':
			cTui.state.mu.Lock()
			cTui.state.autoScroll = !cTui.state.autoScroll
			if cTui.state.autoScroll {
				view.SetTitle(" LOGS - Autoscroll Enabled ")
				view.ScrollToEnd()
			} else {
				cTui.state.mu.Lock()
				cTui.state.follow = false
				cTui.state.mu.Unlock()
				cTui.stepsPane.SetTitle(" STEPS - Follow Disabled ")
				view.SetTitle(" LOGS - Autoscroll Disabled ")
			}
			cTui.state.mu.Unlock()

		case 'b':
			cTui.clearAll()
			cTui.config.Branch = ""
			cTui.app.SetFocus(cTui.branchSelect)

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
		view.SetBorderColor(tcell.ColorDefault)
		cTui.paneControls.SetText("Toggle Autoscroll\t[A]")
	})

	return view
}

// func (l *logsPane) watchLogs(ctx context.Context, cTui *CirclogTui) {
// 	logsChan := make(chan string)
// 	ticker := time.NewTicker(refreshInterval)

// LOOP:
// 	for {
// 		go func() {
// 			logs, _ := circleci.GetLogs(
// 				cTui.config,
// 				cTui.state.job.JobNumber,
// 				cTui.state.action.Step,
// 				cTui.state.action.Index,
// 				cTui.state.action.AllocationId,
// 			)

// 			logsChan <- logs
// 		}()

// 		select {
// 		case <-ctx.Done():
// 			ticker.Stop()
// 			break LOOP

// 		case logs := <-logsChan:
// 			cTui.app.QueueUpdateDraw(func() {
// 				row, col := l.view.GetScrollOffset()
// 				l.updateLogsView(logs)
// 				if l.autoScroll {
// 					l.view.ScrollToEnd()
// 				} else {
// 					l.view.ScrollTo(row, col)
// 				}
// 			})

// 			<-ticker.C
// 		}
// 	}

// }

func (cTui *CirclogTui) updateLogsView(logs string) {
	cTui.logsPane.SetText(tview.TranslateANSI(logs))
}
