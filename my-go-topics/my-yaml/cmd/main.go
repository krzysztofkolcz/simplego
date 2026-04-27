package main

import (
	"fmt"
	"time"
)

func main() {
	var t int64
	t = 1806566033
	fmt.Printf(formatDate(&t))
}

func formatDate(ts *int64) string {
	if ts == nil {
		return "-"
	}

	t := time.Unix(*ts, 0)

	// Format: DD.MM.YYYY HH:mm (typowo polski)
	return t.Format("02.01.2006 15:04")
}
