package parse_test

import (
	"testing"

	"github.com/kaey/deform/parse"
	"github.com/kr/pretty"
)

func TestLex(t *testing.T) {
	for _, f := range files {
		pretty.Println(f)
		ss, err := parse.ParseFile(f)
		if err != nil {
			t.Fatal(err)
		}

		for _, s := range ss {
			if s, ok := s.(parse.UnknownSentence); ok {
				pretty.Println("FIXME sentence:", s)
			}
		}
	}
	/*
		s, err := parse.ParseFile("../testdata/aaa.i7x")
		if err != nil {
			t.Fatal(err)
		}

		pretty.Println(s)
	*/
}
