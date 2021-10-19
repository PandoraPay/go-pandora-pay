package gui_interactive

import (
	"encoding/hex"
	"errors"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"pandora-pay/addresses"
	"pandora-pay/gui/gui_interface"
	"path"
	"strconv"
)

var notAcceptedCharacters = map[string]bool{
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
	{Name: "Wallet", Text: "Show Mnemnonic"},
	{Name: "Wallet", Text: "List Addresses"},
	{Name: "Wallet", Text: "Create New Address"},
	{Name: "Wallet", Text: "Show Private Key"},
	{Name: "Wallet", Text: "Import Private Key"},
	{Name: "Wallet", Text: "Remove Address"},
	{Name: "Wallet", Text: "Derive Delegated Stake"},
	{Name: "Wallet:TX", Text: "Private Transfer"},
	{Name: "Wallet:TX", Text: "Update Delegate"},
	{Name: "Wallet:TX", Text: "Unstake"},
	{Name: "Wallet", Text: "Export Addresses"},
	{Name: "Wallet", Text: "Export Address JSON"},
	{Name: "Wallet", Text: "Import Address JSON"},
	{Name: "Wallet", Text: "Export Wallet JSON"},
	{Name: "Wallet", Text: "Import Wallet JSON"},
	{Name: "Wallet", Text: "Encrypt Wallet"},
	{Name: "Wallet", Text: "Decrypt Wallet"},
	{Name: "Wallet", Text: "Remove Encryption"},
	{Name: "Mempool", Text: "Show Txs"},
	{Name: "App", Text: "Exit"},
}

func (g *GUIInteractive) CommandDefineCallback(Text string, callback func(string) error, useIt bool) {

	if !useIt {
		callback = nil
	}

	g.cmdMutex.RLock()
	for i := range commands {
		if commands[i].Text == Text {
			commands[i].Callback = callback
			g.cmdMutex.RUnlock()
			return
		}
	}
	g.cmdMutex.RUnlock()

	g.Error(errors.New("Command " + Text + " was not found"))
}

func (g *GUIInteractive) cmdProcess(e ui.Event) {

	var command *Command
	g.cmdMutex.RLock()
	status := g.cmdStatus
	input := g.cmdInput
	cn := g.cmdInputCn
	if status == "cmd" {
		g.cmd.Lock()
		if g.cmd.SelectedRow >= len(commands) && len(commands) > 0 {
			g.cmd.SelectedRow = len(commands) - 1
		}
		command = &commands[g.cmd.SelectedRow]
		g.cmd.Unlock()
	}
	g.cmdMutex.RUnlock()

	switch e.ID {
	case "<C-c>":
		if status == "read" {
			if cn != nil {
				close(cn)
			}
			return
		}
		g.Close()
	case "<Down>":
		g.cmd.ScrollDown()
	case "<Up>":
		g.cmd.ScrollUp()
	case "<C-d>":
		g.cmd.ScrollHalfPageDown()
	case "<C-u>":
		g.cmd.ScrollHalfPageUp()
	case "<C-f>":
		g.cmd.ScrollPageDown()
	case "<C-b>":
		g.cmd.ScrollPageUp()
	case "<Home>":
		g.cmd.ScrollTop()
	case "<End>":
		g.cmd.ScrollBottom()
	case "<Enter>":

		if status == "cmd" {

			if command.Callback != nil {
				g.outputClear("", nil)
				go func() {

					if err := command.Callback(command.Text); err != nil {
						g.OutputWrite(err)
						g.cmdMutex.Lock()
						g.cmdStatus = "output done"
						g.cmdMutex.Unlock()
					} else {
						g.outputDone()
					}

				}()
			}
		} else if status == "output done" {
			g.outputRestore()
		} else if status == "read" {
			cn <- input
		}

	}

	if status == "read" && !notAcceptedCharacters[e.ID] {

		str := input

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

		g.cmdMutex.Lock()
		g.cmdInput = str
		g.cmdMutex.Unlock()

		g.cmd.Lock()
		g.cmd.Rows[len(g.cmd.Rows)-1] = "-> " + str
		g.cmd.Unlock()
	}

	// previousKey = e.ID
}

func (g *GUIInteractive) OutputWrite(any ...interface{}) {
	str := gui_interface.ProcessArgument(any...)
	g.cmd.Lock()
	g.cmd.Rows = append(g.cmd.Rows, str)
	g.cmd.SelectedRow = len(g.cmd.Rows) - 1
	g.cmd.Unlock()
}

func (g *GUIInteractive) outputRead(text string) <-chan string {

	g.cmd.Lock()
	g.cmd.Rows = append(g.cmd.Rows, "")
	g.cmd.Rows = append(g.cmd.Rows, text)
	g.cmd.Rows = append(g.cmd.Rows, "-> ")
	g.cmd.SelectedRow = len(g.cmd.Rows) - 1
	g.cmd.Unlock()

	cn := make(chan string)
	g.cmdMutex.Lock()
	g.cmdInput = ""
	g.cmdStatus = "read"
	g.cmdInputCn = cn
	defer g.cmdMutex.Unlock()
	return cn
}

