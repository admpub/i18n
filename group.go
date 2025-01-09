package i18n

import "strings"

func TrimGroupPrefix(format string) string {
	if len(format) > 1 && format[0] == '#' {
		s := format[1:]
		parts := strings.SplitN(s, `#`, 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return format
}
