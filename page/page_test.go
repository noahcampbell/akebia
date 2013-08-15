package page

import (
	"testing"
	"strings"
)

var (
	CONTENT_EMPTY = ""
	CONTENT_NO_FRONTMATTER = "a page with no front matter"
	CONTENT_WITH_FRONTMATTER = `---
	front matter
	---
	Content with front matter`
	CONTENT_HTML_NODOCTYPE = `<html><body></body></html>`
	CONTENT_HTML_WITHDOCTYPE = `<!doctype html><html><body></body></html>`
	CONTENT_HTML_WITH_FRONTMATTER = `---
	front matter
	---
	<!doctype><html><body></body></html>`
)

func pageMust(page *Page, err error) *Page {
	if err != nil {
		panic(err)
	}
	return page
}

func TestCreatePageFrom(t *testing.T) {
	_, err := ReadFrom(strings.NewReader(CONTENT_EMPTY))
	if err == nil {
		t.Fatalf("Should not be able to read empty page.")
	}
}

func TestStandaloneCreatePageFrom(t *testing.T) {
	tests := []struct {
		content string
		expectedMustRender bool
	}{
		{CONTENT_NO_FRONTMATTER, true},
		{CONTENT_WITH_FRONTMATTER, true},
		{CONTENT_HTML_NODOCTYPE, false},
		{CONTENT_HTML_WITHDOCTYPE, false},
		{CONTENT_HTML_WITH_FRONTMATTER, true},
	}
	
	for _, test := range tests {
		p := pageMust(ReadFrom(strings.NewReader(test.content)))
		if p.mustRender != test.expectedMustRender {
			t.Fatalf("page.mustRender should be %t, got: %t", test.expectedMustRender, p.mustRender)
		}
	}
}

