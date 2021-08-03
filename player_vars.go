package main

func (terp *Interp) initBuiltinVars() {
	terp.Vars = []PVar{
		0: {Name: "maximum score", Kind: "number", Val: 150},
		1: {Name: "A", Kind: "number"}, // TODO: remove
		2: {Name: "B", Kind: "number"},
	}
	terp.dict.Add(&terp.Vars[0], "maximum score")
	terp.dict.Add(&terp.Vars[1], "A")
	terp.dict.Add(&terp.Vars[2], "B")
}
