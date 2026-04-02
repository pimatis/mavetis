package output

import (
	"os"
	"strings"
)

type palette struct {
	critical string
	high     string
	medium   string
	low      string
	label    string
	muted    string
	reset    string
}

func colors() palette {
	enabled := paletteEnabled()
	if !enabled {
		return palette{}
	}
	return palette{
		critical: "\033[1;31m",
		high:     "\033[31m",
		medium:   "\033[33m",
		low:      "\033[34m",
		label:    "\033[36m",
		muted:    "\033[90m",
		reset:    "\033[0m",
	}
}

func paletteEnabled() bool {
	force := strings.TrimSpace(os.Getenv("FORCE_COLOR"))
	if force != "" {
		if force != "0" {
			return true
		}
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	term := strings.TrimSpace(os.Getenv("TERM"))
	if term == "" {
		return false
	}
	if term == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	return true
}

func paint(code string, value string, tone palette) string {
	if code == "" {
		return value
	}
	return code + value + tone.reset
}

func severityColor(severity string, tone palette) string {
	if severity == "critical" {
		return tone.critical
	}
	if severity == "high" {
		return tone.high
	}
	if severity == "medium" {
		return tone.medium
	}
	if severity == "low" {
		return tone.low
	}
	return tone.label
}
