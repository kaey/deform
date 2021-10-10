package main

import (
	"fmt"
	"strings"
)

type PKind struct {
	Pos       Pos
	Name      string
	Plural    string
	Parent    *PKind
	DefVal    interface{}
	Props     []*PVar
	ValidVals []string // Enum kind only
}

func (terp *Interp) addStandardKinds() {
	terp.addKind(&PKind{Name: "enum"})
	terp.addKind(&PKind{Name: "rule", Plural: "rules"})
	terp.addKind(&PKind{Name: "rulebook", Plural: "rulebooks", Props: []*PVar{
		{Name: "default", Val: 2, Kind: terp.getKind("enum"), EnumVals: []string{"success", "failure", "no outcome"}},
	}})
	terp.addKind(&PKind{Name: "text", Plural: "texts", DefVal: ""})
	terp.addKind(&PKind{Name: "number", Plural: "numbers", DefVal: 0})
	terp.addKind(&PKind{Name: "real number", Plural: "real numbers", DefVal: float32(0)})
	terp.addKind(&PKind{Name: "truth state", DefVal: false})
	terp.addKind(&PKind{Name: "time"})
	terp.addKind(&PKind{Name: "object", Plural: "objects"})
	terp.addKind(&PKind{Name: "value"})
	terp.addKind(&PKind{Name: "direction", Plural: "directions", Parent: terp.getKind("object")})
	terp.addKind(&PKind{Name: "room", Plural: "rooms", Parent: terp.getKind("object")})
	terp.addKind(&PKind{Name: "region", Plural: "regions", Parent: terp.getKind("object")})
	terp.addKind(&PKind{Name: "action", Plural: "actions", Parent: terp.getKind("object")})

	thing := terp.addKind(&PKind{Name: "thing", Plural: "things", Parent: terp.getKind("object"), Props: []*PVar{
		{Name: "text-shortcut", Val: "", Kind: terp.getKind("text")},
		{Name: "clothingFocusPriority", Val: 10, Kind: terp.getKind("number")},
	}})
	terp.addKind(&PKind{Name: "door", Plural: "doors", Parent: thing})
	terp.addKind(&PKind{Name: "container", Plural: "containers", Parent: thing})
	terp.addKind(&PKind{Name: "player's holdall", Parent: terp.getKind("container")})
	terp.addKind(&PKind{Name: "backdrop", Parent: thing})
	terp.addKind(&PKind{Name: "device", Plural: "devices", Parent: thing})
	terp.addKind(&PKind{Name: "focus-thing", Plural: "focus-things", Parent: thing}) // TODO: remove this
	terp.addKind(&PKind{Name: "figure-name", Parent: thing})

	person := terp.addKind(&PKind{Name: "person", Plural: "people", Parent: thing})
	terp.addKind(&PKind{Name: "man", Plural: "men", Parent: person})
	terp.addKind(&PKind{Name: "woman", Plural: "women", Parent: person})
	terp.addKind(&PKind{Name: "animal", Plural: "animals", Parent: person})
}

func (terp *Interp) addKind(k *PKind) *PKind {
	for _, ki := range terp.Kinds {
		if strings.EqualFold(k.Name, ki.Name) || strings.EqualFold(k.Name, ki.Plural) {
			panic(fmt.Sprintf("%v: kind %v already defined at %v", k.Pos, k.Name, ki.Pos))
		}
	}
	if k.Plural == "" {
		switch k.Name {
		case "bag lunch":
			k.Plural = "bag lunches"
		case "wrench":
			k.Plural = "wrenches"
		default:
			k.Plural = k.Name + "s"
		}
	}
	terp.Kinds = append(terp.Kinds, k)
	terp.dict.Add(k, k.Name)
	terp.dict.Add(k, k.Plural)
	return k
}

func (terp *Interp) getKind(v string) *PKind {
	for _, k := range terp.Kinds {
		if strings.EqualFold(v, k.Name) || strings.EqualFold(v, k.Plural) {
			return k
		}
	}
	return nil
}

func (terp *Interp) getKindWithProps(val string) (*PKind, []string) {
	k := terp.getKind(val)
	if k != nil {
		return k, nil
	}
	parts := strings.Split(val, " ")
	for i := 1; i < len(parts); i++ {
		k = terp.getKind(strings.Join(parts[i:], " "))
		if k != nil {
			return k, reduceProps(parts[:i])
		}
	}

	return nil, nil
}

func reduceProps(parts []string) []string {
	if len(parts) == 0 {
		return nil
	}
	props := make([]string, 0, len(parts))
	i := 0
loop:
	for {
		if i+1 >= len(parts) {
			props = append(props, parts[i:]...)
			return props
		}
		for j := range mwordprops {
			if parts[i] == mwordprops[j][0] && parts[i+1] == mwordprops[j][1] {
				props = append(props, strings.Join(parts[i:i+2], " "))
				i += 2
				continue loop
			}
		}
		props = append(props, parts[i])
		i++
	}
}

func (terp *Interp) addKindProp(kind *PKind, v *PVar) *PVar {
	for _, p := range kind.Props {
		if strings.EqualFold(p.Name, v.Name) {
			panic(fmt.Sprintf("%v: prop %v already defined at %v", v.Pos, v.Name, p.Pos))
		}
	}
	if len(v.EnumVals) > 0 {
		fmt.Fprintf(Log("props"), "%v: %v\n", kind.Name, v.EnumVals)
		for _, e := range v.EnumVals {
			terp.dict.Add(e, e)
		}
	} else {
		terp.dict.Add(v, v.Name+" of @")
	}

	kind.Props = append(kind.Props, v)
	return v
}

func (terp *Interp) getKindProp(kind *PKind, v string) *PVar {
	k := kind
	for {
		for i := range k.Props {
			if strings.EqualFold(v, k.Props[i].Name) {
				if k != kind {
					return terp.addKindProp(kind, k.Props[i].copy())
				}
				return k.Props[i]
			}
			for _, e := range k.Props[i].EnumVals {
				if strings.EqualFold(v, e) {
					if k != kind {
						return terp.addKindProp(kind, k.Props[i].copy())
					}
					return k.Props[i]
				}
			}
		}
		if k.Parent == nil {
			return nil
		}
		k = k.Parent
	}
}
