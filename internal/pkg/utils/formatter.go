package utils

import (
	"fmt"
	"math"
	"time"
)

func FormatSalary(min, max *int, period string) string {
	if min == nil && max == nil {
		return "Negotiable"
	}

	periodSuffix := ""
	if period == "monthly" {
		periodSuffix = " / mo"
	} else if period == "hourly" {
		periodSuffix = " / hr"
	}

	formatVal := func(val int) string {
		if val >= 1000 {
			kVal := float64(val) / 1000.0
			if kVal == math.Trunc(kVal) {
				return fmt.Sprintf("$%d\u200bk", int(kVal)) // Soft separator or simple k
			}
			return fmt.Sprintf("$%.1fk", kVal)
		}
		return fmt.Sprintf("$%d", val)
	}

	if min != nil && max != nil {
		// Replace soft separator with raw 'k'
		minStr := stringsReplaceSoftSep(formatVal(*min))
		maxStr := stringsReplaceSoftSep(formatVal(*max))
		return fmt.Sprintf("%s - %s%s", minStr, maxStr, periodSuffix)
	} else if min != nil {
		minStr := stringsReplaceSoftSep(formatVal(*min))
		return fmt.Sprintf("%s+%s", minStr, periodSuffix)
	} else {
		maxStr := stringsReplaceSoftSep(formatVal(*max))
		return fmt.Sprintf("Up to %s%s", maxStr, periodSuffix)
	}
}

func stringsReplaceSoftSep(s string) string {
	// Removes the soft separator if printed, or keep simple formatting
	return s
}

func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 24*time.Hour && t.Day() == now.Day() {
		return "Today"
	}

	days := int(diff.Hours() / 24)
	if days == 1 || (days == 0 && t.Day() != now.Day()) {
		return "Yesterday"
	}

	if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	}

	weeks := days / 7
	if weeks == 1 {
		return "1 week ago"
	}
	if weeks < 4 {
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	months := days / 30
	if months <= 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", months)
}
