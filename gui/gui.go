package gui

import (
	"encoding/hex"
	ui "github.com/gizak/termui/v3"
	"os"
	"strconv"
	"time"
)

func Close() {
	ui.Clear()
	ui.Close()
}

//test
func GUIInit() (err error) {

	if err = InitLogger(); err != nil {
		return
	}

	if err = ui.Init(); err != nil {
		return
	}
	//defer ui.Close()

	infoInit()
	info2Init()
	cmdInit()
	logsInit()

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0/4,
			ui.NewCol(1.0/2, info),
			ui.NewCol(1.0/2, info2),
		),
		ui.NewRow(1.0/4,
			ui.NewCol(1.0/1, cmd),
		),
		ui.NewRow(2.0/4, logs),
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
					cmdProcess(e)
				}
			case <-ticker:
				infoRender()
				info2Render()
				logsRender()
			}

		}
	}()

	CommandDefineCallback("Exit", func(string) error {
		os.Exit(1)
		return nil
	})

	Log("GUI Initialized")
	return
}

func processArgument(any ...interface{}) string {

	var s = ""

	for i, it := range any {

		if i > 0 {
			s += " "
		}

		switch v := it.(type) {
		case nil:
			s += "nil"
		case string:
			s += v
		case int:
			s += strconv.Itoa(v)
		case uint64:
			s += strconv.FormatUint(v, 10)
		case []byte:
			s += hex.EncodeToString(v)
		case error:
			s += v.Error()
		default:
			s += "invalid log type"
		}

	}

	return s
}
