package gui

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sort"
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
}
