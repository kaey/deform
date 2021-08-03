package main

import (
	"fmt"
)

func (terp *Interp) initBuiltinFuncs() {
	terp.Funcs = []PFunc{
		0: {Name: "rng", NativeFn: terp.funcRng},
		1: {Name: "say", NativeFn: terp.funcSay},
		2: {Name: "now X is Y", NativeFn: terp.funcNow},
	}

	terp.dict.AddFuncStub("clear the screen")
	terp.dict.AddFunc(&terp.Funcs[0], []FuncPart{
		{Word: "a"},
		{Word: "random"},
		{Word: "number"},
		{Word: "between"},
		{ArgName: "A", ArgKind: "number"},
		{Word: "and"},
		{ArgName: "B", ArgKind: "number"},
	})
	terp.dict.AddFunc(&terp.Funcs[1], []FuncPart{
		{Word: "say"},
		{ArgName: "v", ArgKind: "text"},
	})
	terp.dict.AddFunc(&terp.Funcs[2], []FuncPart{
		{Word: "now"},
		{ArgName: "v", ArgKind: "object"},
		{Word: "is"},
		{ArgName: "val", ArgKind: "value"},
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
	if v.Array {
		panic("TODO: unimplemented")
	}
	v.Val = val
	return nil
}
