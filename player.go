package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "net/http/pprof"
)

func main() {
	go http.ListenAndServe(":3998", nil)
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Main() error {
	terp := newInterp()
	start := time.Now()
	sentences := make([]Sentence, 0, 100)
	for _, f := range files {
		s, err := ParseFile(f)
		if err != nil {
			return err
		}
		sentences = append(sentences, s...)
	}
	log.Println("Parse done in", time.Since(start))

	start = time.Now()
	errs := make(errs, 0, 10)
	for _, s := range sentences {
		if err := terp.evalSentence(s); err != nil {
			errs = append(errs, err)
			if len(errs) > 9 {
				break
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	log.Println("Eval done in", time.Since(start))

	start = time.Now()
	if err := terp.secondStage(); err != nil {
		return err
	}
	log.Println("Second stage done in", time.Since(start))

	terp.outPrintGo("tq_ng/game")

	return nil
}

type errs []error

func (e errs) Error() string {
	r := new(strings.Builder)
	for _, ee := range e {
		fmt.Fprintf(r, "%v\n", ee)
	}
	return r.String()
}

type Interp struct {
	dict *Dict
	tmp  *interpTmp

	Funcs     []*PFunc
	Kinds     []*PKind
	Vars      []*PVar
	Objects   []*PObject
	RelInRoom []*PRel
	Tables    map[string]PTable // TODO
}

type interpTmp struct {
	funcs []Func
	defs  []Definition
	rules []Rule
}

func newInterp() *Interp {
	terp := &Interp{
		dict:   new(Dict),
		tmp:    new(interpTmp),
		Tables: make(map[string]PTable, 100),
	}

	terp.addStandardKinds()
	terp.addStandardFuncs()
	terp.addStandardRelations()

	return terp
}

func (terp *Interp) evalSentence(s Sentence) error {
	switch v := s.(type) {
	case Subheader:
	case Figure:
	case FileOf:
	case Understand:
	case DoesThePlayerMean:
	case RuleFor:
	case Comment:
	case BeginsHere:
	case EndsHere:
	case Table:
		// TODO
		name := strings.ToLower(v.Name)
		t := terp.Tables[name]
		if t.Kinds == nil {
			t.Pos = v.Pos
			t.Kinds = make(map[string]string)
			for i := range v.ColKinds {
				t.Kinds[v.ColNames[i]] = v.ColKinds[i]
			}
		}
		for _, row := range v.Rows {
			m := make(map[string]interface{}, len(v.ColNames))
			for i, colname := range v.ColNames {
				m[colname] = row[i]
			}
			t.Rows = append(t.Rows, m)
		}
		terp.Tables[name] = t
	case Variable:
		k := terp.getKind(v.Kind)
		if k == nil {
			return fmt.Errorf("%v: kind not found for variable: %v", v.Pos, v.Name)
		}

		terp.addVar(&PVar{
			Pos:   v.Pos,
			Name:  v.Name,
			Array: v.Array,
			Kind:  k,
		})
	case Is:
		if v.Direction != "" {
			// TODO
			return nil
		}

		// TODO: handle negate and never

		k := terp.getKind(v.Object)
		if k != nil {
			if len(v.EnumVal) > 0 {
				if k.Parent.Name != "value" {
					return fmt.Errorf("%v: enum can only be assigned to a kind of value", v.Pos)
				}
				k.Parent = terp.getKind("enum")
				k.ValidVals = v.EnumVal
				return nil
			}
			if val, ok := v.Value.(string); ok {
				// TODO: add setProp
				p := terp.getKindProp(k, val)
				if p == nil {
					return fmt.Errorf("%v: kind %v has no prop %v", v.Pos, v.Object, val)
				}
				if len(p.EnumVals) > 0 {
					p.Val = p.enumVal(val)
				} else {
					p.Val = true
				}
				return nil
			}

			return fmt.Errorf("%v: kind value not understood", v.Pos)
		}

		va := terp.getVar(v.Object)
		if va != nil {
			// TODO: add kind check and maybe setVal()?
			switch val := v.Value.(type) {
			case int, bool, QuotedString:
				va.Val = v.Value
				return nil
			case string:
				o := terp.getObject(val)
				if o == nil {
					fmt.Fprintf(Log("errors"), "%v: object not defined: %v\n", v.Pos, val)
					return nil
					// return fmt.Errorf("%v: object not defined: %v", v.Pos, val)
				}
				va.Val = o
				return nil
			}
			return fmt.Errorf("%v: var value not understood", v.Pos)
		}

		o := terp.getObject(v.Object)
		if o != nil {
			if val, ok := v.Value.(string); ok {
				if val == "everywhere" {
					// TODO: scenery can be everywhere
					return nil
				}
				if err := terp.setObjectProp(o, val, !(v.Negate || v.Never)); err != nil {
					return fmt.Errorf("%v: %v", v.Pos, err)
				}
			}

			return nil
		}

		if v.Initially || v.Always { // TODO: make "always" a constant
			terp.addVar(&PVar{
				Pos:  v.Pos,
				Name: v.Object,
				Val:  v.Value,
				Kind: terp.getKind("value"), // TODO: maybe use more specific kind based on value.
			})
			return nil
		}

		if v.Usually || v.Always || v.Never || v.Negate {
			fmt.Fprintf(Log("errors"), "%v: %v = %v\n", v.Pos, v.Object, v.Value)
			return nil
			return fmt.Errorf("%v: bug: unparsed value", v.Pos)
		}

		if val, ok := v.Value.(string); ok {
			k, props := terp.getKindWithProps(val)
			if k == nil {
				return fmt.Errorf("%v: kind %q is not defined ", v.Pos, val)
			}

			o := terp.addObject(&PObject{
				Pos:  v.Pos,
				Name: v.Object,
				Kind: k,
			})
			for _, p := range props {
				if err := terp.setObjectProp(o, p, true); err != nil {
					return fmt.Errorf("%v: %v", v.Pos, err)
				}
			}

			return nil
		}

		return fmt.Errorf("does not exist: %v = %v", v.Object, v.Value)
	case IsIn:
		where := terp.getObject(v.Where)
		if where == nil {
			return fmt.Errorf("%v: room %v is not defined", v.Pos, v.Where)
		}

		for _, obj := range v.Objects {
			var o *PObject
			if v.Decl {
				k, props := terp.getKindWithProps(obj)
				if k == nil {
					return fmt.Errorf("%v: kind %v is not defined", v.Pos, obj)
				}
				// Don't use addObjects because dups here are allowed.
				o := &PObject{
					Pos:  v.Pos,
					Name: k.Name,
					Kind: k,
				}
				terp.Objects = append(terp.Objects, o)
				for _, p := range props {
					if err := terp.setObjectProp(o, p, true); err != nil {
						return fmt.Errorf("%v: %v", v.Pos, err)
					}
				}
			} else {
				o = terp.getObject(obj)
				if o == nil {
					return fmt.Errorf("%v: object %v is not defined", v.Pos, obj)
				}
			}
			terp.RelInRoom = append(terp.RelInRoom, &PRel{
				Left:  o,
				Right: where,
			})
		}
	case Func:
		terp.addFunc(v)
	case Definition:
		terp.tmp.defs = append(terp.tmp.defs, v)
	case ListedInRulebook:
		// TODO: priority
		// panic(fmt.Sprintf("listed: %v", v))
	case Rule:
		terp.tmp.rules = append(terp.tmp.rules, v)
	case Action:
		s := strings.Split(v.Name, " ")
		for i := range s {
			if s[i] == "it" {
				s[i] = "@"
			}
		}
		if v.NThings > 1 {
			s = append(s, "@")
		}

		terp.dict.Add2(v.Name, s...)
		terp.dict.Add2(v.Name, append(s, "instead")...)
	case Prop:
		if k := terp.getKind(v.Object); k != nil {
			pk := terp.getKind(v.Kind)
			if pk != nil {
				if pk.Parent != nil && pk.Parent.Name == "enum" {
					name := v.Name
					if name == "" {
						name = pk.Name
					}
					terp.addKindProp(k, &PVar{
						Pos:      v.Pos,
						Name:     name,
						Kind:     pk.Parent,
						Val:      pk.DefVal,
						EnumVals: pk.ValidVals,
					})
				} else {
					terp.addKindProp(k, &PVar{
						Pos:  v.Pos,
						Name: v.Name,
						Kind: pk,
						Val:  pk.DefVal,
					})
				}

				return nil
			}

			// has [prop] [propval]
			parts := strings.Split(v.Kind, " ")
			if len(parts) != 2 {
				return fmt.Errorf("%v: prop not understood: %v", v.Pos, v.Kind)
			}
			p := terp.getKindProp(k, parts[0])
			if p == nil {
				return fmt.Errorf("%v: %v doesn't have prop: %v", v.Pos, v.Object, parts[0])
			}
			p.Val = parts[1] // TODO: parse val?
			return nil
		}

		if o := terp.getObject(v.Object); o != nil {
			pk := terp.getKind(v.Kind)
			if pk != nil {
				terp.addObjectProp(o, &PVar{
					Pos:  v.Pos,
					Name: v.Name,
					Kind: pk,
					Val:  pk.DefVal,
				})
				return nil
			}

			// has [prop] [propval]
			parts := strings.Split(v.Kind, " ")
			if len(parts) != 2 {
				return fmt.Errorf("%v: prop not understood: %v", v.Pos, v.Kind)
			}
			p := terp.getObjectProp(o, parts[0])
			if p == nil {
				return fmt.Errorf("%v: %v doesn't have prop: %v", v.Pos, v.Object, parts[0])
			}
			p.Val = parts[1] // TODO: parse val? also consider p.setVal
			return nil
		}

		return fmt.Errorf("%v: prop for undefined object: %v", v.Pos, v.Object)
	case PropVal:
		k := terp.getKind(v.Object)
		if k != nil {
			p := terp.getKindProp(k, v.Prop)
			if p == nil {
				return fmt.Errorf("%v: kind %v has no prop %q", v.Pos, v.Object, v.Prop)
			}
			p.Val = v.Val // TODO: verify kind
			return nil
		}

		o := terp.getObject(v.Object)
		if o != nil {
			p := terp.getObjectProp(o, v.Prop)
			if p == nil {
				return fmt.Errorf("%v: object %v has no prop %q", v.Pos, v.Object, v.Prop)
			}
			p.Val = v.Val // TODO: verify kind
			return nil
		}

		return fmt.Errorf("%v: prop for undefined object or kind: %v", v.Pos, v.Object)
	case PropEnum:
		// TODO: len vals == 1 should be bool
		k := terp.getKind(v.Object)
		if k != nil {
			name := v.Name
			if name == "" {
				name = fmt.Sprintf("%v-enum-%v", k.Name, strings.Join(v.Vals, ","))
			}
			terp.addKindProp(k, &PVar{
				Pos:      v.Pos,
				Name:     name,
				Val:      len(v.Vals) - 1,
				Kind:     terp.getKind("enum"),
				EnumVals: v.Vals,
			})
			return nil
		}

		o := terp.getObject(v.Object)
		if o != nil {
			name := v.Name
			if name == "" {
				name = fmt.Sprintf("%v-enum-%v", o.Name, strings.Join(v.Vals, ","))
			}
			terp.addObjectProp(o, &PVar{
				Pos:      v.Pos,
				Name:     name,
				Val:      len(v.Vals) - 1,
				Kind:     terp.getKind("enum"),
				EnumVals: v.Vals,
			})
			return nil
		}

		return fmt.Errorf("%v: %s is not defined", v.Pos, v.Object)

	case Kind:
		par := terp.getKind(v.Kind)
		if par == nil {
			return fmt.Errorf("%v: parent kind undefined: %v", v.Pos, v.Kind)
		}

		terp.addKind(&PKind{
			Pos:    v.Pos,
			Name:   v.Name,
			Parent: par,
		})
	case ThereAre:
		k, props := terp.getKindWithProps(v.Kind)
		if k == nil {
			return fmt.Errorf("%v: kind %q not defined", v.Pos, v.Kind)
		}
		var w *PObject
		if v.Where != "" {
			w = terp.getObject(v.Where)
			if w == nil {
				return fmt.Errorf("%v: object %q not defined", v.Pos, v.Where)
			}
		}

		for i := 0; i < v.N; i++ {
			// Don't use addObjects because dups here are allowed.
			o := &PObject{
				Pos:  v.Pos,
				Name: k.Name,
				Kind: k,
			}
			for _, p := range props {
				if err := terp.setObjectProp(o, p, true); err != nil {
					return fmt.Errorf("%v: %v", v.Pos, err)
				}
			}
			terp.Objects = append(terp.Objects, o)
			if w != nil {
				terp.RelInRoom = append(terp.RelInRoom, &PRel{
					Left:  o,
					Right: w,
				})
			}
		}
	case QuotedString:
		// TODO
	case Vector:
		// TODO
	case Relation:
		// TODO
	case Verb:
		terp.dict.AddBinary(v.Rel, v.Name)
		for _, a := range v.Alts {
			terp.dict.AddBinary(v.Rel, a)
		}
	case Activity:
		// TODO
	default:
		return fmt.Errorf("unknown sentence %T", v)
	}

	return nil
}

func (terp *Interp) secondStage() error {
	for _, d := range terp.tmp.defs {
		k := terp.getKind(d.Object)
		if k == nil {
			o := terp.getObject(d.Object)
			if o == nil {
				return fmt.Errorf("%v: kind or object %v not defined", d.Pos, d.Object)
			}
		}
		terp.dict.Add(d.Prop, d.Prop) // TODO: proper target
		if d.NegProp != "" {
			terp.dict.Add("not "+d.Prop, d.NegProp) // TODO: proper target
		}
	}

	terp.dict.Sort()

	for _, pat := range terp.dict.ident {
		fmt.Fprintf(Log("dict-ident"), "%v\n", pat.parts)
	}

	for _, pat := range terp.dict.binary {
		fmt.Fprintf(Log("dict-binary"), "%v\n", pat.parts)
	}

	/*{
		wg := new(sync.WaitGroup)
		li := make(chan struct{}, 6)
		for _, r := range terp.tmp.rules {
			if r.RawCond == nil {
				continue
			}
			r := r
			wg.Add(1)
			li <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-li
				}()

				_, err := ParseExpr(r.RawCond, terp.dict)
				log.Println("Rule cond:", r.Pos, r.Rulebook)
				if err != nil {
					log.Println(err)
					return
				}
			}()
		}
		wg.Wait()
	}*/

	/*{
		wg := new(sync.WaitGroup)
		li := make(chan struct{}, 6)
		for _, r := range terp.tmp.rules {
			r := r
			wg.Add(1)
			li <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-li
				}()

				dict := terp.dict.Clone()
				_, err := ParsePhrases(r.RawPhrases, dict)
				log.Println("Rule:", r.Pos, r.Name, r.Rulebook)
				if err != nil {
					log.Println(err)
					return
				}
			}()
		}
		wg.Wait()
	}*/

	{
		wg := new(sync.WaitGroup)
		li := make(chan struct{}, 6)
		for _, d := range terp.tmp.defs {
			d := d
			wg.Add(1)
			li <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-li
				}()
				dict := terp.dict.Clone()
				if d.Called != "" {
					dict.Add("local var", d.Called)
					dict.Sort()
				}
				_, err := ParsePhrases(d.RawPhrases, dict)
				log.Println("Definition:", d.Pos, d.Object, d.Prop)
				if err != nil {
					log.Println(err)
					return
				}
			}()
		}
		wg.Wait()
	}

	/*{
		wg := new(sync.WaitGroup)
		li := make(chan struct{}, 6)
		for _, f := range terp.tmp.funcs {
			f := f
			wg.Add(1)
			li <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-li
				}()
				dict := terp.dict.Clone()
				for _, p := range f.Parts {
					if p.Word != "" {
						continue
					}
					dict.Add("func arg", p.ArgName)
				}
				_, err := ParsePhrases(f.RawPhrases, dict)
				log.Println("Func:", f.Pos, funcName(f.Parts))
				if err != nil {
					log.Println(err)
					return
				}
			}()
		}
		wg.Wait()
	}*/

	return nil
}
