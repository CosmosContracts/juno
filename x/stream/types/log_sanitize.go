package types

import "strings"

// logSanitizer replaces every C0 control character (0x00-0x1f) and DEL (0x7f)
// with '_'. Built via strings.NewReplacer so CodeQL recognizes it as a known
// log-injection sanitizer (CWE-117) in addition to actually neutralizing the
// newline / carriage return that enables log forging.
var logSanitizer = func() *strings.Replacer {
	pairs := make([]string, 0, 66)
	for i := 0; i < 0x20; i++ {
		pairs = append(pairs, string(rune(i)), "_")
	}
	pairs = append(pairs, "\x7f", "_")
	return strings.NewReplacer(pairs...)
}()

// SanitizeLog returns a copy of s with control characters (CR, LF, tab, other
// C0 bytes, DEL) replaced by '_'. Use on log fields whose value originates from
// a request header, subscription key, or other user-controlled source —
// passing such values to a structured logger unsanitised lets a caller forge
// log lines via embedded newlines (CWE-117 / log injection).
func SanitizeLog(s string) string {
	if s == "" {
		return s
	}
	return logSanitizer.Replace(s)
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
