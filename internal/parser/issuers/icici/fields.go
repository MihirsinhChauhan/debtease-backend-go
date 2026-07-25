package icici

import (
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	amountPattern = regexp.MustCompile(`[\d,]+(?:\.\d{1,2})?`)
	datePatterns  = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\d{1,2}[-/][A-Za-z]{3}[-/]\d{2,4})`),
		regexp.MustCompile(`(\d{1,2}/\d{1,2}/\d{4})`),
		regexp.MustCompile(`(\d{1,2}-\d{1,2}-\d{4})`),
	}
	last4Pattern = regexp.MustCompile(`(?i)(?:X{4}[\s-]*){2,3}(\d{4})|ending\s+(\d{4})|(?:\d{4}[\s-]){3}(\d{4})|\d{4}X{4,}(\d{4})`)
	longFormDate = regexp.MustCompile(`(?i)([A-Za-z]+)\s+(\d{1,2}),\s+(\d{4})`)
)

func ParseAmount(raw string) (decimal.Decimal, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, false
	}
	m := amountPattern.FindString(raw)
	if m == "" {
		return decimal.Zero, false
	}
	m = strings.ReplaceAll(m, ",", "")
	v, err := decimal.NewFromString(m)
	if err != nil {
		return decimal.Zero, false
	}
	return v, true
}

func ParseDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"02-Jan-2006", "2-Jan-2006", "02-Jan-06", "2-Jan-06",
		"02/01/2006", "2/1/2006", "02-01-2006", "2-1-2006",
		"January 2, 2006", "Jan 2, 2006", "January 02, 2006", "Jan 02, 2006",
	}
	for _, re := range datePatterns {
		if m := re.FindString(raw); m != "" {
			raw = m
			break
		}
	}
	if m := longFormDate.FindStringSubmatch(raw); len(m) == 4 {
		raw = m[1] + " " + m[2] + ", " + m[3]
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func FindLabelValue(text, label string) string {
	lines := strings.Split(text, "\n")
	labelLower := strings.ToLower(label)
	for i, line := range lines {
		lineLower := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(lineLower, labelLower) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				return strings.TrimSpace(parts[1])
			}
			for j := i + 1; j < len(lines) && j <= i+2; j++ {
				next := strings.TrimSpace(lines[j])
				if next != "" {
					return next
				}
			}
		}
	}
	return ""
}

func FindLast4(text string) string {
	matches := last4Pattern.FindStringSubmatch(text)
	for _, g := range matches[1:] {
		if g != "" {
			return g
		}
	}
	return ""
}

func CollectLast4s(text string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range last4Pattern.FindAllStringSubmatch(text, -1) {
		for _, g := range m[1:] {
			if g == "" {
				continue
			}
			if _, ok := seen[g]; ok {
				continue
			}
			seen[g] = struct{}{}
			out = append(out, g)
		}
	}
	return out
}
