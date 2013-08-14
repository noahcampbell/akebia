package page

import (
	"testing"
	"strings"
)

var (
	EMPTY_PAGE = ""
)

func TestCreatePageFrom(t *testing.T) {
	_, err := ReadFrom(strings.NewReader(EMPTY_PAGE))
	if err == nil {
		t.Fatalf("Should not be able to read empty page.")
	}
}
