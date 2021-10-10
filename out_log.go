package main

import (
	"io"
	"os"
	"path/filepath"
)

var logfiles = make(map[string]*os.File, 10)

func Log(v string) io.Writer {
	f, ok := logfiles[v]
	if !ok {
		ff, err := os.Create(filepath.Join("tq_ng", "log", v))
		if err != nil {
			panic(err)
		}
		logfiles[v] = ff
		return ff
	}

	return f
}
