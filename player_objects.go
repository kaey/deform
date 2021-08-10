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

func (o *PObject) addProp(v *PVar) *PVar {
	// TODO: check dups
	o.Props = append(o.Props, v)
	return v
}

func (o *PObject) setProp(prop string, set bool) error {
	p := o.getProp(prop)
	if p == nil {
		return fmt.Errorf("%v has no prop %q", o.Name, prop)
	}

	switch p.Kind {
	case PKindBool:
		p.Val = set
	case PKindEnum:
		p.Val = p.enumVal(prop)
	default:
		return fmt.Errorf("can't set prop %q for %v to val %q, expected prop kind enum or bool", p.Name, o.Name, prop)
	}

	return nil
}

func (o *PObject) getProp(v string) *PVar {
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

	p := o.Kind.getProp(v)
	if p != nil {
		return o.addProp(p.copy())
	}

	return nil
}
