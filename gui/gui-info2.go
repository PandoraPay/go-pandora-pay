package gui

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"sort"
	"strconv"
	"sync"
	"time"
)

var info2 *widgets.List
var info2Map sync.Map

func info2Process() {
	Info2Update("time", time.Now().Format("2006.01.02 15:04:05"))

	rows := []string{}
	info2Map.Range(func(key, value interface{}) bool {
		rows = append(rows, key.(string)+": "+value.(string))
		return true
	})
	sort.Strings(rows)
	info2.Rows = rows
	ui.Render(info, info2)
}

func Info2Update(key string, text string) {
	info2Map.Store(key, text)
}

func info2Init() {
	info2 = widgets.NewList()
	info2.Title = "Statistics"

	go func() {

		for {

			memory, err := memory.Get()
			if err == nil {
				Info2Update("memory", bToMb(memory.Used)+"M "+bToMb(memory.Total)+"M "+bToMb(memory.Cached)+"M "+bToMb(memory.Free)+"M")
			}

			before, err1 := cpu.Get()
			time.Sleep(time.Duration(1) * time.Second)
			after, err2 := cpu.Get()
			if err1 == nil && err2 == nil {
				total := float64(after.Total - before.Total)
				Info2Update("cpu", strconv.FormatFloat(float64(after.User-before.User)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.System-before.System)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.Idle-before.Idle)/total*100, 'f', 2, 64)+"% ")
			}

		}

	}()

}

func bToMb(b uint64) string {
	mb := b / 1024 / 1024
	return strconv.FormatUint(mb, 10)
}
