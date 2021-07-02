package parse_test

import (
	"testing"

	"github.com/kaey/deform/parse"
)

func TestSimple(t *testing.T) {
	for _, f := range files {
		_, err := parse.ParseFile(f)
		if err != nil {
			t.Fatal(err)
		}
	}
}
