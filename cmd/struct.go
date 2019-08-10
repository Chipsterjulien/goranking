package cmd

import "time"

type thing struct {
	date     time.Time
	filename string
	path     string
	remove   bool
}

type byDate []thing

func (bd byDate) Len() int {
	return len(bd)
}

func (bd byDate) Less(i, j int) bool {
	return bd[i].date.Before(bd[j].date)
}

func (bd byDate) Swap(i, j int) {
	bd[i], bd[j] = bd[j], bd[i]
}
