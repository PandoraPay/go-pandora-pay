package gui

import (
	"encoding/hex"
	ui "github.com/gizak/termui/v3"
	"log"
	"os"
	"strconv"
	"time"
)

//test
func GUIInit() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
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

	//go func() {
	//	for {
	//		termWidth2, termHeight2 := ui.TerminalDimensions()
	//		if termWidth != termWidth2 || termHeight2 != termHeight {
	//			termWidth = termWidth2
	//			termHeight = termHeight2
	//			grid.SetRect(0, 0, termWidth, termHeight)
	//			ui.Render(grid)
	//		}
	//		time.Sleep(2 * time.Second)
	//
	//	}
	//}()

	ui.Render(grid)

	go func() {

		ticker := time.NewTicker(100 * time.Millisecond).C

		uiEvents := ui.PollEvents()
		for {

			select {
			case e := <-uiEvents:
				cmdProcess(e)
			case <-ticker:
				infoProcess()
				info2Process()
				ui.Render(logs)
			}

		}
	}()

	CommandDefineCallback("Exit", func(string) {
		os.Exit(1)
		return
	})

	Log("GUI Initialized")

}

func processArgument(any ...interface{}) string {

	var s = ""

	for i, it := range any {

		if i > 0 {
			s += "\n"
		}

		switch v := it.(type) {
		case nil:
			s += " "
		case string:
			s += v
		case int:
			s += strconv.Itoa(v)
		case []byte:
			s += hex.EncodeToString(v)
		case [32]byte:
			s += hex.EncodeToString(v[:])
		case error:
			s += v.Error()
		default:
			s += "invalid log type"
		}

	}

	return s
}
