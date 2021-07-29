package main

import (
	"testing"
)

func TestSimple(t *testing.T) {
	for _, f := range files {
		_, err := ParseFile(f)
		if err != nil {
			t.Fatal(err)
		}
	}
}
