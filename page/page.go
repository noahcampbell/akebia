package page

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode"
)

const (
	HTML_LEAD       = "<"
	YAML_LEAD       = "-"
	YAML_DELIM_UNIX = "---\n"
	YAML_DELIM_DOS  = "---\r\n"
)

type Page struct {
	render      bool
	frontmatter []byte
}

type FrontMatter []byte

func ReadFrom(r io.Reader) (page *Page, err error) {
	reader := bufio.NewReader(r)

	if err = chompWhitespace(reader); err != nil {
		return
	}

	firstFive := make([]byte, 5)
	if _, err = reader.Read(firstFive); err != nil {
		return
	}

	page = new(Page)
	page.render = shouldRender(firstFive)

	if page.render && isFrontMatterDelim(firstFive) {
		fm, err := extractFrontMatter(reader)
		if err != nil {
			return nil, err
		}
		page.frontmatter = fm
	}

	return
}

func chompWhitespace(r io.RuneScanner) (err error) {
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			return err
		}
		if !unicode.IsSpace(c) {
			r.UnreadRune()
			return nil
		}
	}
	return
}

func shouldRender(lead []byte) (frontmatter bool) {
	if len(lead) <= 0 {
		return
	}

	if bytes.Equal(lead[:1], []byte(HTML_LEAD)) {
		return
	}
	return true
}

func isFrontMatterDelim(data []byte) bool {
	if !bytes.Equal(data[:1], []byte(YAML_LEAD)) {
		return false
	}

	if len(data) >= 4 && bytes.Equal(data[:4], []byte(YAML_DELIM_UNIX)) {
		return true
	}

	if len(data) >= 5 && bytes.Equal(data[:5], []byte(YAML_DELIM_DOS)) {
		return true
	}

	return false
}

// extractFrontMatter looks for the --- sequence followed by a newline
func extractFrontMatter(r io.Reader) (fm FrontMatter, err error) {
	buf := make([]byte, 1024) // TODO make this not a fixed buffer
	if _, err = r.Read(buf); err != nil {
		return nil, err
	}

	// strip off front matter delim if it's present.
	if isFrontMatterDelim(buf) {
		buf = buf[4:]
	}

	for i, c := range buf {
		switch c {
		case '\r':
			continue
		case '\n':
			if isFrontMatterDelim(buf[i+1:]) {
				return FrontMatter(buf[:i]), nil
			}
		}
	}
	return nil, errors.New("Could not find Front Matter.")
}
