package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/term"
)

var logfile *os.File

func init() {
	f, err := os.Create("log.txt")
	if err != nil {
		panic(err)
	}
	logfile = f
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Main() error {
	devtty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open tty error error: %w", err)
	}
	defer devtty.Close()

	stty, err := term.MakeRaw(int(devtty.Fd()))
	if err != nil {
		return fmt.Errorf("make raw tty error: %w", err)
	}
	defer fmt.Fprintln(devtty)
	defer term.Restore(int(devtty.Fd()), stty)
	tty := term.NewTerminal(devtty, "> ")

	terp := newInterp(tty)
	for _, f := range files {
		ss, err := ParseFile(f)
		if err != nil {
			return err
		}
		for _, s := range ss {
			terp.evalSentence(s)
		}
	}

	if err := terp.secondStage(); err != nil {
		return err
	}

	if err := terp.mainLoop(); err != nil {
		return err
	}

	return nil
}

type Interp struct {
	tty  *term.Terminal
	dict *Dict
	rng  uint64
	tmp  *interpTmp

	Funcs     []*PFunc
	Kinds     []*PKind
	Vars      []*PVar
	Objects   []*PObject
	RelInRoom []*PRel
	Tables    map[string]PTable  // TODO
	Rulebooks map[string][]PRule // TODO
	Actions   map[string]Action  // TODO
}

type interpTmp struct {
	prev  Sentence
	funcs []Func
	defs  []Definition // TODO
}

func newInterp(tty *term.Terminal) *Interp {
	terp := &Interp{
		tty:       tty,
		dict:      new(Dict),
		tmp:       new(interpTmp),
		Tables:    make(map[string]PTable, 100),
		Rulebooks: make(map[string][]PRule, 100),
	}

	terp.addStandardKinds()
	terp.addStandardFuncs()

	return terp
}

