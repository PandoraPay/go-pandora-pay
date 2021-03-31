package gui

import (
	"github.com/gizak/termui/v3/widgets"
	"github.com/mackerelio/go-osstat/cpu"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

var info *widgets.List
var infoMap sync.Map

func infoRender() {
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

	go func() {

		for {

			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			InfoUpdate("memory", bToMb(m.Alloc)+"M "+bToMb(m.TotalAlloc)+"M "+bToMb(m.Alloc)+"M "+strconv.FormatUint(uint64(m.NumGC), 10))

			before, err1 := cpu.Get()
			time.Sleep(1 * time.Second)
			after, err2 := cpu.Get()
			if err1 == nil && err2 == nil {
				total := float64(after.Total - before.Total)
				InfoUpdate("cpu", strconv.FormatFloat(float64(after.User-before.User)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.System-before.System)/total*100, 'f', 2, 64)+"% "+strconv.FormatFloat(float64(after.Idle-before.Idle)/total*100, 'f', 2, 64)+"% ")
			}

			time.Sleep(1 * time.Second)
		}

	}()

}

func bToMb(b uint64) string {
	mb := b / 1024 / 1024
	return strconv.FormatUint(mb, 10)
}
