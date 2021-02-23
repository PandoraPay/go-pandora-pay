package gui

import (
	"encoding/hex"
	"errors"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"os"
	"strconv"
	"time"
	"unicode"
)

var logs, statistics *widgets.Paragraph
var cmd, info *widgets.List
var cmdStatus string = "cmd"
var cmdInput string = ""
var cmdInputChannel = make(chan string)

func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

type Command struct {
	Text     string
	Callback func(string)
}

var commands = []Command{
	{Text: "Decrypt Addresses"},
	{Text: "Show Mnemnonic"},
	{Text: "List Addresses"},
	{Text: "Show Private Key"},
	{Text: "Remove Address"},
	{Text: "Create New Address"},
	{Text: "Save (JSON) Address"},
	{Text: "Exit"},
}

var infoMap = make(map[string]string)

//test
func GUIInit() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	//defer ui.Close()

	info = widgets.NewList()
	info.Title = "Node Info"

	cmd = widgets.NewList()
	cmd.Title = "Commands"
	var rows []string
	for i, command := range commands {
		rows = append(rows, strconv.Itoa(i)+" "+command.Text)
	}
	cmd.Rows = rows
	cmd.TextStyle = ui.NewStyle(ui.ColorYellow)
	cmd.WrapText = true

	ui.Render(cmd)

	logs = widgets.NewParagraph()
	logs.Title = "Logs"
	logs.Text = ""
	logs.WrapText = true

	ui.Render(logs)

	statistics = widgets.NewParagraph()
	statistics.Title = "Statistics"
	statistics.Text = "empty"

	ui.Render(statistics)

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0/4,
			ui.NewCol(1.0/2, info),
			ui.NewCol(1.0/2, statistics),
		),
		ui.NewRow(2.0/4,
			ui.NewCol(1.0/1, cmd),
		),
		ui.NewRow(1.0/4, logs),
	)

	ui.Render(grid)

	drawStatistics := func(count int) {
		statistics.Text = "Time: " + time.Now().Format("2006.01.02 15:04:05") + "\n"
		ui.Render(statistics)
	}

	run := func() {

		ticker := time.NewTicker(time.Second).C
		tickerCount := 1
		drawStatistics(tickerCount)
		tickerCount++

		uiEvents := ui.PollEvents()
		for {

			select {
			case e := <-uiEvents:
				switch e.ID {
				case "<C-c>":
					os.Exit(1)
					return
				case "<Down>":
					cmd.ScrollDown()
				case "<Up>":
					cmd.ScrollUp()
				case "<C-d>":
					cmd.ScrollHalfPageDown()
				case "<C-u>":
					cmd.ScrollHalfPageUp()
				case "<C-f>":
					cmd.ScrollPageDown()
				case "<C-b>":
					cmd.ScrollPageUp()
				case "<Home>":
					cmd.ScrollTop()
				case "<End>":
					cmd.ScrollBottom()
				case "<Enter>":

					if cmdStatus == "cmd" {
						command := commands[cmd.SelectedRow]
						cmd.SelectedRow = 0
						if command.Callback != nil {
							OutputClear()
							go command.Callback(command.Text)
						}
					} else if cmdStatus == "output done" {
						OutputClear()
						cmd.SelectedRow = 0
						cmd.Rows = rows
						cmdStatus = "cmd"
					} else if cmdStatus == "read" {
						cmdInputChannel <- cmdInput
					}

				}

				if cmdStatus == "read" && (IsLetter(e.ID) || e.ID == "<Backspace>" || e.ID == "<Space>") {
					char := e.ID
					if char == "<Space>" {
						char = " "
					}
					if char == "<Backspace>" {
						char = ""
						cmdInput = cmdInput[:len(cmdInput)-1]
					}
					cmdInput = cmdInput + char
					cmd.Rows[len(cmd.Rows)-1] = "-> " + cmdInput
					ui.Render(cmd)
				}

				// previousKey = e.ID

				ui.Render(cmd)
			case <-ticker:
				drawStatistics(tickerCount)
				tickerCount++
			}

		}
	}

	go run()

	CommandDefineCallback("Exit", func(string) {
		os.Exit(1)
	})

	Log("GUI Initialized")

}

func CommandDefineCallback(Text string, callback func(string)) {

	for i := range commands {
		if commands[i].Text == Text {
			commands[i].Callback = callback
			return
		}
	}

	Error("Command "+Text+" was not found", errors.New("Command not found"))

}

func InfoUpdate(key string, text string) {
	infoMap[key] = text
	rows := []string{}
	for key, value := range infoMap {
		rows = append(rows, key+": "+value)
	}
	info.Rows = rows
	ui.Render(info)
}

func processArgument(any interface{}) string {
	switch v := any.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case []byte:
		return hex.EncodeToString(v)
	case [32]byte:
		return hex.EncodeToString(v[:])
	case error:
		return v.Error()
	default:
		return "invalid log type"
	}
}

func OutputWrite(any interface{}) {
	cmd.Rows = append(cmd.Rows, processArgument(any))
	ui.Render(cmd)
}

func outputRead(any interface{}) <-chan string {

	cmdInput = ""
	cmd.Rows = append(cmd.Rows, "")
	cmd.Rows = append(cmd.Rows, processArgument(any)+" : ")
	cmd.Rows = append(cmd.Rows, "-> ")
	cmdStatus = "read"
	ui.Render(cmd)

	return cmdInputChannel
}

func OutputReadString(any interface{}) <-chan string {
	return outputRead(any)
}

func OutputReadInt(any interface{}) <-chan int {
	r := make(chan int)

	go func() {

		for {
			str := <-outputRead(any)
			no, err := strconv.Atoi(str)
			if err != nil {
				OutputWrite("Invalid Number")
				continue
			}
			r <- no
			return
		}
	}()

	return r
}

func OutputClear() {
	cmd.Rows = []string{}
	ui.Render(cmd)
}

func OutputDone() {
	OutputWrite("")
	OutputWrite("Press space to return...")
	cmdStatus = "output done"
}

func message(any interface{}, color ui.Color) {
	logs.TextStyle = ui.NewStyle(color)
	logs.Text = logs.Text + processArgument(any) + "\n"
	ui.Render(logs)
}

func Fatal(any interface{}) {
	message(any, ui.ColorRed)
	os.Exit(1)
}

func Log(any interface{}) {
	message(any, ui.ColorClear)
}

func Info(any interface{}) {
	message(any, ui.ColorBlue)
}

func Error(any interface{}, err error) error {
	message(any, ui.ColorRed)
	message(err, ui.ColorRed)
	return err
}
