package gui_interactive

import (
	"github.com/gizak/termui/v3/widgets"
	"sort"
)

func (g *GUIInteractive) info2Render() {

	rows := []string{}
	g.info2Map.Range(func(key, value interface{}) bool {
		rows = append(rows, key.(string)+": "+value.(string))
		return true
	})
	sort.Strings(rows)
	g.info2.Lock()
	g.info2.Rows = rows
	g.info2.Unlock()
}

func (g *GUIInteractive) Info2Update(key string, text string) {
	if text == "" {
		g.info2Map.Delete(key)
		return
	}
	g.info2Map.Store(key, text)
}

func (g *GUIInteractive) info2Init() {
	g.info2 = widgets.NewList()
	g.info2.Title = "Statistics"
}
