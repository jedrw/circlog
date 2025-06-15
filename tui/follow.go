package tui

import "github.com/jedrw/circlog/circleci"

func toggleFollow(cTui *CirclogTui) {
	cTui.state.mu.Lock()
	if !cTui.state.follow {
		cTui.state.follow = true
		cTui.state.autoScroll = true
		cTui.stepsPane.SetTitle(" STEPS - Follow Enabled ")
		cTui.logsPane.SetTitle(" LOGS - Autoscroll Enabled ")
		steps := cTui.stepsPane.GetRoot().GetChildren()
		latestStepActions := steps[len(steps)-1].GetChildren()
		for n := len(latestStepActions) - 1; n >= 0; n-- {
			if n == 0 {
				cTui.stepsPane.SetCurrentNode(latestStepActions[n])
				cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
			} else if latestStepActions[n].GetReference().(circleci.Action).Status == "running" {
				cTui.stepsPane.SetCurrentNode(latestStepActions[n])
				cTui.state.action = latestStepActions[n].GetReference().(circleci.Action)
			}
		}
	} else {
		cTui.state.follow = false
		cTui.stepsPane.SetTitle(" STEPS - Follow Disabled ")
	}
	cTui.state.mu.Unlock()
}
