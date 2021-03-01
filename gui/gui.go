package gui

import (
	"encoding/hex"
	"errors"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var logs, statistics *widgets.Paragraph
var cmd, info *widgets.List
var cmdStatus string = "cmd"
var cmdInput string = ""
var cmdInputChannel = make(chan string)

func isLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

type Command struct {
	Text     string
	Callback func(string) error
}

var commands = []Command{
	{Text: "Decrypt Addresses"},
	{Text: "Show Mnemnonic"},
	{Text: "List Addresses"},
	{Text: "Show Private Key"},
	{Text: "Remove Address"},
	{Text: "Create New Address"},
	{Text: "Export (JSON) Wallet"},
	{Text: "Import (JSON) Wallet"},
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
		ui.NewRow(1.0/4,
			ui.NewCol(1.0/1, cmd),
		),
		ui.NewRow(2.0/4, logs),
	)

	//go func(){
	//	for{
	//		termWidth2, termHeight2 := ui.TerminalDimensions()
	//		if termWidth != termWidth2 || termHeight2 != termHeight {
	//			termWidth = termWidth2
	//			termHeight = termHeight2
	//			grid.SetRect(0, 0, termWidth, termHeight)
	//			ui.Render(grid)
	//		}
	//		time.Sleep(1000 * time.Millisecond)
	//
	//	}
	//}()

	ui.Render(grid)

	drawStatistics := func(count int) {
		statistics.Text = "Time: " + time.Now().Format("2006.01.02 15:04:05") + "\n"
		ui.Render(statistics)
	}

	go func() {

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
							go func() {
								err := command.Callback(command.Text)
								if err != nil {
									Error(err)
								} else {
									OutputDone()
								}
							}()
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

				if cmdStatus == "read" && (isLetter(e.ID) || e.ID == "<Backspace>" || e.ID == "<Space>") {
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
	}()

	CommandDefineCallback("Exit", func(string) error {
		os.Exit(1)
		return nil
	})

	Log("GUI Initialized")

}

func CommandDefineCallback(Text string, callback func(string) error) {

	for i := range commands {
		if commands[i].Text == Text {
			commands[i].Callback = callback
			return
		}
	}

	Error(errors.New("Command " + Text + " was not found"))
}

func InfoUpdate(key string, text string) {
	infoMap[key] = text
	rows := []string{}
	for key, value := range infoMap {
		rows = append(rows, key+": "+value)
	}
	sort.Strings(rows)
	info.Rows = rows
	ui.Render(info)
}

func processArgument(any ...interface{}) string {

	var s = ""

	for _, it := range any {

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

func message(color string, any ...interface{}) {
	ss := strings.Split(logs.Text, "\n")
	pos := len(ss) - 16
	if pos < 0 {
		pos = 0
	}
	logs.Text = strings.Join(ss[pos:], "\n") + "[" + processArgument(any...) + "]" + color + "\n"
	ui.Render(logs)
}

func Log(any ...interface{}) {
	message("()", any...)
}

func Info(any ...interface{}) {
	message("(fg:blue)", any...)
}

func Warning(any ...interface{}) {
	message("(fg:yellow)", any...)
}

func Fatal(any ...interface{}) error {
	message("(fg:red,fg:bold)", any...)
	os.Exit(1)
	return nil
}

func Error(any ...interface{}) error {
	message("(fg:red)", any...)
	for _, it := range any {

		switch v := it.(type) {
		case error:
			return v
		default:

		}

	}
	return nil
}
