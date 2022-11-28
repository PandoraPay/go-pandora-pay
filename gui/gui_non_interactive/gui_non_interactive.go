package gui_non_interactive

import (
	"context"
	"pandora-pay/gui/gui_logger"
	"runtime"
	"sync"
)

type GUINonInteractive struct {
	logger       *gui_logger.GUILogger
	colorError   string
	colorWarning string
	colorInfo    string
	colorLog     string
	colorFatal   string
	writingMutex sync.Mutex
}

func (g *GUINonInteractive) Close() {
}

func CreateGUINonInteractive() (*GUINonInteractive, error) {

	g := &GUINonInteractive{}

	switch runtime.GOARCH {
	default:
		g.colorError = "\x1b[31m"
		g.colorWarning = "\x1b[32m"
		g.colorInfo = "\x1b[34m"
		g.colorLog = "\x1b[37m"
		g.colorFatal = "\x1b[31m\x1b[43m"
	}

	return g, nil
}

func (g *GUINonInteractive) InfoUpdate(key string, text string) {
}

func (g *GUINonInteractive) Info2Update(key string, text string) {
}

func (g *GUINonInteractive) OutputWrite(any ...interface{}) {
}

func (g *GUINonInteractive) CommandDefineCallback(Text string, callback func(string, context.Context) error, useIt bool) {
}

func (g *GUINonInteractive) OutputReadBool(text string, allowEmpty bool, emptyValue bool) bool {
	return false
}
func (g *GUINonInteractive) OutputReadBytes(text string, validateCb func([]byte) bool) (data []byte) {
	return
}
func (g *GUINonInteractive) OutputReadFilename(text, extension string, allowEmpty bool) string {
	return ""
}
func (g *GUINonInteractive) OutputReadInt(text string, allowEmpty bool, emptyValue int, validateCb func(int) bool) int {
	return 0
}
func (g *GUINonInteractive) OutputReadUint64(text string, allowEmpty bool, emptyValue uint64, validateCb func(uint64) bool) uint64 {
	return 0
}
func (g *GUINonInteractive) OutputReadFloat64(text string, allowEmpty bool, emptyValue float64, validateCb func(float64) bool) float64 {
	return 0
}
func (g *GUINonInteractive) OutputReadString(text string) string {
	return ""
}
