package page

import (
	"io"
	"bufio"
	"bytes"
)

type Page struct {
	mustRender bool
}

func ReadFrom(r io.Reader) (page *Page, err error) {
	page = new(Page)

	reader := bufio.NewReader(r)
	c, err := reader.Peek(1)
	if err != nil {
		return
	}
	
	if !bytes.Equal(c, []byte("<")) {
		page.mustRender = true
	}
	return
}
