package utils

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// DebugMode is the global debug flag
var DebugMode bool

// Compile the regex pattern once at package initialization
var durationRegex = regexp.MustCompile(`^([+-])((\d+)D)?((\d+)h)?((\d+)m)?((\d+)s)?$`)

// IsPrefixIn checks if string s starts with any of the given prefixes
func IsPrefixIn(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// ParseDurationString parses a duration string and returns the total seconds
func ParseDurationString(durationStr string) (float64, bool) {
	matches := durationRegex.FindStringSubmatch(durationStr)

	if len(matches) == 0 {
		return 0, false
	}

	// Determine the sign of the duration
	sign := 1.0
	if matches[1] == "-" {
		sign = -1.0
	}

	// Parse each matched group and calculate total seconds
	var totalSeconds float64 = 0

	// Days
	if matches[3] != "" {
		days, err := strconv.ParseFloat(matches[3], 64)
		if err != nil {
			log.Printf("Warning: failed to parse days value '%s': %v", matches[3], err)
		} else {
			totalSeconds += days * 86400 // 86400 seconds in a day
		}
	}

	// Hours
	if matches[5] != "" {
		hours, err := strconv.ParseFloat(matches[5], 64)
		if err != nil {
			log.Printf("Warning: failed to parse hours value '%s': %v", matches[5], err)
		} else {
			totalSeconds += hours * 3600 // 3600 seconds in an hour
		}
	}

	// Minutes
	if matches[7] != "" {
		minutes, err := strconv.ParseFloat(matches[7], 64)
		if err != nil {
			log.Printf("Warning: failed to parse minutes value '%s': %v", matches[7], err)
		} else {
			totalSeconds += minutes * 60 // 60 seconds in a minute
		}
	}

	// Seconds
	if matches[9] != "" {
		seconds, err := strconv.ParseFloat(matches[9], 64)
		if err != nil {
			log.Printf("Warning: failed to parse seconds value '%s': %v", matches[9], err)
		} else {
			totalSeconds += seconds
		}
	}

	// Apply the sign
	totalSeconds *= sign

	return totalSeconds, true
}

// DebugLog logs debug messages if DebugMode is enabled
func DebugLog(format string, args ...interface{}) {
	if DebugMode {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// SanitizeMetricName sanitizes metric names for Prometheus
func SanitizeMetricName(name string) string {
	// Replace invalid characters with underscores
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "+", "_")
	return result
}
