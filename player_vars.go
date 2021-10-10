package main

import (
	"fmt"
	"strings"
)

type PVar struct {
	Pos      Pos
	Name     string
	Val      interface{}
	Kind     *PKind
	Array    bool
	EnumVals []string
}

func (v *PVar) copy() *PVar {
	vv := *v
	return &vv
}

func (terp *Interp) addVar(v *PVar) *PVar {
	if vv := terp.getVar(v.Name); vv != nil {
		panic(fmt.Sprintf("%v: var %v is already defined at %v", v.Pos, v.Name, vv.Pos))
	}

	terp.Vars = append(terp.Vars, v)
	terp.dict.Add(v, v.Name)
	return v
}

func (terp *Interp) getVar(v string) *PVar {
	for i := range terp.Vars {
		if strings.EqualFold(v, terp.Vars[i].Name) {
			return terp.Vars[i]
		}
	}

	return nil
}

func (v *PVar) isKind(kind string) bool {
	k := v.Kind
	for {
		if k.Name == kind {
			return true
		}
		if k.Parent == nil {
			return false
		}
		k = k.Parent
	}
}

func (v *PVar) enumVal(val string) int {
	if len(v.EnumVals) == 0 {
		panic(fmt.Sprintf("enumVal called on non-enum value %v", v))
	}

	for i := range v.EnumVals {
		if strings.EqualFold(val, v.EnumVals[i]) {
			return i
		}
	}

	panic(fmt.Sprintf("enum %v has no value %v", v, val))
}
