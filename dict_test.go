package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

var phs = []string{
	"clear the screen",
	"A + B",
	"A + B / C",
	"10 + 5 / 3 ",
	"10 + (15 - 20)",
	`say "some [A] text"`,
}

func TestDict(t *testing.T) {
	d := new(Dict)
	d.AddFuncStub("clear the screen")
	d.AddFunc(&PFunc{Name: "say"}, []FuncPart{{Word: "say"}, {ArgName: "v", ArgKind: "text"}})
	d.Add(&PVar{Name: "A"}, "A")
	d.Add(&PVar{Name: "B"}, "B")
	d.Add(&PVar{Name: "C"}, "C")
	for _, ph := range phs {
		t.Run(ph, func(t *testing.T) {
			r, err := ParsePhrases(lex(ph, ph), d)
			if err != nil {
				t.Error(err)
			}

			spew.Dump(r)
		})
	}
}
