package gui

import (
	"encoding/hex"
	"errors"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"pandora-pay/addresses"
	"strconv"
	"sync/atomic"
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
	Name     string
	Text     string
	Callback func(string) error
}

var commands = []Command{
	{Name: "Wallet", Text: "Decrypt"},
	{Name: "Wallet", Text: "Show Mnemnonic"},
	{Name: "Wallet", Text: "List Addresses"},
	{Name: "Wallet", Text: "Show Private Key"},
	{Name: "Wallet", Text: "Import Private Key"},
	{Name: "Wallet", Text: "Remove Address"},
	{Name: "Wallet", Text: "Create New Address"},
	{Name: "Wallet:TX", Text: "Transfer"},
	{Name: "Wallet:TX", Text: "Delegate"},
	{Name: "Wallet:TX", Text: "Unstake"},
	{Name: "Wallet", Text: "Export JSON"},
	{Name: "Wallet", Text: "Import JSON"},
	{Name: "Mempool", Text: "Show Txs"},
	{Name: "App", Text: "Exit"},
}

var cmd *widgets.List
var cmdStatus = &atomic.Value{}
var cmdInput = &atomic.Value{}
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
		if cmdStatus.Load().(string) == "read" {
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

		if cmdStatus.Load().(string) == "cmd" {

			cmd.Lock()
			command := commands[cmd.SelectedRow]
			cmd.SelectedRow = 0
			cmd.Unlock()

			if command.Callback != nil {
				OutputClear()
				go func() {

					if err := command.Callback(command.Text); err != nil {
						Error(err)
						cmdStatus.Store("output done")
					} else {
						OutputDone()
					}

				}()
			}
		} else if cmdStatus.Load().(string) == "output done" {
			OutputRestore()
		} else if cmdStatus.Load().(string) == "read" {
			cmdInputCn <- cmdInput.Load().(string)
		}

	}

	if cmdStatus.Load().(string) == "read" && !NotAcceptedCharacters[e.ID] {
		cmd.Lock()
		str := cmdInput.Load().(string)

		char := e.ID
		if char == "<Space>" {
			char = " "
		}
		if char == "<Backspace>" {
			char = ""
			if len(str) > 0 {
				str = str[:len(str)-1]
			}
		}
		str += char
		cmdInput.Store(str)
		cmd.Rows[len(cmd.Rows)-1] = "-> " + str
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
	cmdInput.Store("")
	cmd.Rows = append(cmd.Rows, "")
	cmd.Rows = append(cmd.Rows, text)
	cmd.Rows = append(cmd.Rows, "-> ")
	cmd.SelectedRow = len(cmd.Rows) - 1
	cmdStatus.Store("read")
	cmd.Unlock()
	ui.Render(cmd)

	return cmdInputCn
}

func OutputReadString(text string) (out string, ok bool) {
	out, ok = <-outputRead(text)
	return
}

func OutputReadInt(text string, acceptedValues []int) (out int, ok bool) {
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
		if acceptedValues != nil {
			acceptedValuesStr := ""
			for _, acceptedValue := range acceptedValues {
				if out == acceptedValue {
					return
				}
				acceptedValuesStr += strconv.Itoa(acceptedValue) + " "
			}
			OutputWrite("Invalid values. Values accepted: " + acceptedValuesStr)
		}
		return
	}
}

func OutputReadUint64(text string, acceptedValues []uint64, acceptEmpty bool) (out uint64, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-outputRead(text); !ok {
			return
		}
		if acceptEmpty && str == "" {
			return
		}

		if out, err = strconv.ParseUint(str, 10, 64); err != nil {
			OutputWrite("Invalid Number")
			continue
		}
		if acceptedValues != nil {
			acceptedValuesStr := ""
			for _, acceptedValue := range acceptedValues {
				if out == acceptedValue {
					return
				}
				acceptedValuesStr += strconv.FormatUint(acceptedValue, 64) + " "
			}
			OutputWrite("Invalid values. Values accepted: " + acceptedValuesStr)
		}
		return
	}
}

func OutputReadFloat64(text string, acceptedValues []float64) (out float64, ok bool) {
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
		if acceptedValues != nil {
			acceptedValuesStr := ""
			for _, acceptedValue := range acceptedValues {
				if out == acceptedValue {
					return
				}
				acceptedValuesStr += strconv.FormatFloat(acceptedValue, 'f', 10, 64) + " "
			}
			OutputWrite("Invalid values. Values accepted: " + acceptedValuesStr)
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

		if acceptedLengths != nil {
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
	cmdStatus.Store("output done")
}

func OutputRestore() {
	OutputClear()
	cmd.Lock()
	cmd.SelectedRow = 0
	cmd.Rows = cmdRows
	cmd.Unlock()
	ui.Render(cmd)
	cmdStatus.Store("cmd")
}

func cmdInit() {
	cmdStatus.Store("cmd")
	cmdInput.Store("")

	cmd = widgets.NewList()
	cmd.Title = "Commands"
	cmdRows = make([]string, len(commands))
	for i, command := range commands {
		cmdRows[i] = fmt.Sprintf("%2d %10s %s", i, command.Name, command.Text)
	}
	cmd.Rows = cmdRows
	cmd.TextStyle = ui.NewStyle(ui.ColorYellow)
	//cmd.WrapText = true
}
