package gui_interactive

import (
	"context"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"pandora-pay/gui/gui_interface"
	"pandora-pay/gui/gui_logger"
	"pandora-pay/helpers/generics"
	"time"
)

type GUIInteractiveData struct {
	cmdStatus          string
	cmdInput           string
	cmdInputCn         chan string
	cmdStatusCtx       context.Context
	cmdStatusCtxCancel context.CancelFunc
}

type GUIInteractive struct {
	gui_interface.GUIInterface
	logger *gui_logger.GUILogger

	cmd     *widgets.List
	cmdRows []string

	cmdData *generics.Value[*GUIInteractiveData]

	logs *widgets.Paragraph

	info2    *widgets.List
	info2Map *generics.Map[string, string]

	info    *widgets.List
	infoMap *generics.Map[string, string]

	tickerRender *time.Ticker

	closed bool
}

func (g *GUIInteractive) Close() {
	if g.closed {
		return
	}
	g.closed = true
	g.tickerRender.Stop()
	ui.Clear()
	ui.Close()
	g.logger.GeneralLog.Close()
}

func CreateGUIInteractive() (*GUIInteractive, error) {

	logger, err := gui_logger.CreateLogger()
	if err != nil {
		return nil, err
	}

	g := &GUIInteractive{
		logger:   logger,
		infoMap:  &generics.Map[string, string]{},
		info2Map: &generics.Map[string, string]{},
		cmdData:  &generics.Value[*GUIInteractiveData]{},
	}

	if err = ui.Init(); err != nil {
		return nil, err
	}

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

	g.tickerRender = time.NewTicker(100 * time.Millisecond)
	ticker := g.tickerRender.C
	go func() {

		uiEvents := ui.PollEvents()
		for {

			select {
			case e, ok := <-uiEvents:
				if !ok {
					return
				}
				switch e.ID {
				case "<Resize>":
					payload := e.Payload.(ui.Resize)
					grid.SetRect(0, 0, payload.Width, payload.Height)
					ui.Clear()
					ui.Render(grid)
				default:
					g.cmdProcess(e)
				}
			case _, ok := <-ticker:
				if !ok {
					return
				}
				g.infoRender()
				g.info2Render()
				g.logsRender()

				ui.Render(g.info, g.info2, g.logs, g.cmd)
			}

		}
	}()

	g.CommandDefineCallback("Exit", func(string, context.Context) error {
		os.Exit(1)
		return nil
	}, true)

	return g, nil
}
