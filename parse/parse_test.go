package parse_test

import (
	"testing"

	"github.com/kaey/deform/parse"
)

func TestLex(t *testing.T) {
	for _, f := range files {
		ss, err := parse.ParseFile(f)
		if err != nil {
			t.Fatal(err)
		}
		_ = ss
	}
	/*
		s, err := parse.ParseFile("../testdata/aaa.i7x")
		if err != nil {
			t.Fatal(err)
		}

		pretty.Println(s)
	*/
}
