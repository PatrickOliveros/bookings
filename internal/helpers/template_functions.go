package helpers

import (
	"time"
)

func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func CalendarDate(t time.Time) string {
	return t.Format("January 2, 2006")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func Iterate(count int) []int {
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i)
	}

	return items
}

func Add(a, b int) int {
	return a + b
}
