package message

import (
	"fmt"
	"net/mail"
	"os"
	"strings"
)

func HeaderID(m *mail.Message) string {
	id := m.Header.Get("Message-ID")

	if strings.HasPrefix(id, "<") && strings.HasSuffix(id, ">") {
		return id[1 : len(id)-1]
	}

	if len(id) == 0 {
		panic(fmt.Sprintf("empty or not found header: Message-ID\n\n%#+v\n", m))
	}

	return id
}

func HeaderChecksum(m *mail.Message) string {
	hash := m.Header.Get("X-Checksum")

	if len(hash) == 0 {
		panic(fmt.Sprintf("empty or not found header: Message-ID\n\n%#+v\n", m))
	}

	return hash
}

func GetChecksum(fp string) (string, error) {
	fd, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	m, err := mail.ReadMessage(fd)
	if err != nil {
		return "", err
	}

	return HeaderChecksum(m), nil
}
