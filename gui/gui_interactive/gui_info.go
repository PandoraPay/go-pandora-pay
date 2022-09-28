package gui_interactive

import (
	"github.com/gizak/termui/v3/widgets"
	"runtime"
	"sort"
	"strconv"
	"time"
)

func (g *GUIInteractive) infoRender() {
	rows := []string{}
	g.infoMap.Range(func(key, value string) bool {
		rows = append(rows, key+": "+value)
		return true
	})
	sort.Strings(rows)
	g.info.Lock()
	g.info.Rows = rows
	g.info.Unlock()
}

func (g *GUIInteractive) InfoUpdate(key string, text string) {
	if text == "" {
		g.infoMap.Delete(key)
		return
	}
	g.infoMap.Store(key, text)
}

func (g *GUIInteractive) infoInit() {

	g.info = widgets.NewList()
	g.info.Title = "Node Info"

	go func() {

		for {

			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			g.InfoUpdate("memory", bToMb(m.Alloc)+"M "+bToMb(m.TotalAlloc)+"M "+bToMb(m.Alloc)+"M "+strconv.FormatUint(uint64(m.NumGC), 10))

			g.cpuStatistics()

			time.Sleep(1 * time.Second)
		}

	}()

}

func bToMb(b uint64) string {
	mb := b / 1024 / 1024
	return strconv.FormatUint(mb, 10)
}
