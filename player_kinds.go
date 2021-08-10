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

var (
	PKindRule     = &PKind{Name: "rule", Plural: "rules"}
	PKindRulebook = &PKind{Name: "rulebook", Plural: "rulebooks", Props: []*PVar{
		{Name: "default", Val: 2, Kind: PKindEnum, EnumVals: []string{"success", "failure", "no outcome"}},
	}}
	PKindText   = &PKind{Name: "text", Plural: "texts", DefVal: ""}
	PKindNumber = &PKind{Name: "number", Plural: "numbers", DefVal: 0}
	PKindBool   = &PKind{Name: "truth state", DefVal: false}
	PKindTime   = &PKind{Name: "time"}
	PKindObject = &PKind{Name: "object", Plural: "objects"}
	PKindEnum   = &PKind{Name: "enum"}
	PKindValue  = &PKind{Name: "value"}
)

func (terp *Interp) addStandardKinds() {
	terp.addKind(&PKind{Name: "direction", Plural: "directions", Parent: PKindObject})
	terp.addKind(&PKind{Name: "room", Plural: "rooms", Parent: PKindObject})
	terp.addKind(&PKind{Name: "region", Plural: "regions", Parent: PKindObject})

	thing := terp.addKind(&PKind{Name: "thing", Plural: "things", Parent: PKindObject})
	terp.addKind(&PKind{Name: "door", Plural: "doors", Parent: thing})
	terp.addKind(&PKind{Name: "container", Plural: "containers", Parent: thing})
	terp.addKind(&PKind{Name: "player's holdall", Parent: terp.getKind("container")})
	/*terp.addKind(&PKind{Name: "supporter", Plural: "supporters", Parent: thing})*/
	terp.addKind(&PKind{Name: "scenery", Parent: thing})
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
		k.Plural = k.Name + "s"
	}
	terp.Kinds = append(terp.Kinds, k)
	return k
}

func (terp *Interp) getKind(v string) *PKind {
	switch strings.ToLower(v) {
	case PKindRulebook.Name, PKindRulebook.Plural:
		return PKindRulebook
	case PKindRule.Name, PKindRule.Plural:
		return PKindRule
	case PKindText.Name, PKindText.Plural:
		return PKindText
	case PKindNumber.Name, PKindNumber.Plural:
		return PKindNumber
	case PKindBool.Name, "bool":
		return PKindBool
	case PKindTime.Name:
		return PKindTime
	case PKindObject.Name, PKindObject.Plural:
		return PKindObject
	case PKindEnum.Name:
		return PKindEnum
	case PKindValue.Name:
		return PKindValue
	}

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
			return k, parts[:i]
		}
	}

	return nil, nil
}

func (kind *PKind) addProp(v *PVar) *PVar {
	// TODO: check dups
	kind.Props = append(kind.Props, v)
	return v
}

func (kind *PKind) getProp(v string) *PVar {
	k := kind
	for {
		for i := range k.Props {
			if strings.EqualFold(v, k.Props[i].Name) {
				if k != kind {
					return kind.addProp(k.Props[i].copy())
				}
				return k.Props[i]
			}
			for _, e := range k.Props[i].EnumVals {
				if strings.EqualFold(v, e) {
					if k != kind {
						return kind.addProp(k.Props[i].copy())
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
