package gui

import (
	"github.com/gizak/termui/v3/widgets"
	"sort"
	"sync"
)

var info *widgets.List
var infoMap sync.Map

func infoProcess() {
	rows := []string{}
	infoMap.Range(func(key, value interface{}) bool {
		rows = append(rows, key.(string)+": "+value.(string))
		return true
	})
	sort.Strings(rows)
	info.Rows = rows
}

func InfoUpdate(key string, text string) {
	infoMap.Store(key, text)
}

func infoInit() {
	info = widgets.NewList()
	info.Title = "Node Info"
}
