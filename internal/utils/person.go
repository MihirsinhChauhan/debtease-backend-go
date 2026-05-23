package utils

import "strings"

// JoinPersonName builds a single display string from trimmed first / last tokens.
func JoinPersonName(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	switch {
	case first == "" && last == "":
		return ""
	case last == "":
		return first
	case first == "":
		return last
	default:
		return first + " " + last
	}
}