func (terp *Interp) evalSentence(s Sentence) {
	defer func() {
		terp.tmp.prev = s
	}()

	switch v := s.(type) {
	case Subheader:
	case Figure:
	case FileOf:
	case Understand:
	case DoesThePlayerMean:
	case RuleFor:
	case Comment:
	case Table:
		// TODO
		name := strings.ToLower(v.Name)
		t := terp.Tables[name]
		if t.Kinds == nil {
			t.Pos = v.Pos // TODO: remember pos for continued tables.
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
		// TODO: use terp.addVar
		if v2 := terp.getVar(v.Name); v2 != nil {
			panic(fmt.Sprintf("var %s already defined at %v", v.Name, v2.Pos))
		}

		if k := terp.getKind(v.Kind); k != nil {
			terp.Vars = append(terp.Vars, &PVar{
				Pos:   v.Pos,
				Name:  v.Name,
				Array: v.Array,
				Kind:  k,
			})
			terp.dict.Add(&terp.Vars[len(terp.Vars)-1], v.Name)
			return
		}

		panic(fmt.Sprintf("%v: kind not found for variable: %v", v.Pos, v.Name))

	case Is:
		if v.Direction != "" {
			// TODO
			return
		}

		// TODO: handle negate and never

		k := terp.getKind(v.Object)
		if k != nil {
			if len(v.EnumVal) > 0 {
				if k.Parent != PKindValue {
					panic(fmt.Sprintf("%v: enum can only be assigned to a kind of value", v.Pos))
				}
				k.Parent = PKindEnum
				k.ValidVals = v.EnumVal
				return
			}
			if val, ok := v.Value.(string); ok {
				// TODO: add setProp
				p := k.getProp(val)
				if p == nil {
					spew.Fdump(terp.tty, k)
					panic(fmt.Sprintf("%v: kind %v has no prop %v", v.Pos, v.Object, val))
				}
				if len(p.EnumVals) > 0 {
					p.Val = p.enumVal(val)
				} else {
					p.Val = true
				}
				return
			}

			panic(fmt.Sprintf("%v: kind value not understood", v.Pos))
		}

		va := terp.getVar(v.Object)
		if va != nil {
			// TODO: add kind check and maybe setVal()?
			switch val := v.Value.(type) {
			case int, bool, QuotedString:
				va.Val = v.Value
				return
			case string:
				o := terp.getObject(val)
				if o == nil {
					fmt.Fprintf(logfile, "%v: object not defined: %v\n", v.Pos, val)
					return
					panic(fmt.Sprintf("%v: object not defined: %v", v.Pos, val))
				}
				va.Val = o
				return
			}
			panic(fmt.Sprintf("%v: var value not understood", v.Pos))
		}

		o := terp.getObject(v.Object)
		if o != nil {
			if val, ok := v.Value.(string); ok {
				if val == "everywhere" {
					// TODO: scenery can be everywhere
					return
				}
				if err := o.setProp(val, !(v.Negate || v.Never)); err != nil {
					panic(fmt.Sprintf("%v: %v", v.Pos, err))
				}
			}

			return
		}

		if v.Initially || v.Usually || v.Always || v.Never || v.Negate {
			fmt.Fprintf(logfile, "%v: %v = %v\n", v.Pos, v.Object, v.Value)
			return // TODO: create implicit var
			panic(fmt.Sprintf("%v: bug: unparsed value", v.Pos))
		}

		if val, ok := v.Value.(string); ok {
			k, props := terp.getKindWithProps(val)
			if k == nil {
				panic(fmt.Sprintf("%v: kind %q is not defined ", v.Pos, val))
			}

			o := terp.addObject(&PObject{
				Pos:  v.Pos,
				Name: v.Object,
				Kind: k,
			})
			for _, p := range props {
				if err := o.setProp(p, true); err != nil {
					panic(fmt.Sprintf("%v: %v", v.Pos, err))
				}
			}

			return
		}

		// TODO: check enum val
		// TODO
		/*name := strings.ToLower(v.Object)
		/*if _, exists := terp.Vars[name]; exists {
			vv := terp.Vars[name]
			vv.Val = v.Value
			terp.Vars[name] = vv
			return
		}

		if v.Value == "rulebook" {
			if _, exists := terp.Rulebooks[name]; exists {
				panic(fmt.Sprintf("rulebook exists: %v", v.Object))
			}
			// TODO: do we need it?
			return
		}

		// TODO
		// panic(fmt.Sprintf("does not exist: %+#v", v))*/
	case IsIn:
		where := terp.getObject(v.Where)
		if where == nil {
			panic(fmt.Sprintf("%v: room %v is not defined", v.Pos, v.Where))
		}

		for _, obj := range v.Objects {
			var o *PObject
			if v.Decl {
				k, props := terp.getKindWithProps(obj)
				if k == nil {
					panic(fmt.Sprintf("%v: kind %v is not defined", v.Pos, obj))
				}
				// Don't use addObjects because dups here are allowed.
				o := &PObject{
					Pos:  v.Pos,
					Name: k.Name,
					Kind: k,
				}
				terp.Objects = append(terp.Objects, o)
				for _, p := range props {
					if err := o.setProp(p, true); err != nil {
						panic(fmt.Sprintf("%v: %v", v.Pos, err))
					}
				}
			} else {
				o = terp.getObject(obj)
				if o == nil {
					panic(fmt.Sprintf("%v: object %v is not defined", v.Pos, obj))
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
		// TODO
		terp.tmp.defs = append(terp.tmp.defs, v)
	case ListedInRulebook:
		// TODO: priority
		// panic(fmt.Sprintf("listed: %v", v))
	case Rule:
		// TODO
		/*if v.Rulebook == "" {
			// TODO
			return
		}
		rb := strings.ToLower(v.Rulebook)
		rules := terp.tmp.rulebooks[rb]
		rules = append(rules, v)
		terp.tmp.rulebooks[rb] = rules*/
	case Action:
		// TODO
	case Prop:
		if k := terp.getKind(v.Object); k != nil {
			pk := terp.getKind(v.Kind)
			if pk != nil {
				if pk.Parent == PKindEnum {
					k.addProp(&PVar{
						Pos:      v.Pos,
						Name:     v.Name,
						Kind:     pk.Parent,
						Val:      pk.DefVal,
						EnumVals: pk.ValidVals,
					})
				} else {
					k.addProp(&PVar{
						Pos:  v.Pos,
						Name: v.Name,
						Kind: pk,
						Val:  pk.DefVal,
					})
				}

				return
			}

			// has [prop] [propval]
			parts := strings.Split(v.Kind, " ")
			if len(parts) != 2 {
				panic(fmt.Sprintf("%v: prop not understood: %v", v.Pos, v.Kind))
			}
			p := k.getProp(parts[0])
			if p == nil {
				panic(fmt.Sprintf("%v: %v doesn't have prop: %v", v.Pos, v.Object, parts[0]))
			}
			p.Val = parts[1] // TODO: parse val?
			return
		}

		if o := terp.getObject(v.Object); o != nil {
			pk := terp.getKind(v.Kind)
			if pk != nil {
				o.addProp(&PVar{
					Pos:  v.Pos,
					Name: v.Name,
					Kind: pk,
					Val:  pk.DefVal,
				})
				return
			}

			// has [prop] [propval]
			parts := strings.Split(v.Kind, " ")
			if len(parts) != 2 {
				panic(fmt.Sprintf("%v: prop not understood: %v", v.Pos, v.Kind))
			}
			p := o.getProp(parts[0])
			if p == nil {
				panic(fmt.Sprintf("%v: %v doesn't have prop: %v", v.Pos, v.Object, parts[0]))
			}
			p.Val = parts[1] // TODO: parse val? also consider p.setVal
			return
		}

		panic(fmt.Sprintf("%v: prop for undefined object: %v", v.Pos, v.Object))
	case PropVal:
		// TODO
	case PropEnum:
		// TODO: len vals == 1 should be bool
		k := terp.getKind(v.Object)
		if k != nil {
			name := v.Name
			if name == "" {
				name = fmt.Sprintf("%v-enum-%v", k.Name, strings.Join(v.Vals, ","))
			}
			k.addProp(&PVar{
				Pos:      v.Pos,
				Name:     name,
				Val:      len(v.Vals) - 1,
				Kind:     PKindEnum,
				EnumVals: v.Vals,
			})
			return
		}

		o := terp.getObject(v.Object)
		if o != nil {
			name := v.Name
			if name == "" {
				name = fmt.Sprintf("%v-enum-%v", o.Name, strings.Join(v.Vals, ","))
			}
			o.addProp(&PVar{
				Pos:      v.Pos,
				Name:     name,
				Val:      len(v.Vals) - 1,
				Kind:     PKindEnum,
				EnumVals: v.Vals,
			})
			return
		}

		panic(fmt.Sprintf("%v: %s is not defined", v.Pos, v.Object))

	case Kind:
		par := terp.getKind(v.Kind)
		if par == nil {
			panic(fmt.Sprintf("%v: parent kind undefined: %v", v.Pos, v.Kind))
		}

		terp.addKind(&PKind{
			Pos:    v.Pos,
			Name:   v.Name,
			Parent: par,
		})
	case ThereAre:
		// TODO
	case QuotedString:
		// TODO
	case Vector:
		// TODO
	case Relation:
		// TODO
	case Verb:
		// TODO
	default:
		panic(fmt.Sprintf("unknown sentence %T", v))
	}
}

func (terp *Interp) secondStage() error {
	// TODO: check dup kinds
	// TODO: sort kinds
	// TODO: sort dict
	// TODO: parse funcs and implode kinds in dict
	// TODO: test for props on kinds that don't have such props

	/*sort.Slice(terp.tmp.Props, func(i, j int) bool {
		if terp.tmp.Props[i].Object != terp.tmp.Props[j].Object {
			return terp.tmp.Props[i].Object < terp.tmp.Props[j].Object
		}
		if terp.tmp.Props[i].Name != terp.tmp.Props[j].Name {
			return terp.tmp.Props[i].Name < terp.tmp.Props[j].Name
		}
		return terp.tmp.Props[i].Kind < terp.tmp.Props[j].Kind
	})*/

	/*for _, d := range terp.Defs {
		if d.RawCond.items != nil {
			e, err := ParseExpr("TODO", d.RawCond, nil)
			if err != nil {
				return err
			}
			pretty.Fprintf(tty, "%v\n", e)
		}
	}*/

	/*for rb, rules := range terp.tmp.Rulebooks {
		nrules := terp.Rulebooks[rb]
		for _, r := range rules {
			nr := PRule{
				Pos: r.Pos,
			}
			e, err := ParseExpr(r.RawCond, nil)
			if err != nil {
				return err
			}
			nr.Cond = e
			phs, err := ParsePhrases(r.RawPhrases, nil)
			if err != nil {
				return err
			}
			nr.Phrases = phs
			nrules = append(nrules, nr)
		}
		terp.Rulebooks[rb] = nrules
	}*/

	return nil
}

func (terp *Interp) mainLoop() error {
	for {
		line, err := terp.tty.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		line = strings.TrimSpace(line)
		cmd := line
		if i := strings.IndexByte(line, ' '); i > 0 {
			cmd = line[:i]
			line = line[i+1:]
		}
		switch cmd {
		case "q", "quit":
			return fmt.Errorf("quit") // TODO: change to nil
		case "v", "vars":
			for k, v := range terp.Vars {
				arr := ""
				if v.Array {
					arr = "[]"
				}
				fmt.Fprintf(terp.tty, "%v: %v%v = %v\n", k, arr, v.Kind, v.Val)
			}
		case "k", "kinds":
			for _, k := range terp.Kinds {
				fmt.Fprintf(terp.tty, "%v", k.Name)
				for p := k.Parent; p != nil; {
					fmt.Fprintf(terp.tty, " -> %v", p.Name)
					p = p.Parent
				}
				fmt.Fprintf(terp.tty, "\n")
			}
		case "t", "tables":
			for n, t := range terp.Tables {
				for i, row := range t.Rows {
					fmt.Fprintf(terp.tty, "%v[%v]: %v\n", n, i, row)
				}
			}
		case "p", "props":
			/*for _, p := range terp.tmp.props {
				arr := ""
				if p.Array {
					arr = "[]"
				}
				fmt.Fprintf(terp.tty, "%v.%v: %v%v\n", p.Object, p.Name, arr, p.Kind)
			}*/
		case "ve", "verbs":
		case "e", "eval":
			e, err := ParseExpr(lex("tty", line), terp.dict)
			if err != nil {
				fmt.Fprintf(terp.tty, "err: %v\n", err)
				continue
			}
			r := terp.EvalExpr(e)
			if r != nil {
				fmt.Fprintf(terp.tty, "%+#v", r)
			}
			fmt.Fprintf(terp.tty, "\n")
		case "go":
			// run When play begins
		}
	}
}

func (terp *Interp) EvalExpr(e interface{}) interface{} {
	if e == nil {
		return nil
	}

	switch e := e.(type) {
	case PFuncCall:
		if e.Func.NativeFn != nil {
			return e.Func.NativeFn(e.Args)
		}
		return nil
	case *PVar:
		return e.Val
	case PExprOp:
		a := terp.EvalExpr(e.Left).(int)
		b := terp.EvalExpr(e.Right).(int)
		var r interface{}
		switch e.Op {
		case '+':
			r = a + b
		case '-':
			r = a - b
		case '*':
			r = a * b
		case '/':
			r = a / b
		case '>':
			r = a > b
		case '<':
			r = a < b
		}
		return r
	case PQuotedString:
		buf := new(strings.Builder)
		buf.Grow(20)
		for _, a := range e.Parts {
			fmt.Fprint(buf, terp.EvalExpr(a))
		}
		return buf.String()
	case string:
		return e
	case int:
		return e
	case float64:
		return e
	case bool:
		return e
	}

	panic(fmt.Sprintf("bug: unhandled expr type: %T", e))
}
