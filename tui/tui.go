package tui

import (
	"fmt"

	"github.com/lupinelab/circlog/config"

	"github.com/rivo/tview"
)

func Run(config config.CirclogConfig, project string) {
	app := tview.NewApplication()

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.SetTitle("circlog").SetBorder(true).SetBorderPadding(1, 1, 1, 1)

	heading := tview.NewFlex()
	heading.AddItem(tview.NewTextArea().SetText(fmt.Sprintf("Organisation: %s\nProject: %s", config.Org, project), false), 0, 1, false)
	layout.AddItem(heading, 3, 1, false)

	pipelinesTable := ShowPipelines(config, project, app, layout)

	err := app.SetRoot(layout, true).SetFocus(pipelinesTable).Run()
	if err != nil {
		panic(err)
	}
}
