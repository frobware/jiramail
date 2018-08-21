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
	if _, err := writeMIMEHeader(w, header); err != nil {
		return err
	}
	if _, err := io.Copy(w, body); err != nil {
		return err
	}
	return nil
}

func Write(w io.Writer, m *mail.Message) (string, error) {
	readerSeeker, ok := m.Body.(io.ReadSeeker)
	if !ok {
		return "", fmt.Errorf("unable to write such message")
	}

	_, err := readerSeeker.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	hash := sha256.New()

	err = writeMessage(hash, m.Header, readerSeeker)
	if err != nil {
		return "", err
	}

	chksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))

	hdr := textproto.MIMEHeader(m.Header)
	hdr.Set("X-Checksum", chksum)

	if _, err = readerSeeker.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	if err = writeMessage(w, mail.Header(hdr), readerSeeker); err != nil {
		return "", err
	}

	return chksum, err
}
