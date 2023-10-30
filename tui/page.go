package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
)

var ColourByStatus = map[string]tcell.Color{
	"success":      tcell.ColorDarkGreen,
	"running":      tcell.ColorLightGreen,
	"not_run":      tcell.ColorGray,
	"failed":       tcell.ColorDarkRed,
	"error":        tcell.ColorDarkRed,
	"failing":      tcell.ColorPink,
	"on_hold":      tcell.ColorYellow,
	"canceled":     tcell.ColorDarkRed,
	"unauthorized": tcell.ColorDarkRed,

	"created":       tcell.ColorLightGreen,
	"errored":       tcell.ColorDarkRed,
	"setup-pending": tcell.ColorGrey,
	"setup":         tcell.ColorDefault,
	"pending":       tcell.ColorYellowGreen,
}

func StyleForStatus(status string) tcell.Style {
	return tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(ColourByStatus[status])
}

func branchOrTag(pipeline circleci.Pipeline) string {
	branchOrTag := pipeline.Vcs.Branch
	if branchOrTag == "" {
		branchOrTag = pipeline.Vcs.Tag
	}

	return branchOrTag
}
