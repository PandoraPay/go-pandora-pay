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
	{Name: "Wallet:TX", Text: "Private Delegate Stake"},
	{Name: "Wallet:TX", Text: "Private Claim Stake"},
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

	g.cmdChanges.Lock()
	for i := range commands {
		if commands[i].Text == Text {
			commands[i].Callback = callback
			g.cmdChanges.Unlock()
			return
		}
	}
	g.cmdChanges.Unlock()

	g.Error(errors.New("Command " + Text + " was not found"))
}

func (g *GUIInteractive) cmdProcess(e ui.Event) {

	var command *Command

	g.cmdChanges.RLock()
	status := g.cmdStatus
	input := g.cmdInput
	cn := g.cmdInputCn
	if status == "cmd" {
		g.cmd.Lock()
		if g.cmd.SelectedRow < len(commands) && g.cmd.SelectedRow >= 0 {
			command = &Command{
				commands[g.cmd.SelectedRow].Name,
				commands[g.cmd.SelectedRow].Text,
				commands[g.cmd.SelectedRow].Callback,
			}
		}
		g.cmd.Unlock()
	}
	g.cmdChanges.RUnlock()

	unlockRequired := false
	switch e.ID {
	case "<Down>", "<Up>", "<C-d>", "<C-u>", "<C-f>", "<C-b>", "<Home>", "<End>":
		unlockRequired = true
		g.cmd.Lock()
	}

	switch e.ID {
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
	}

	if unlockRequired {
		if g.cmd.SelectedRow >= len(commands) {
			g.cmd.SelectedRow = len(commands) - 1
		}
		if g.cmd.SelectedRow < 0 {
			g.cmd.SelectedRow = 0
		}

		g.cmd.Unlock()
	}

	switch e.ID {
	case "<C-c>":
		if status == "read" {
			if cn != nil {
				close(cn)
				g.cmdChanges.Lock()
				g.cmdInputCn = nil
				g.cmdChanges.Unlock()
			}
			return
		}
		g.Close()
	case "<Enter>":

		if status == "cmd" {

			if command != nil && command.Callback != nil {
				g.outputClear("", nil)

				g.OutputWrite(fmt.Sprintf("Executing cmd %s::%s ...", command.Name, command.Text))
				g.OutputWrite("")

				go func() {

					defer func() {
						err := recover()
						if err != nil {
							g.Error(err)
							g.outputDone()
						}
					}()

					if err := command.Callback(command.Text); err != nil {
						g.OutputWrite(err)

						g.cmdChanges.Lock()
						g.cmdStatus = "output done"
						g.cmdChanges.Unlock()

					} else {
						g.outputDone()
					}

				}()
			}
		} else if status == "output done" {
			g.outputRestore()
		} else if status == "read" {
			if cn != nil {
				cn <- input
			}
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

		g.cmdChanges.Lock()
		g.cmdInput = str
		g.cmdChanges.Unlock()

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
	g.cmdChanges.Lock()
	g.cmdInput = ""

	g.cmdStatus = "read"
	g.cmdInputCn = cn
	g.cmdChanges.Unlock()
	return cn
}

func (g *GUIInteractive) OutputReadString(text string) string {
	out, ok := <-g.outputRead(text)
	if !ok {
		panic(gui_interface.GUIInterfaceError)
	}
	return out
}

func (g *GUIInteractive) OutputReadFilename(text, extension string) string {
	out := g.OutputReadString(text)
	if path.Ext(out) == "" {
		out += "." + extension
	}
	return out
}

func (g *GUIInteractive) OutputReadInt(text string, allowEmpty bool, validateCb func(value int) bool) int {
	for {

		str := g.OutputReadString(text)

		out, err := strconv.Atoi(str)
		if !(allowEmpty && str == "") {
			if err != nil {
				g.OutputWrite("Invalid Number")
				continue
			}
		}

		if validateCb != nil && !validateCb(out) {
			g.OutputWrite("Invalid value. Try again")
			continue
		}
		return out
	}
}

func (g *GUIInteractive) OutputReadUint64(text string, allowEmpty bool, validateCb func(value uint64) bool) uint64 {

	for {
		str := g.OutputReadString(text)

		out, err := strconv.ParseUint(str, 10, 64)
		if !(allowEmpty && str == "") {
			if err != nil {
				g.OutputWrite("Invalid Number")
				continue
			}
		}

		if validateCb != nil && !validateCb(out) {
			g.OutputWrite("Invalid value. Try again")
			continue
		}
		return out
	}
}

func (g *GUIInteractive) OutputReadFloat64(text string, allowEmpty bool, validateCb func(float64) bool) float64 {
	for {
		str := g.OutputReadString(text)

		out, err := strconv.ParseFloat(str, 64)
		if !(allowEmpty && str == "") {
			if err != nil {
				g.OutputWrite("Invalid Number")
				continue
			}
		}

		if validateCb != nil && !validateCb(out) {
			g.OutputWrite("Invalid value. Try again")
			continue
		}

		return out
	}
}

func (g *GUIInteractive) OutputReadAddress(text string) *addresses.Address {
	for {
		str := g.OutputReadString(text)
		address, err := addresses.DecodeAddr(str)
		if err != nil {
			g.OutputWrite("Invalid Address")
			continue
		}
		return address
	}
}

func (g *GUIInteractive) OutputReadBool(text string) bool {
	for {
		str := g.OutputReadString(text)
		if str == "y" {
			return true
		} else if str == "n" {
			return false
		}
		g.OutputWrite("Invalid boolean answer")
	}
}

func (g *GUIInteractive) OutputReadBytes(text string, validateCb func([]byte) bool) []byte {

	for {
		str := g.OutputReadString(text)
		input, err := hex.DecodeString(str)

		if err != nil {
			g.OutputWrite("Invalid Data. The input has to be a hex")
			continue
		}

		if validateCb != nil && !validateCb(input) {
			g.OutputWrite("Invalid value. Try again")
			continue
		}
		return input

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

	g.cmdChanges.Lock()
	g.cmdStatus = newCmdStatus
	g.cmdChanges.Unlock()
}

func (g *GUIInteractive) outputDone() {
	g.OutputWrite("------------------------")
	g.OutputWrite("Press space to return...")
	g.cmdChanges.Lock()
	g.cmdStatus = "output done"
	g.cmdChanges.Unlock()
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