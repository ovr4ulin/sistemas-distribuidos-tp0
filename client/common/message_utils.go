package common

import "strings"

func matchTag(s, expected string) bool {
	parts := strings.SplitN(s, "^", 2)
	return len(parts) > 0 && parts[0] == expected
}
