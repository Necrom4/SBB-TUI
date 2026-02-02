// Package utils
package utils

import "fmt"

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func RenderLink(text, url string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}
