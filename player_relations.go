package main

type PRel struct {
	Left  *PObject
	Right *PObject
}

func (terp *Interp) addStandardRelations() {
	terp.dict.AddBinary("<", "is less than")
	terp.dict.AddBinary("<", "<")
	terp.dict.AddBinary("<=", "is less or equal than")
	terp.dict.AddBinary("<=", "<=")
	terp.dict.AddBinary(">", "is greater than")
	terp.dict.AddBinary(">", ">")
	terp.dict.AddBinary(">=", "is greater or equal than")
	terp.dict.AddBinary(">=", ">=")
}
