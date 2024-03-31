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
	view        *tview.TextView
	autoScroll  bool
	watchCtx    context.Context
	watchCancel context.CancelFunc
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
			cTui.logs.watchCancel()
			view.Clear()
			view.SetBorderColor(tcell.ColorGrey)
			if cTui.steps.follow {
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = !cTui.steps.follow
					cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
				})
			}

			cTui.app.SetFocus(cTui.steps.tree)

			return event

		case tcell.KeyUp:
			cTui.logs.restartWatcher(cTui, func() {
				cTui.logs.autoScroll = false
				view.SetTitle(" LOGS - Autoscroll Disabled ")
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = false
					cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
				})
			})

			return event
		}

		switch event.Rune() {

		case 'f':
			cTui.logs.restartWatcher(cTui, func() {
				cTui.logs.autoScroll = true
				cTui.logs.view.SetTitle(" LOGS - Autoscroll Enabled ")
				cTui.steps.restartWatcher(cTui, func() {
					cTui.steps.follow = !cTui.steps.follow
					if cTui.steps.follow {
						cTui.steps.tree.SetTitle(" STEPS - Follow Enabled ")
						steps := cTui.steps.tree.GetRoot().GetChildren()
						latestStepActions := steps[len(steps)-1].GetChildren()
						for n := len(latestStepActions) - 1; n >= 0; n-- {
							if n == 0 {
								cTui.steps.tree.SetCurrentNode(latestStepActions[n])
								cTui.app.SetFocus(cTui.steps.tree)
								cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
								cTui.steps.tree.SetCurrentNode(latestStepActions[n])
								cTui.app.SetFocus(cTui.steps.tree)
								cTui.app.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 13, 0))
							}
						}
					} else {
						cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
					}
				})
			})

		case 'a':
			cTui.logs.restartWatcher(cTui, func() {
				cTui.logs.autoScroll = !cTui.logs.autoScroll
				cTui.steps.restartWatcher(cTui, func() {
					if cTui.logs.autoScroll {
						view.SetTitle(" LOGS - Autoscroll Enabled ")
						view.ScrollToEnd()
					} else {
						cTui.steps.follow = false
						cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
						view.SetTitle(" LOGS - Autoscroll Disabled ")
					}
				})
			})

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
		cTui.logs.restartWatcher(cTui, func() {
			view.SetBorderColor(tcell.ColorDefault)
			cTui.paneControls.SetText(`Toggle Autoscroll        [A]`)
		})
	})

	watchCtx, watchCancel := context.WithCancel(context.Background())

	return logsPane{
		view:        view,
		autoScroll:  true,
		watchCtx:    watchCtx,
		watchCancel: watchCancel,
	}
}

func (l *logsPane) watchLogs(ctx context.Context, cTui *CirclogTui) {
	logsChan := make(chan string)
	ticker := time.NewTicker(refreshInterval)

LOOP:
	for {
		go func() {
			logs, _ := circleci.GetStepLogs(
				cTui.config,
				cTui.state.job.JobNumber,
				cTui.state.action.Step,
				cTui.state.action.Index,
				cTui.state.action.AllocationId,
			)

			logsChan <- logs
		}()

		select {
		case <-ctx.Done():
			ticker.Stop()
			break LOOP

		case <-ticker.C:
			logs := <-logsChan
			cTui.app.QueueUpdateDraw(func() {
				row, col := l.view.GetScrollOffset()
				l.updateLogsView(logs)
				if l.autoScroll {
					l.view.ScrollToEnd()
				} else {
					l.view.ScrollTo(row, col)
				}
			})
		}
	}
}

func (l *logsPane) restartWatcher(cTui *CirclogTui, fn func()) {
	l.watchCancel()
	fn()
	l.watchCtx, l.watchCancel = context.WithCancel(context.TODO())
	go l.watchLogs(l.watchCtx, cTui)
}

func (l *logsPane) updateLogsView(logs string) {
	l.view.SetText(tview.TranslateANSI(logs))
}
