package page

import (
	"io"
	"bufio"
)

type Page struct {
}

func ReadFrom(r io.Reader) (page *Page, err error) {
	reader := bufio.NewReader(r)
	_, err = reader.Peek(1)
	if err != nil {
		return
	}
	return
}
