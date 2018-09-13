package message

import (
	"fmt"
	"net/mail"
	"strings"
)

type Address struct {
	Name    string
	Address string
	Tags    map[string]struct{}
}

func (a *Address) Mail() *mail.Address {
	return &mail.Address{
		Name:    a.Name,
		Address: a.Address,
	}
}

func parseMailAddress(m *mail.Address) (a *Address, err error) {
	a = &Address{
		Name: m.Name,
		Tags: make(map[string]struct{}),
	}

	components := strings.Split(m.Address, "@")
	if len(components) > 2 {
		err = fmt.Errorf("only one '@' expected: %s", m.Address)
		return
	}

	recipient := strings.Split(components[0], "+")

	a.Address = recipient[0] + "@" + components[1]
	for _, name := range recipient[1:] {
		a.Tags[strings.ToLower(name)] = struct{}{}
	}

	return
}

func ParseAddress(s string) (*Address, error) {
	m, err := mail.ParseAddress(s)
	if err != nil {
		return nil, err
	}

	return parseMailAddress(m)
}

func ParseAddressList(s string) ([]*Address, error) {
	var ret []*Address

	list, err := mail.ParseAddressList(s)
	if err != nil {
		return nil, err
	}

	for _, m := range list {
		a, err := parseMailAddress(m)
		if err != nil {
			return nil, err
		}
		ret = append(ret, a)
	}

	return ret, nil
}
