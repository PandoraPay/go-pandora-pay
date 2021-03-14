package gui

import (
	"errors"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"pandora-pay/helpers"
	"strconv"
	"unicode"
)

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
	Callback func(string)
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

var cmd *widgets.List
var cmdStatus = "cmd"
var cmdInput = ""
var cmdInputChannel = make(chan string)
var cmdRows []string

func CommandDefineCallback(Text string, callback func(string)) {

	for i := range commands {
		if commands[i].Text == Text {
			commands[i].Callback = callback
			return
		}
	}

	Error(errors.New("Command " + Text + " was not found"))
}

func cmdProcess(e ui.Event) {
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

					defer func() {
						if err := recover(); err != nil {
							Error(helpers.ConvertRecoverError(err))
						} else {
							OutputDone()
						}
					}()

					command.Callback(command.Text)

				}()
			}
		} else if cmdStatus == "output done" {
			OutputClear()
			cmd.Lock()
			cmd.SelectedRow = 0
			cmd.Rows = cmdRows
			cmd.Unlock()
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
		cmd.Lock()
		cmd.Rows[len(cmd.Rows)-1] = "-> " + cmdInput
		cmd.Unlock()
	}

	// previousKey = e.ID

	ui.Render(cmd)
}

func OutputWrite(any interface{}) {
	cmd.Lock()
	cmd.Rows = append(cmd.Rows, processArgument(any))
	cmd.Unlock()
	ui.Render(cmd)
}

func outputRead(any interface{}) <-chan string {

	cmd.Lock()
	cmdInput = ""
	cmd.Rows = append(cmd.Rows, "")
	cmd.Rows = append(cmd.Rows, processArgument(any)+" : ")
	cmd.Rows = append(cmd.Rows, "-> ")
	cmdStatus = "read"
	cmd.Unlock()
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
	cmd.Lock()
	cmd.Rows = []string{}
	cmd.Unlock()
	ui.Render(cmd)
}

func OutputDone() {
	OutputWrite("")
	OutputWrite("Press space to return...")
	cmdStatus = "output done"
}

func cmdInit() {
	cmd = widgets.NewList()
	cmd.Title = "Commands"
	cmdRows = make([]string, len(commands))
	for i, command := range commands {
		cmdRows[i] = strconv.Itoa(i) + " " + command.Text
	}
	cmd.Rows = cmdRows
	cmd.TextStyle = ui.NewStyle(ui.ColorYellow)
	cmd.WrapText = true
}
