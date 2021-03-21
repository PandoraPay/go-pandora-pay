package gui

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sort"
	"sync"
)

var info2 *widgets.List
var info2Map sync.Map

func info2Render() {

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
