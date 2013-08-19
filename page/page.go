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
	content     Content
	parsedFM    map[interface{}]interface{}
}

type FrontMatter []byte
type Content []byte

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

	firstLine, err := peekLine(reader)
	if err != nil {
		return
	}

	page = new(Page)
	page.render = shouldRender(firstLine)

	if page.render && isFrontMatterDelim(firstLine) {
		fm, err := extractFrontMatter(reader)
		if err != nil {
			return nil, err
		}
		page.frontmatter = fm
	}

	content, err := extractContent(reader)
	if err != nil {
		return nil, err
	}

	page.content = content

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

func peekLine(r *bufio.Reader) (line []byte, err error) {
	line, err = r.Peek(5)
	if err != nil {
		return
	}

	if line[3] == '\n' {
		return line[:4], nil
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
func extractFrontMatter(r *bufio.Reader) (fm FrontMatter, err error) {

	// strip off front matter delim if it's present.
	firstLine, err := peekLine(r)
	if err != nil {
		return
	}
	if isFrontMatterDelim(firstLine) {
		if _, err = r.Read(firstLine); err != nil {
			return
		}
	}

	wr := new(bytes.Buffer)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		if isFrontMatterDelim(line) {
			return FrontMatter(wr.Bytes()), nil
		}
		_, err = wr.Write(line)
		if err != nil {
			return nil, err
		}
	}
	return nil, errors.New("Could not find Front Matter.")
}

func extractContent(r io.Reader) (content Content, err error) {
	wr := new(bytes.Buffer)
	if _, err = io.Copy(wr, r); err != nil {
		return
	}
	return wr.Bytes(), nil
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
