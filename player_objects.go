package main

import (
	"fmt"
	"strings"
)

type PObject struct {
	Pos   Pos
	Name  string
	Kind  *PKind
	Props []*PVar
}

func (terp *Interp) addObject(o *PObject) *PObject {
	for _, oi := range terp.Objects {
		if strings.EqualFold(o.Name, oi.Name) {
			panic(fmt.Sprintf("%v: object %v already defined at %v", o.Pos, o.Name, oi.Pos))
		}
	}
	terp.Objects = append(terp.Objects, o)
	terp.dict.Add(o, o.Name)
	return o
}

func (terp *Interp) getObject(v string) *PObject {
	for _, o := range terp.Objects {
		if strings.EqualFold(v, o.Name) {
			return o
		}
	}

	return nil
}

func (terp *Interp) addObjectProp(o *PObject, v *PVar) *PVar {
	for _, p := range o.Props {
		if strings.EqualFold(p.Name, v.Name) {
			panic(fmt.Sprintf("%v: prop %v already defined at %v", v.Pos, v.Name, p.Pos))
		}
	}
	if len(v.EnumVals) > 0 {
		fmt.Fprintf(Log("props"), "%v: %v\n", o.Name, v.EnumVals)
		for _, e := range v.EnumVals {
			terp.dict.Add(e, e)
		}
	} else {
		terp.dict.Add(v, v.Name+" of @")
	}

	o.Props = append(o.Props, v)
	return v
}

func (terp *Interp) setObjectProp(o *PObject, prop string, set bool) error {
	p := terp.getObjectProp(o, prop)
	if p == nil {
		return fmt.Errorf("%v has no prop %q", o.Name, prop)
	}

	switch p.Kind.Name {
	case "truth state":
		p.Val = set
	case "enum":
		p.Val = p.enumVal(prop)
	default:
		return fmt.Errorf("can't set prop %q for %v to val %q, expected prop kind enum or bool", p.Name, o.Name, prop)
	}

	return nil
}

func (terp *Interp) getObjectProp(o *PObject, v string) *PVar {
	for i := range o.Props {
		if strings.EqualFold(v, o.Props[i].Name) {
			return o.Props[i]
		}

		for _, e := range o.Props[i].EnumVals {
			if strings.EqualFold(v, e) {
				return o.Props[i]
			}
		}
	}

	p := terp.getKindProp(o.Kind, v)
	if p != nil {
		return terp.addObjectProp(o, p.copy())
	}

	return nil
}