func (g *GUIInteractive) OutputReadString(text string) (out string, ok bool) {
	out, ok = <-g.outputRead(text)
	return
}

func (g *GUIInteractive) OutputReadFilename(text, extension string) (out string, ok bool) {
	if out, ok = <-g.outputRead(text); !ok {
		return "", false
	}
	if path.Ext(out) == "" {
		out += "." + extension
	}
	return
}

func (g *GUIInteractive) OutputReadInt(text string, acceptedValues []int) (out int, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-g.outputRead(text); !ok {
			return
		}
		if out, err = strconv.Atoi(str); err != nil {
			g.OutputWrite("Invalid Number")
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
			g.OutputWrite("Invalid values. Values accepted: " + acceptedValuesStr)
			continue
		}
		return
	}
}

func (g *GUIInteractive) OutputReadUint64(text string, acceptedValues []uint64, acceptEmpty bool) (out uint64, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-g.outputRead(text); !ok {
			return
		}
		if acceptEmpty && str == "" {
			return
		}

		if out, err = strconv.ParseUint(str, 10, 64); err != nil {
			g.OutputWrite("Invalid Number")
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
			g.OutputWrite("Invalid values. Values accepted: " + acceptedValuesStr)
			continue
		}
		return
	}
}

func (g *GUIInteractive) OutputReadFloat64(text string, acceptedValues []float64) (out float64, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-g.outputRead(text); !ok {
			return
		}
		if out, err = strconv.ParseFloat(str, 64); err != nil {
			g.OutputWrite("Invalid Number")
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
			g.OutputWrite("Invalid values. Values accepted: " + acceptedValuesStr)
			continue
		}
		return
	}
}

func (g *GUIInteractive) OutputReadAddress(text string) (address *addresses.Address, ok bool) {
	var str string
	var err error

	for {
		if str, ok = <-g.outputRead(text); !ok {
			return
		}
		address, err = addresses.DecodeAddr(str)
		if err != nil {
			g.OutputWrite("Invalid Address")
			continue
		}
		return
	}
}

func (g *GUIInteractive) OutputReadBool(text string) (out bool, ok bool) {
	var str string
	for {
		if str, ok = <-g.outputRead(text); !ok {
			return
		}
		if str == "y" {
			return true, true
		} else if str == "n" {
			return false, true
		} else {
			g.OutputWrite("Invalid boolean answer")
			continue
		}
	}
}

func (g *GUIInteractive) OutputReadBytes(text string, acceptedLengths []int) (input []byte, ok bool) {
	var str string
	var err error
	for {
		if str, ok = <-g.outputRead(text); !ok {
			return
		}
		if input, err = hex.DecodeString(str); err != nil {
			g.OutputWrite("Invalid Data. The input has to be a hex")
			continue
		}

		if acceptedLengths != nil {
			acceptedLengthsStr := ""
			for _, acceptedLength := range acceptedLengths {
				acceptedLengthsStr = acceptedLengthsStr + strconv.Itoa(acceptedLength) + " , "
				if len(input) == acceptedLength {
					return
				}
			}
			g.OutputWrite("Invalid value. Lengths accepted: " + acceptedLengthsStr)
			continue
		}
	}
}

func (g *GUIInteractive) outputClear(newCmdStatus string, rows []string) {
	if rows == nil {
		rows = []string{}
	}
	g.cmd.Lock()
	g.cmd.Rows = rows
	g.cmd.SelectedRow = 0
	g.cmd.Unlock()
	if newCmdStatus != "" {
		g.cmdMutex.Lock()
		g.cmdStatus = newCmdStatus
		g.cmdMutex.Unlock()
	}
}

func (g *GUIInteractive) outputDone() {
	g.OutputWrite("------------------------")
	g.OutputWrite("Press space to return...")
	g.cmdMutex.Lock()
	g.cmdStatus = "output done"
	g.cmdMutex.Unlock()
}

func (g *GUIInteractive) outputRestore() {
	g.outputClear("cmd", g.cmdRows)
}

func (g *GUIInteractive) cmdInit() {
	g.cmdStatus = "cmd"
	g.cmdInput = ""

	g.cmd = widgets.NewList()
	g.cmd.Title = "Commands"
	g.cmdRows = make([]string, len(commands))
	for i, command := range commands {
		g.cmdRows[i] = fmt.Sprintf("%2d %10s %s", i, command.Name, command.Text)
	}
	g.cmd.Rows = g.cmdRows
	g.cmd.TextStyle = ui.NewStyle(ui.ColorYellow)
	//cmd.WrapText = true
}
