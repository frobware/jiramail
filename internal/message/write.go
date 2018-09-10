package message

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/mail"
	"net/textproto"
	"sort"
)

func writeMIMEHeader(w io.Writer, header mail.Header) (N int, err error) {
	var names sort.StringSlice
	var n int

	for name := range header {
		names = append(names, name)
	}

	names.Sort()

	for _, name := range names {
		for _, value := range header[name] {
			n, err = io.WriteString(w, name+": "+value+"\n")
			N += n
			if err != nil {
				return
			}
		}
	}

	n, err = io.WriteString(w, "\n")
	N += n
	return
}

func writeMessage(w io.Writer, header mail.Header, body io.ReadSeeker) error {
	if _, err := body.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := writeMIMEHeader(w, header); err != nil {
		return err
	}
	if _, err := io.Copy(w, body); err != nil {
		return err
	}
	return nil
}

func MakeChecksum(m *mail.Message) (string, error) {
	body, ok := m.Body.(io.ReadSeeker)
	if !ok {
		return "", fmt.Errorf("unable to write such message")
	}

	hdr := make(textproto.MIMEHeader)
	for k, v := range m.Header {
		if k != "X-Checksum" {
			hdr[k] = v
		}
	}

	h := sha256.New()

	err := writeMessage(h, mail.Header(hdr), body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

func Write(w io.Writer, m *mail.Message) error {
	body, ok := m.Body.(io.ReadSeeker)
	if !ok {
		return fmt.Errorf("unable to write such message")
	}
	return writeMessage(w, m.Header, body)
}
