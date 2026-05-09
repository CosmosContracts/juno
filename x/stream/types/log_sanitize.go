package types

import "strings"

// SanitizeLog returns a copy of s with control characters (CR, LF, tab, etc.)
// replaced. It is meant for log fields whose value originates from a request
// header, subscription key, or other user-controlled source — passing such
// values to a structured logger unsanitised lets a caller forge log lines via
// embedded newlines (CWE-117 / log injection).
func SanitizeLog(s string) string {
	if s == "" {
		return s
	}
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || (r >= 0 && r < 0x20) || r == 0x7f {
			return '_'
		}
		return r
	}, s)
}

// SanitizeLogMap returns a new map with both keys and values passed through
// SanitizeLog. Use it when logging a user-controlled string→string map.
func SanitizeLogMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return m
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[SanitizeLog(k)] = SanitizeLog(v)
	}
	return out
}
