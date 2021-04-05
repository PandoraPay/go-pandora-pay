package gui

import (
	"encoding/hex"
	"errors"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"pandora-pay/addresses"
	"strconv"
)

var NotAcceptedCharacters = map[string]bool{
	"<Ctrl>":                true,
	"<Enter>":               true,
	"<MouseWheelUp>":        true,
	"<MouseWheelDown>":      true,
	"<MouseLeft>":           true,
	"<MouseRelease>":        true,
	"<Shift>":               true,
	"<Down>":                true,
	"<Up>":                  true,
	"<Left>":                true,
	"<Right>":               true,
	"<Tab>":                 true,
	"NotAcceptedCharacters": true,
}

type Command struct {
	Text     string
	Callback func(string) error
}

var commands = []Command{
	{Text: "Wallet : Decrypt"},
	{Text: "Wallet : Show Mnemnonic"},
	{Text: "Wallet : List Addresses"},
	{Text: "Wallet : Show Private Key"},
	{Text: "Wallet : Remove Address"},
	{Text: "Wallet : Create New Address"},
	{Text: "Wallet : TX: Transfer"},
	{Text: "Wallet : TX: Delegate"},
	{Text: "Wallet : TX: Withdraw"},
	{Text: "Wallet : Export JSON"},
	{Text: "Wallet : Import JSON"},
	{Text: "Exit"},
}

var cmd *widgets.List
var cmdStatus = "cmd"
var cmdInput = ""
var cmdInputCn = make(chan string)
var cmdRows []string

func CommandDefineCallback(Text string, callback func(string) error) {

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
		if cmdStatus == "read" {
			close(cmdInputCn)
			cmdInputCn = make(chan string)
			return
		}
		os.Exit(1)
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

					if err := command.Callback(command.Text); err != nil {
						Error(err)
						cmdStatus = "output done"
					} else {
						OutputDone()
					}

				}()
			}
		} else if cmdStatus == "output done" {
			OutputRestore()
		} else if cmdStatus == "read" {
			cmdInputCn <- cmdInput
		}

	}

	if cmdStatus == "read" && !NotAcceptedCharacters[e.ID] {
		cmd.Lock()
		char := e.ID
		if char == "<Space>" {
			char = " "
		}
		if char == "<Backspace>" {
			char = ""
			if len(cmdInput) > 0 {
				cmdInput = cmdInput[:len(cmdInput)-1]
			}
		}
		cmdInput = cmdInput + char
		cmd.Rows[len(cmd.Rows)-1] = "-> " + cmdInput
		cmd.Unlock()
	}

	// previousKey = e.ID

	ui.Render(cmd)
}

func OutputWrite(any interface{}) {
	cmd.Lock()
	cmd.Rows = append(cmd.Rows, processArgument(any))
	cmd.SelectedRow = len(cmd.Rows) - 1
	cmd.Unlock()
	ui.Render(cmd)
}

func outputRead(text string) <-chan string {

	cmd.Lock()
	cmdInput = ""
	cmd.Rows = append(cmd.Rows, "")
	cmd.Rows = append(cmd.Rows, text)
	cmd.Rows = append(cmd.Rows, "-> ")
	cmd.SelectedRow = len(cmd.Rows) - 1
	cmdStatus = "read"
	cmd.Unlock()
	ui.Render(cmd)

	return cmdInputCn
}

func OutputReadString(text string) (out string, ok bool) {
	out, ok = <-outputRead(text)
	return
}

func OutputReadInt(text string) (out int, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		if out, err = strconv.Atoi(str); err != nil {
			OutputWrite("Invalid Number")
			continue
		}
		return
	}
}

func OutputReadUint64(text string) (out uint64, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		if out, err = strconv.ParseUint(str, 10, 64); err != nil {
			OutputWrite("Invalid Number")
			continue
		}
		return
	}
}

func OutputReadFloat64(text string) (out float64, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		if out, err = strconv.ParseFloat(str, 64); err != nil {
			OutputWrite("Invalid Number")
			continue
		}
		return
	}
}

func OutputReadAddress(text string) (address *addresses.Address, ok bool) {
	var str string
	var err error

	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		address, err = addresses.DecodeAddr(str)
		if err != nil {
			OutputWrite("Invalid Address")
			continue
		}
		return
	}
}

func OutputReadBool(text string) (out bool, ok bool) {
	var str string
	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		if str == "y" {
			return true, true
		} else if str == "n" {
			return false, true
		} else {
			OutputWrite("Invalid boolean answer")
			continue
		}
	}
}

func OutputReadBytes(text string, acceptedLengths []int) (token []byte, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		if token, err = hex.DecodeString(str); err != nil {
			OutputWrite("Invalid Token. The token has to be a hex")
			continue
		}

		acceptedLengthsStr := ""
		for _, acceptedLength := range acceptedLengths {
			acceptedLengthsStr = acceptedLengthsStr + strconv.Itoa(acceptedLength) + " , "
			if len(token) == acceptedLength {
				return
			}
		}
		OutputWrite("Invalid value. Lengths accepted: " + acceptedLengthsStr)
	}
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

func OutputRestore() {
	OutputClear()
	cmd.Lock()
	cmd.SelectedRow = 0
	cmd.Rows = cmdRows
	cmd.Unlock()
	ui.Render(cmd)
	cmdStatus = "cmd"
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
