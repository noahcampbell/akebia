package page

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	CONTENT_EMPTY                   = ""
	CONTENT_NO_FRONTMATTER          = "a page with no front matter"
	CONTENT_WITH_FRONTMATTER        = "---\ntitle: front matter\n---\nContent with front matter"
	CONTENT_HTML_NODOCTYPE          = "<html>\n\t<body>\n\t</body>\n</html>"
	CONTENT_HTML_WITHDOCTYPE        = "<!doctype html><html><body></body></html>"
	CONTENT_HTML_WITH_FRONTMATTER   = "---\ntitle: front matter\n---\n<!doctype><html><body></body></html>"
	CONTENT_LWS_HTML                = "    <html><body></body></html>"
	CONTENT_LWS_LF_HTML             = "\n<html><body></body></html>"
	CONTENT_INCOMPLETE_BEG_FM_DELIM = "--\ntitle: incomplete beg fm delim\n---\nincomplete frontmatter delim"
	CONTENT_INCOMPLETE_END_FM_DELIM = "---\ntitle: incomplete end fm delim\n--\nincomplete frontmatter delim"
	CONTENT_MISSING_END_FM_DELIM    = "---\ntitle: incomplete end fm delim\nincomplete frontmatter delim"
	CONTENT_FM_NO_DOC               = "---\ntitle: no doc\n---"
)

var lineEndings = []string{"\n", "\r\n"}
var delimiters = []string{"-", "+"}

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

func checkPageFrontMatterContent(t *testing.T, p *Page, frontMatter string) {
	if p.frontmatter == nil {
		return
	}
	if !bytes.Equal(p.frontmatter, []byte(frontMatter)) {
		t.Errorf("expected frontmatter %q, got %q", frontMatter, p.frontmatter)
	}
}

func checkPageContent(t *testing.T, p *Page, expected string) {
	if !bytes.Equal(p.content, []byte(expected)) {
		t.Errorf("expected content %q, got %q", expected, p.content)
	}
}

func TestStandaloneCreatePageFrom(t *testing.T) {
	tests := []struct {
		content            string
		expectedMustRender bool
		frontMatterIsNil   bool
		frontMatter        string
		bodycontent        string
	}{
		{CONTENT_NO_FRONTMATTER, true, true, "", "a page with no front matter"},
		{CONTENT_WITH_FRONTMATTER, true, false, "title: front matter\n", "Content with front matter"},
		{CONTENT_HTML_NODOCTYPE, false, true, "", "<html>\n\t<body>\n\t</body>\n</html>"},
		{CONTENT_HTML_WITHDOCTYPE, false, true, "", "<!doctype html><html><body></body></html>"},
		{CONTENT_HTML_WITH_FRONTMATTER, true, false, "title: front matter\n", "<!doctype><html><body></body></html>"},
		{CONTENT_LWS_HTML, false, true, "", "<html><body></body></html>"},
		{CONTENT_LWS_LF_HTML, false, true, "", "<html><body></body></html>"},
	}

	for _, test := range tests {
		for _, ending := range lineEndings {
			test.content = strings.Replace(test.content, "\n", ending, -1)
			test.frontMatter = strings.Replace(test.frontMatter, "\n", ending, -1)
			test.bodycontent = strings.Replace(test.bodycontent, "\n", ending, -1)
			p := pageMust(ReadFrom(strings.NewReader(test.content)))
			checkPageRender(t, p, test.expectedMustRender)
			checkPageFrontMatterIsNil(t, p, test.content, test.frontMatterIsNil)
			checkPageFrontMatterContent(t, p, test.frontMatter)
			checkPageContent(t, p, test.bodycontent)
		}
	}
}

func BenchmarkLongFormRender(b *testing.B) {

	tests := []struct {
		filename string
		buf      []byte
	}{
		{filename: "long_text_test.md"},
	}
	for i, test := range tests {
		path := filepath.FromSlash(test.filename)
		f, err := os.Open(path)
		if err != nil {
			b.Fatalf("Unable to open %s: %s", path, err)
		}
		defer f.Close()
		membuf := new(bytes.Buffer)
		if _, err := io.Copy(membuf, f); err != nil {
			b.Fatalf("Unable to read %s: %s", path, err)
		}
		tests[i].buf = membuf.Bytes()
	}

	b.ResetTimer()

	for i := 0; i <= b.N; i++ {
		for _, test := range tests {
			ReadFrom(bytes.NewReader(test.buf))
		}
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
		{"ralb\nbeeboo\n---\n", []byte("ralb\nbeeboo\n"), true},
		{"minc\n---\ncontent", []byte("minc\n"), true},
		{"cnim\n---\ncontent\n", []byte("cnim\n"), true},
		{"---\nralb\n---\n", []byte("ralb\n"), true},
		{"---\nminc\n---\ncontent", []byte("minc\n"), true},
		{"---\ncnim\n---\ncontent\n", []byte("cnim\n"), true},
	}

	for _, test := range tests {
		for _, ending := range lineEndings {
			test.frontmatter = strings.Replace(test.frontmatter, "\n", ending, -1)
			test.extracted = bytes.Replace(test.extracted, []byte("\n"), []byte(ending), -1)
			for _, delim := range delimiters {
				test.frontmatter = strings.Replace(test.frontmatter, "-", delim, -1)
				fm, err := extractFrontMatter(bufio.NewReader(strings.NewReader(test.frontmatter)))
				if (err == nil) != test.errIsNil {
					t.Logf("\n%q\n", string(test.frontmatter))
					t.Errorf("Expected err == nil => %t, got: %t. err: %s", test.errIsNil, err == nil, err)
					continue
				}
				if !bytes.Equal(fm, test.extracted) {
					t.Logf("\n%q\n", string(test.frontmatter))
					t.Errorf("Expected front matter %q. got %q", string(test.extracted), fm)
				}
			}
		}
	}
}

func TestParseFrontMatter(t *testing.T) {
	type expectedProp []struct {
		key, value string
		ok         bool
	}
	tests := []struct {
		frontmatter string
		properties  expectedProp
	}{
		{"---\ntitle: test\nauthor: me\n---\n",
			expectedProp{
				{"title", "test", true},
				{"author", "me", true},
				{"missing", "", false},
			},
		},
	}

	for _, test := range tests {
		p := pageMust(ReadFrom(strings.NewReader(test.frontmatter)))

		for _, expected := range test.properties {
			v, ok := p.Property(expected.key)
			if v != expected.value {
				t.Errorf("Expected key: %s to be %s, got: %s", expected.key, expected.value, v)
				continue
			}
			if ok != expected.ok {
				t.Errorf("Expected key: %s ok: %t, got %t", expected.key, expected.ok, ok)
			}
		}
	}
}

func TestDegeneratePageParseBadFM(t *testing.T) {
	var badFrontMatter = "---\nfoobar: [fiz}bazz</html>\n---\ncontent\n"
	_, err := ReadFrom(strings.NewReader(badFrontMatter))
	if err == nil {
		t.Fatalf("Expected ReadFrom to return an error.")
	}
}
