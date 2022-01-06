package gui_interactive

import (
	"github.com/gizak/termui/v3/widgets"
	"github.com/mackerelio/go-osstat/cpu"
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

		ticker := time.NewTicker(1 * time.Second).C

		for {

			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			g.InfoUpdate("memory", bToMb(m.Alloc)+"M "+bToMb(m.TotalAlloc)+"M "+bToMb(m.Alloc)+"M "+strconv.FormatUint(uint64(m.NumGC), 10))

			before, err1 := cpu.Get()
			<-ticker

			after, err2 := cpu.Get()
			if err1 == nil && err2 == nil {
				total := float64(after.Total - before.Total)
				g.InfoUpdate("cpu", strconv.FormatFloat(float64(after.User-before.User)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.System-before.System)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.Idle-before.Idle)/total*100, 'f', 2, 64)+"% ")
			}

			<-ticker
		}

	}()

}

func bToMb(b uint64) string {
	mb := b / 1024 / 1024
	return strconv.FormatUint(mb, 10)
}
