package cmd

import (
	"fmt"
)

// Format the duration in minutes as "X h Y m".
func formatTimeMinutesToHours(minutes int) string {
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%d h %d m", hours, minutes)
}

// Format the duration in seconds as "X h Y m".
func formatTimeSecondsToHours(minutes int) string {
	hours := minutes / 3600
	minutes = minutes % 3600
	return fmt.Sprintf("%d h %d m", hours, minutes)
}
