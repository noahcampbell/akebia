package page

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	CONTENT_EMPTY                   = ""
	CONTENT_NO_FRONTMATTER          = "a page with no front matter"
	CONTENT_WITH_FRONTMATTER        = "---\ntitle: front matter\n---\nContent with front matter"
	CONTENT_HTML_NODOCTYPE          = "<html><body></body></html>"
	CONTENT_HTML_WITHDOCTYPE        = "<!doctype html><html><body></body></html>"
	CONTENT_HTML_WITH_FRONTMATTER   = "---\ntilte: front matter\n---\n<!doctype><html><body></body></html>"
	CONTENT_LWS_HTML                = "    <html><body></body></html>"
	CONTENT_LWS_LF_HTML             = "\n<html><body></body></html>"
	CONTENT_INCOMPLETE_BEG_FM_DELIM = "--\ntitle: incomplete beg fm delim\n---\nincomplete frontmatter delim"
	CONTENT_INCOMPLETE_END_FM_DELIM = "---\ntitle: incomplete end fm delim\n--\nincomplete frontmatter delim"
	CONTENT_MISSING_END_FM_DELIM    = "---\ntitle: incomplete end fm delim\nincomplete frontmatter delim"
	CONTENT_FM_NO_DOC               = "---\ntitle: no doc\n---"
)

var lineEndings = []string{"\n", "\r\n"}

func pageMust(page *Page, err error) *Page {
	if err != nil {
		panic(err)
	}
	return page
}

func pageRecoverAndLog(t *testing.T) {
	if err := recover(); err != nil {
		t.Errorf("panic/recover: %s\n", err)
	}
}

func TestDegenerateCreatePageFrom(t *testing.T) {
	tests := []struct {
		content string
	}{
		{CONTENT_EMPTY},
		{CONTENT_MISSING_END_FM_DELIM},
		{CONTENT_INCOMPLETE_END_FM_DELIM},
		{CONTENT_FM_NO_DOC},
		//{CONTENT_INCOMPLETE_BEG_FM_DELIM},
	}

	for _, test := range tests {
		for _, ending := range lineEndings {
			test.content = strings.Replace(test.content, "\n", ending, -1)
			_, err := ReadFrom(strings.NewReader(test.content))
			if err == nil {
				t.Errorf("Content should return an err:\n%s\n", test.content)
			}
		}
	}
}

func checkPageRender(t *testing.T, p *Page, expected bool) {
	if p.render != expected {
		t.Errorf("page.render should be %t, got: %t", expected, p.render)
	}
}

func checkPageFrontMatterIsNil(t *testing.T, p *Page, content string, expected bool) {
	if bool(p.frontmatter == nil) != expected {
		t.Logf("\n%q\n", content)
		t.Errorf("page.frontmatter == nil? %t, got %t", expected, p.frontmatter == nil)
	}
}

func TestStandaloneCreatePageFrom(t *testing.T) {
	tests := []struct {
		content            string
		expectedMustRender bool
		frontMatterIsNil   bool
	}{
		{CONTENT_NO_FRONTMATTER, true, true},
		{CONTENT_WITH_FRONTMATTER, true, false},
		{CONTENT_HTML_NODOCTYPE, false, true},
		{CONTENT_HTML_WITHDOCTYPE, false, true},
		{CONTENT_HTML_WITH_FRONTMATTER, true, false},
		{CONTENT_LWS_HTML, false, true},
		{CONTENT_LWS_LF_HTML, false, true},
	}

	for _, test := range tests {
		for _, ending := range lineEndings {
			test.content = strings.Replace(test.content, "\n", ending, -1)
			p := pageMust(ReadFrom(strings.NewReader(test.content)))
			checkPageRender(t, p, test.expectedMustRender)
			checkPageFrontMatterIsNil(t, p, test.content, test.frontMatterIsNil)
		}
	}
}

func TestLongFormRender(t *testing.T) {

	tests := []struct {
		filename string
	}{
		{"long_text_test.md"},
	}
	for _, test := range tests {
		path := filepath.FromSlash(test.filename)
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("Unable to open %s: %s", path, err)
		}

		defer pageRecoverAndLog(t)
		p := pageMust(ReadFrom(f))
		checkPageFrontMatterIsNil(t, p, "[long file]", false)
	}
}

func TestPageShouldRender(t *testing.T) {
	tests := []struct {
		content  []byte
		expected bool
	}{
		{[]byte{}, false},
		{[]byte{'<'}, false},
		{[]byte{'-'}, true},
		{[]byte("--"), true},
		{[]byte("---"), true},
		{[]byte("---\n"), true},
		{[]byte{'a'}, true},
	}

	for _, test := range tests {
		for _, ending := range lineEndings {
			test.content = bytes.Replace(test.content, []byte("\n"), []byte(ending), -1)
			if render := shouldRender(test.content); render != test.expected {

				t.Errorf("Expected %s to shouldRender = %t, got: %t", test.content, test.expected, render)
			}
		}
	}
}

func TestPageHasFrontMatter(t *testing.T) {
	tests := []struct {
		content  []byte
		expected bool
	}{
		{[]byte{'-'}, false},
		{[]byte("--"), false},
		{[]byte("---"), false},
		{[]byte("---\n"), true},
		{[]byte("---\nA"), true},
		// TODO support MAC encoding {[]byte("---\rA"), true},
		{[]byte{'a'}, false},
	}
	for _, test := range tests {

		for _, ending := range lineEndings {
			test.content = bytes.Replace(test.content, []byte("\n"), []byte(ending), -1)
			if isFrontMatterDelim := isFrontMatterDelim(test.content); isFrontMatterDelim != test.expected {
				t.Errorf("Expected %q isFrontMatterDelim = %t,  got: %t", test.content, test.expected, isFrontMatterDelim)
			}
		}
	}
}

func TestExtractFrontMatter(t *testing.T) {

	tests := []struct {
		frontmatter string
		extracted   []byte
		errIsNil    bool
	}{
		{"", nil, false},
		{"-", nil, false},
		{"---\n", nil, false},
		{"---\nfoobar", nil, false},
		{"---\nfoobar\nbarfoo\nfizbaz\n", nil, false},
		{"---\nblar\n-\n", nil, false},
		{"ralb\n---\n", []byte("ralb"), true},
		{"minc\n---\ncontent", []byte("minc"), true},
		{"cnim\n---\ncontent\n", []byte("cnim"), true},
		{"---\nralb\n---\n", []byte("ralb"), true},
		{"---\nminc\n---\ncontent", []byte("minc"), true},
		{"---\ncnim\n---\ncontent\n", []byte("cnim"), true},
	}

	for _, test := range tests {
		fm, err := extractFrontMatter(strings.NewReader(test.frontmatter))
		if (err == nil) != test.errIsNil {
			t.Logf("\n%q\n", string(test.frontmatter))
			t.Errorf("Expected err == nil => %t, got: %t. err: %s", test.errIsNil, err == nil, err)
			continue
		}
		if !bytes.Equal(fm, test.extracted) {
			t.Errorf("Expected Front Matter %q. got %q", string(test.extracted), fm)
		}
	}
}
