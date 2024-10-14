package tui

import "github.com/lupinelab/circlog/circleci"

func toggleFollow(cTui *CirclogTui) {
	cTui.logs.restartWatcher(cTui, func() {
		cTui.steps.restartWatcher(cTui, func() {
			if !cTui.steps.follow {
				cTui.steps.follow = true
				cTui.logs.autoScroll = true
				cTui.steps.tree.SetTitle(" STEPS - Follow Enabled ")
				cTui.logs.view.SetTitle(" LOGS - Autoscroll Enabled ")
				steps := cTui.steps.tree.GetRoot().GetChildren()
				latestStepActions := steps[len(steps)-1].GetChildren()
				for n := len(latestStepActions) - 1; n >= 0; n-- {
					if n == 0 {
						cTui.steps.tree.SetCurrentNode(latestStepActions[n])
						cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
					} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
						cTui.steps.tree.SetCurrentNode(latestStepActions[n])
						cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
					}
				}
			} else {
				cTui.steps.follow = false
				cTui.steps.tree.SetTitle(" STEPS - Follow Disabled ")
			}
		})
	})
}
