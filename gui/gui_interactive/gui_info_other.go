//go:build !darwin
// +build !darwin

package gui_interactive

import (
	"github.com/mackerelio/go-osstat/cpu"
	"strconv"
	"time"
)

func (g *GUIInteractive) cpuStatistics() {
	before, err1 := cpu.Get()
	time.Sleep(1 * time.Second)

	after, err2 := cpu.Get()
	if err1 == nil && err2 == nil {
		total := float64(after.Total - before.Total)
		g.InfoUpdate("cpu", strconv.FormatFloat(float64(after.User-before.User)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.System-before.System)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.Idle-before.Idle)/total*100, 'f', 2, 64)+"% ")
	}
}
