package jiraconv

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andygrunwald/go-jira"
)

var (
	NobodyUser = &User{EmailAddress: "nobody@jira"}
)

type User struct {
	DisplayName  string
	Name         string
	EmailAddress string
}

func (o User) String() string {
	b, err := o.MarshalText()
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
	return string(b)
}

func (o User) MarshalText() ([]byte, error) {
	var s []string

	if len(o.DisplayName) > 0 {
		s = append(s, o.DisplayName)
	}

	if len(o.Name) > 0 {
		s = append(s, "("+o.Name+")")
	}

	if len(o.EmailAddress) > 0 {
		s = append(s, "<"+o.EmailAddress+">")
	}

	return []byte(strings.Join(s, " ")), nil
}

func (o *User) UnmarshalText(text []byte) error {
	re := regexp.MustCompile(`^\s*(?P<display>[^()<>]+)?(?:\((?P<name>[^()]+)\))?\s*(?:\<(?P<email>[^<>]+)\>)?\s*$`)
	s := re.FindStringSubmatch(string(text))

	o.DisplayName = strings.TrimSpace(s[1])
	o.Name = strings.TrimSpace(s[2])
	o.EmailAddress = strings.TrimSpace(s[3])

	return nil
}

func UserFromJira(data *jira.User) *User {
	if data == nil {
		return nil
	}
	return &User{
		DisplayName:  data.DisplayName,
		EmailAddress: data.EmailAddress,
		Name:         data.Name,
	}
}
