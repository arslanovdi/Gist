package utils

import (
	"fmt"
	"time"
)

// FormatDurationShort Функция для форматирования duration в "месяцы дни часы минуты"
func FormatDurationShort(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	days := int64(d.Hours() / 24)
	months := days / 30 // Примерно месяц = 30 дней
	days = days % 30
	hours := int64(d.Hours()) % 24
	minutes := int64(d.Minutes()) % 60

	if months > 0 {
		return fmt.Sprintf("%dмес %dд %dч %dм", months, days, hours, minutes)
	}
	if days > 0 {
		return fmt.Sprintf("%dд %dч %dм", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dч %dм", hours, minutes)
	}

	return fmt.Sprintf("%dм", minutes)
}

// FormatDateShort Функция для даты-времени в "yyyy-mm-dd-hh-mm"
func FormatDateShort(t time.Time) string {
	return t.Format("2006-01-02 15ч04м")
}
