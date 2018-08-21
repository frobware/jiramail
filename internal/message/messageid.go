package message

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/textproto"
	"strings"
)

func EncodeMessageID(t string, a map[string]string) string {
	b, err := json.Marshal(a)
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
	return fmt.Sprintf("<%s@%s>", base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(b), t)
}

func DecodeMessageID(s string, header textproto.MIMEHeader) error {
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
		s = s[1 : len(s)-1]
	}

	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return fmt.Errorf("only one character '@' expected: %s", s)
	}

	domain := strings.Split(parts[1], ".")
	if len(domain) != 2 {
		return fmt.Errorf("only one character '.' expected in the domain: %s", s)
	}

	data, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(parts[0])
	if err != nil {
		return err
	}

	var a map[string]string

	err = json.Unmarshal(data, &a)
	if err != nil {
		return err
	}

	for k, v := range a {
		header.Set(fmt.Sprintf("X-%s-%s", domain[0], k), v)
	}

	header.Set("X-Type", strings.ToLower(domain[0]))

	return nil
}
