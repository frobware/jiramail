package command

import (
	"fmt"
	"net/textproto"
	"strings"
)

const (
	JiraStart = "{{{ jira"
	JiraEnd   = "}}}"
)

type Mail struct {
	Header textproto.MIMEHeader
	Body   []string
}

func GetBody(msg *Mail) string {
	var x []string

	ignore := false
	for _, s := range msg.Body {
		switch {
		case s == JiraStart || strings.HasSuffix(s, " "+JiraStart):
			ignore = true
		case s == JiraEnd || strings.HasSuffix(s, " "+JiraEnd):
			ignore = false
		case strings.HasPrefix(s, "#") || strings.HasPrefix(s, ">"):
			// ignore
		default:
			if !ignore {
				if s == "\n" {
					s = ""
				}
				x = append(x, s)
			}
		}
	}
	return strings.TrimSpace(strings.Join(x, "\n"))
}

func GetJiraBlock(msg *Mail) []string {
	var x []string

	ignore := true
	for _, s := range msg.Body {
		s = strings.TrimSpace(s)
		switch {
		case s == JiraStart || strings.HasSuffix(s, " "+JiraStart):
			ignore = false
		case s == JiraEnd || strings.HasSuffix(s, " "+JiraEnd):
			ignore = true
		default:
			if !ignore && len(s) > 0 {
				x = append(x, s)
			}
		}
	}
	return x
}

func MakeJiraBlock(a [][]string) string {
	h := JiraStart + "\n"
	h += "# This block will be automatically deleted from the text.\n"

	n := 0
	for _, s := range a {
		if n < len(s[0]) {
			n = len(s[0])
		}
	}

	if n > 0 {
		h += "#\n"
	}

	for _, s := range a {
		h += fmt.Sprintf(fmt.Sprintf("# %%-%ds: %%s\n", n), s[0], s[1])
	}

	if n > 0 {
		h += "#\n"
	}

	h += JiraEnd + "\n\n"
	return h
}
