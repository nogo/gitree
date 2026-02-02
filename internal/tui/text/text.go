// Package text provides text manipulation utilities for TUI rendering.
// Centralizes ANSI-aware string operations to avoid duplication.
package text

import "strings"

// Width returns the display width of a string, excluding ANSI escape codes.
func Width(s string) int {
	count := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

// Strip removes ANSI escape codes from a string.
func Strip(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// Truncate truncates a plain string to max runes, adding ellipsis if needed.
func Truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// TruncateAnsi truncates a string with ANSI codes to max display width.
func TruncateAnsi(s string, max int) string {
	if max <= 0 {
		return ""
	}

	var result strings.Builder
	displayCount := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if displayCount >= max {
			break
		}
		result.WriteRune(r)
		displayCount++
	}

	return result.String()
}

// PadRight pads a string on the right to reach the target width.
func PadRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// PadLeft pads a string on the left to reach the target width (right-align).
func PadLeft(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(runes)) + s
}

// PadRightAnsi pads a string with ANSI codes on the right to reach target display width.
func PadRightAnsi(s string, width int) string {
	if width <= 0 {
		return ""
	}
	displayLen := Width(s)
	if displayLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-displayLen)
}

// Fit truncates or pads a plain string to exactly the target width.
func Fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) == width {
		return s
	}
	if len(runes) > width {
		return Truncate(s, width)
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// FitAnsi truncates or pads a string with ANSI codes to exactly the target display width.
func FitAnsi(s string, width int) string {
	if width <= 0 {
		return ""
	}
	displayLen := Width(s)
	if displayLen == width {
		return s
	}
	if displayLen > width {
		return TruncateAnsi(s, width)
	}
	return s + strings.Repeat(" ", width-displayLen)
}

// FitLeft truncates or left-pads a plain string to exactly the target width (right-align).
func FitLeft(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) == width {
		return s
	}
	if len(runes) > width {
		return string(runes[len(runes)-width:])
	}
	return strings.Repeat(" ", width-len(runes)) + s
}

