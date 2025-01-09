package i18n

import "strings"

func TrimGroupPrefix(format string) string {
	if len(format) > 1 && format[0] == '#' {
		format = format[1:]
		if pos := strings.Index(format, `#`); pos > -1 && pos < len(format)-1 {
			format = format[pos+1:]
		}
	}
	return format
}
