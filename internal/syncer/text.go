package syncer

import (
	"regexp"
	"strings"
)

var (
	re = regexp.MustCompile(`[\s/-]+`)
)

func ReplaceStringTrash(s string) string {
	return re.ReplaceAllString(strings.TrimSpace(s), " ")
}
