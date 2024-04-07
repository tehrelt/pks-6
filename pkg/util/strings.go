package util

import "strings"

func StripNull(s string) string {
	return strings.Split(s, "\x00")[0]
}
