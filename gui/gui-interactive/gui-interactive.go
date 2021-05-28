package gui_interactive

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	gui_interface "pandora-pay/gui/gui-interface"
	"pandora-pay/gui/gui-logger"
	"sync"
	"time"
)

type GUIInteractive struct {
	gui_interface.GUIInterface
	logger *gui_logger.GUILogger

	cmd        *widgets.List
	cmdRows    []string
	cmdStatus  string //string
	cmdInput   string //string
	cmdInputCn chan string

	logs *widgets.Paragraph

	info2    *widgets.List
	info2Map *sync.Map

	info    *widgets.List
	infoMap *sync.Map
}

func (g *GUIInteractive) Close() {
	ui.Clear()
	ui.Close()
}

func CreateGUIInteractive() (g *GUIInteractive, err error) {

	var logger *gui_logger.GUILogger
	if logger, err = gui_logger.CreateLogger(); err != nil {
		return
	}

	g = &GUIInteractive{
		logger: logger,
	}

	if err = ui.Init(); err != nil {
		return
	}

	//defer ui.Close()

	g.infoInit()
	g.info2Init()
	g.cmdInit()
	g.logsInit()

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0/4,
			ui.NewCol(1.0/2, g.info),
			ui.NewCol(1.0/2, g.info2),
		),
		ui.NewRow(1.0/4,
			ui.NewCol(1.0/1, g.cmd),
		),
		ui.NewRow(2.0/4, g.logs),
	)

	ui.Render(grid)

	ticker := time.NewTicker(100 * time.Millisecond).C
	go func() {

		uiEvents := ui.PollEvents()
		for {

			select {
			case e := <-uiEvents:
				switch e.ID {
				case "<Resize>":
					payload := e.Payload.(ui.Resize)
					grid.SetRect(0, 0, payload.Width, payload.Height)
					ui.Clear()
					ui.Render(grid)
				default:
					g.cmdProcess(e)
				}
			case <-ticker:
				g.infoRender()
				g.info2Render()
				g.logsRender()
				ui.Render(g.info, g.info2, g.logs)
			}

		}
	}()

	g.CommandDefineCallback("Exit", func(string) error {
		os.Exit(1)
		return nil
	})

	return
}
