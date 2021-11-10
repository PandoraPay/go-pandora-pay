package gui_interactive

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"pandora-pay/gui/gui_interface"
	"path"
	"strconv"
	"sync"
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
	Callback func(string, context.Context) error
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
	{Name: "Wallet:TX", Text: "Private Claim"},
	{Name: "Wallet:TX", Text: "Private Asset Create"},
	{Name: "Wallet:TX", Text: "Private Asset Supply Increase"},
	{Name: "Wallet:TX", Text: "Update Delegate"},
	{Name: "Wallet:TX", Text: "Unstake"},
	{Name: "Wallet:TX", Text: "Update Asset Fee Liquidity"},
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
var commandsLock sync.Mutex

func (g *GUIInteractive) CommandDefineCallback(Text string, callback func(string, context.Context) error, useIt bool) {

	if !useIt {
		callback = nil
	}

	commandsLock.Lock()
	defer commandsLock.Unlock()

	for i := range commands {
		if commands[i].Text == Text {
			commands[i].Callback = callback
			return
		}
	}

	g.Error(errors.New("Command " + Text + " was not found"))
}

func (g *GUIInteractive) cmdProcess(e ui.Event) {

	cmdData := g.cmdData.Load().(*GUIInteractiveData)

	var command *Command
	if cmdData.cmdStatus == "cmd" {
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
		if g.cmd.SelectedRow >= len(g.cmd.Rows) {
			g.cmd.SelectedRow = len(g.cmd.Rows) - 1
		}
		if g.cmd.SelectedRow < 0 {
			g.cmd.SelectedRow = 0
		}

		g.cmd.Unlock()
	}

	switch e.ID {
	case "<C-c>":
		if cmdData.cmdStatus == "read" {
			cmdData.cmdStatusCtxCancel()
			return
		}
		g.Close()
	case "<Enter>":

		if cmdData.cmdStatus == "cmd" {

			if command != nil && command.Callback != nil {

				ctx, cancel := context.WithCancel(context.Background())

				g.outputClear("", nil, ctx, cancel)

				g.OutputWrite(fmt.Sprintf("Executing cmd %s::%s ...", command.Name, command.Text))
				g.OutputWrite("")

				go func() {

					defer cancel()

					defer func() {
						err := recover()
						if err != nil {
							g.Error(err)
							g.outputDone()
						}
					}()

					if err := command.Callback(command.Text, ctx); err != nil {
						g.OutputWrite(fmt.Sprintf("Error: %s", err.Error()))

						g.cmdData.Store(&GUIInteractiveData{
							cmdStatus: "output done",
						})

					} else {
						g.outputDone()
					}

				}()
			}
		} else if cmdData.cmdStatus == "output done" {
			g.outputRestore()
		} else if cmdData.cmdStatus == "read" {
			if cmdData.cmdInputCn != nil {
				cmdData.cmdInputCn <- cmdData.cmdInput
			}
		}

	}

	if cmdData.cmdStatus == "read" && !notAcceptedCharacters[e.ID] {

		str := cmdData.cmdInput

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

		g.cmdData.Store(&GUIInteractiveData{
			cmdData.cmdStatus,
			str,
			cmdData.cmdInputCn,
			cmdData.cmdStatusCtx,
			cmdData.cmdStatusCtxCancel,
		})

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

func (g *GUIInteractive) outputRead(text string) (<-chan string, context.Context) {

	g.cmd.Lock()
	g.cmd.Rows = append(g.cmd.Rows, "")
	g.cmd.Rows = append(g.cmd.Rows, text)
	g.cmd.Rows = append(g.cmd.Rows, "-> ")
	g.cmd.SelectedRow = len(g.cmd.Rows) - 1
	g.cmd.Unlock()

	cn := make(chan string)
	cmdData := g.cmdData.Load().(*GUIInteractiveData)

	g.cmdData.Store(&GUIInteractiveData{
		"read",
		"",
		cn,
		cmdData.cmdStatusCtx,
		cmdData.cmdStatusCtxCancel,
	})

	return cn, cmdData.cmdStatusCtx
}

func (g *GUIInteractive) OutputReadString(text string) string {

	dataCn, ctx := g.outputRead(text)

	select {
	case out, ok := <-dataCn:
		if !ok {
			panic(gui_interface.GUIInterfaceError)
		}
		return out
	case <-ctx.Done():
		panic(gui_interface.GUIInterfaceError)
	}

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

func (g *GUIInteractive) outputClear(newCmdStatus string, rows []string, ctx context.Context, cancel context.CancelFunc) {
	if rows == nil {
		rows = []string{}
	}

	g.cmd.Lock()
	g.cmd.Rows = rows
	g.cmd.SelectedRow = 0
	g.cmd.Unlock()

	g.cmdData.Store(&GUIInteractiveData{
		newCmdStatus,
		"", nil, ctx, cancel,
	})

}

func (g *GUIInteractive) outputDone() {
	g.OutputWrite("------------------------")
	g.OutputWrite("Press space to return...")

	g.cmdData.Store(&GUIInteractiveData{
		"output done",
		"", nil, nil, nil,
	})

}

func (g *GUIInteractive) outputRestore() {
	g.outputClear("cmd", g.cmdRows, nil, nil)
}

func (g *GUIInteractive) cmdInit() {

	g.cmdData.Store(&GUIInteractiveData{
		cmdStatus: "cmd",
		cmdInput:  "",
	})

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
