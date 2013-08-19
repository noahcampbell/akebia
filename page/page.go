package page

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"launchpad.net/goyaml"
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
	frontmatter FrontMatter
	parsedFM    map[interface{}]interface{}
}

type FrontMatter []byte

func (p *Page) Property(key string) (value string, ok bool) {
	err := p.parseFM()
	if err != nil {
		panic(err) // TODO not go crazy if there is an error
	}
	if value, ok := p.parsedFM[key]; ok {
		return value.(string), ok
	}

	return "", false
}

func ReadFrom(r io.Reader) (page *Page, err error) {
	reader := bufio.NewReader(r)

	if err = chompWhitespace(reader); err != nil {
		return
	}

	firstFive := make([]byte, 4)
	if _, err = reader.Read(firstFive); err != nil {
		return
	}

	if firstFive[len(firstFive)-1] == '\r' {
		c, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		firstFive = append(firstFive, c)
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

	err = page.parseFM()
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
		buf = buf[bytes.Index(buf, []byte{'\n'})+1:]
	}

	for i, c := range buf {
		switch c {
		case '\r':
			continue
		case '\n':
			if isFrontMatterDelim(buf[i+1:]) {
				var chop int
				if buf[i-1] == '\r' {
					chop = 1
				}
				return FrontMatter(buf[:i-chop]), nil
			}
		}
	}
	return nil, errors.New("Could not find Front Matter.")
}

func (p *Page) parseFM() error {

	if p.parsedFM != nil {
		return nil
	}

	parsedFM := make(map[interface{}]interface{})
	err := goyaml.Unmarshal(p.frontmatter, &parsedFM)
	if err != nil {
		return err
	}
	p.parsedFM = parsedFM
	return nil
}
