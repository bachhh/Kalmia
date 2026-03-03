package utils

// NOTE: just good enough, days duration are ambiguous due to timezone and leap seconds.
func DaysToSecond(days int) int {
	return 60 * 60 * 24 * days
}
