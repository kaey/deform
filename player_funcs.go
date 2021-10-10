package main

import (
	"strings"
)

type PFunc struct {
	Pos     Pos
	Name    string
	Phrases []Phrase
}

type PFuncCall struct {
	Func *PFunc
	Args []interface{}
}

func (terp *Interp) addFunc(fn Func) {
	pfn := &PFunc{
		Pos:  fn.Pos,
		Name: funcName(fn.Parts),
	}
	terp.Funcs = append(terp.Funcs, pfn)
	terp.tmp.funcs = append(terp.tmp.funcs, fn)
	terp.dict.AddFunc(pfn, fn.Parts)
}

func funcName(parts []FuncPart) string {
	r := make([]string, len(parts))
	for i := range parts {
		if parts[i].Word != "@" {
			r[i] = parts[i].Word
		} else {
			r[i] = "{" + parts[i].ArgKind + "}"
		}
	}

	return strings.Join(r, " ")
}

func (terp *Interp) addStandardFuncs() {
	terp.dict.AddFuncStub("clear", "the", "screen")
	terp.dict.AddFuncStub("there", "is", "@")
	terp.dict.AddFuncStub("there", "are", "@")
	terp.dict.AddFuncStub("there", "is", "@", "in", "@")
	terp.dict.AddFuncStub("random", "number", "between", "@", "and", "@")
	terp.dict.AddFuncStub("random", "number", "from", "@", "to", "@")
	terp.dict.AddFuncStub("remainder", "after", "dividing", "@", "by", "@")
	terp.dict.AddFuncStub("say", "@")
	terp.dict.AddFuncStub("say", "@", "instead")
	terp.dict.AddFuncStub("now", "@", "is", "@")
	terp.dict.AddFuncStub("todo")
	terp.dict.AddFuncStub("decide", "yes")
	terp.dict.AddFuncStub("decide", "no")
	terp.dict.AddFuncStub("decide", "on", "@")
	terp.dict.AddFuncStub("room", "@", "from", "location")
	terp.dict.AddFuncStub("room", "at", "@")
	terp.dict.AddFuncStub("W", "part", "of", "the", "shape", "of", "@")
	terp.dict.AddFuncStub("E", "part", "of", "the", "shape", "of", "@")
	terp.dict.AddFuncStub("N", "part", "of", "the", "shape", "of", "@")
	terp.dict.AddFuncStub("S", "part", "of", "the", "shape", "of", "@")
	terp.dict.AddFuncStub("number", "of", "entries", "in", "@")
	terp.dict.AddFuncStub("number", "of", "@", "in", "@")
	terp.dict.AddFuncStub("entry", "@", "in", "@")
	terp.dict.AddFuncStub("increase", "@", "by", "@")
	terp.dict.AddFuncStub("decrease", "@", "by", "@")
	terp.dict.AddFuncStub("square", "root", "of", "@")
	terp.dict.AddFuncStub("all", "@")
	terp.dict.AddFuncStub("choice", "in", "row", "@", "of", "@")
}
