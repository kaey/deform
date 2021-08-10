package main

import (
	"fmt"
	"strings"
)

type PFunc struct {
	Pos      Pos
	Name     string
	NativeFn func([]interface{}) interface{}
	Phrases  []Phrase
}

type PFuncCall struct {
	Func *PFunc
	Args []interface{}
}

func (terp *Interp) addFuncNative(fn func([]interface{}) interface{}, parts []FuncPart) {
	terp.Funcs = append(terp.Funcs, &PFunc{
		Name:     funcName(parts),
		NativeFn: fn,
	})
	terp.dict.AddFunc(terp.Funcs[len(terp.Funcs)-1], parts)
}

func (terp *Interp) addFunc(fn Func) {
	terp.Funcs = append(terp.Funcs, &PFunc{
		Pos:  fn.Pos,
		Name: funcName(fn.Parts),
	})
	terp.tmp.funcs = append(terp.tmp.funcs, fn)
	terp.dict.AddFunc(terp.Funcs[len(terp.Funcs)-1], fn.Parts)
}

func funcName(parts []FuncPart) string {
	r := make([]string, len(parts))
	for i := range parts {
		if parts[i].Word != "" {
			r[i] = parts[i].Word
		} else {
			r[i] = "{" + parts[i].ArgKind + "}"
		}
	}

	return strings.Join(r, " ")
}

func (terp *Interp) addStandardFuncs() {
	terp.dict.AddFuncStub("clear the screen")
	terp.addFuncNative(terp.funcRng, []FuncPart{
		{Word: "a"}, {Word: "random"}, {Word: "number"}, {Word: "between"}, {ArgKind: "number"}, {Word: "and"}, {ArgKind: "number"},
	})
	terp.addFuncNative(terp.funcSay, []FuncPart{
		{Word: "say"}, {ArgKind: "text"},
	})
	terp.addFuncNative(terp.funcNow, []FuncPart{
		{Word: "now"}, {ArgKind: "object"}, {Word: "is"}, {ArgKind: "value"},
	})
}

// a random number between A and B.
func (terp *Interp) funcRng(args []interface{}) interface{} {
	a := terp.EvalExpr(args[0]).(int)
	b := terp.EvalExpr(args[1]).(int)
	terp.rng++
	x := rng(terp.rng, 0x4b3a29d76ab579d1)
	return int(rngReduce(x, uint32(b-a))) + a
}

func (terp *Interp) funcSay(args []interface{}) interface{} {
	v := terp.EvalExpr(args[0])
	fmt.Fprint(terp.tty, v)
	if vv, ok := v.(string); ok {
		if r := vv[len(vv)-1]; r == '.' || r == '!' {
			fmt.Fprint(terp.tty, "\n")
		}
	}
	return nil
}

func (terp *Interp) funcNow(args []interface{}) interface{} {
	v := args[0].(*PVar)
	val := terp.EvalExpr(args[1])
	v.Val = val
	return nil
}
