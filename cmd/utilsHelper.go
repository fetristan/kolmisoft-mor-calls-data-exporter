package cmd

import (
	"fmt"
	"strings"
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

// Remove the leading zeros from the destination number.
func removeZero(number string) string {
	if strings.HasPrefix(number, "000") {
		number = number[3:]
	} else if strings.HasPrefix(number, "0") {
		number = "33" + number[1:]
	}
	return number
}
