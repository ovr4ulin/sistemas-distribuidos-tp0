package common

import "strings"

const FieldDelimiter = "^"
const RecordDelimiter = "~"

func matchTag(s, expected string) bool {
	parts := strings.SplitN(s, FieldDelimiter, 2)
	return len(parts) > 0 && parts[0] == expected
}
