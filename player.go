package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/term"
)

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

	// Finalize
	sort.Slice(terp.tmp.Props, func(i, j int) bool {
		if terp.tmp.Props[i].Object != terp.tmp.Props[j].Object {
			return terp.tmp.Props[i].Object < terp.tmp.Props[j].Object
		}
		if terp.tmp.Props[i].Name != terp.tmp.Props[j].Name {
			return terp.tmp.Props[i].Name < terp.tmp.Props[j].Name
		}
		return terp.tmp.Props[i].Kind < terp.tmp.Props[j].Kind
	})

	/*for _, d := range terp.Defs {
		if d.RawCond.items != nil {
			e, err := ParseExpr("TODO", d.RawCond, nil)
			if err != nil {
				return err
			}
			pretty.Fprintf(tty, "%v\n", e)
		}
	}*/

	for rb, rules := range terp.tmp.Rulebooks {
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
			phs, err := ParsePhrases(r.RawPhrases)
			if err != nil {
				return err
			}
			nr.Phrases = phs
			nrules = append(nrules, nr)
		}
		terp.Rulebooks[rb] = nrules
	}

	for {
		line, err := tty.ReadLine()
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
				fmt.Fprintf(tty, "%v: %v%v = %v\n", k, arr, v.Kind, v.Val)
			}
		case "k", "kinds":
			for k, v := range terp.Kinds {
				fmt.Fprintf(tty, "%v", k)
				for p := v.Kind; p != ""; {
					fmt.Fprintf(tty, " -> %v", p)
					p = terp.Kinds[p].Kind
				}
				fmt.Fprintf(tty, "\n")
			}
		case "t", "tables":
			for n, t := range terp.Tables {
				for i, row := range t.Rows {
					fmt.Fprintf(tty, "%v[%v]: %v\n", n, i, row)
				}
			}
		case "p", "props":
			for _, p := range terp.tmp.Props {
				arr := ""
				if p.Array {
					arr = "[]"
				}
				fmt.Fprintf(tty, "%v.%v: %v%v\n", p.Object, p.Name, arr, p.Kind)
			}
		case "ve", "verbs":
			for k, v := range terp.Verbs {
				fmt.Fprintf(tty, "%v -> %v\n", k, v)
			}
		case "e", "eval":
			e, err := ParseExpr(lex("tty", line), nil)
			if err != nil {
				fmt.Fprintf(tty, "err: %v\n", err)
			}
			fmt.Fprintf(tty, "%+#v\n", EvalExpr(e))
		case "go":
			// run When play begins
		}
	}
}

type Interp struct {
	tty  *term.Terminal
	tmp  *interpTmp
	dict *Dict

	Vars      map[string]PVar
	Tables    map[string]PTable
	Rulebooks map[string][]PRule
	Actions   map[string]Action
	Kinds     map[string]Kind
	Verbs     map[string]string
}

type interpTmp struct {
	Props     []Prop
	Defs      []Definition
	Rulebooks map[string][]Rule
}

func newInterp(tty *term.Terminal) *Interp {
	terp := &Interp{
		tty: tty,
		tmp: &interpTmp{
			Props:     make([]Prop, 0, 100),
			Defs:      make([]Definition, 0, 100),
			Rulebooks: make(map[string][]Rule, 100),
		},
		dict:      &Dict{m: new(mnode)},
		Vars:      make(map[string]PVar, 100),
		Tables:    make(map[string]PTable, 100),
		Rulebooks: make(map[string][]PRule, 100),
		Actions:   make(map[string]Action, 100),
		Kinds:     make(map[string]Kind, 100),
		Verbs:     make(map[string]string, 100),
	}

	terp.Vars["maximum score"] = PVar{Kind: "number"}
	terp.Kinds["object"] = Kind{}
	terp.Kinds["direction"] = Kind{Kind: "object"}
	terp.Kinds["room"] = Kind{Kind: "object"}
	terp.Kinds["region"] = Kind{Kind: "object"}
	terp.Kinds["thing"] = Kind{Kind: "object"}
	terp.Kinds["door"] = Kind{Kind: "thing"}
	terp.Kinds["container"] = Kind{Kind: "thing"}
	terp.Kinds["player's holdall"] = Kind{Kind: "container"}
	terp.Kinds["supporter"] = Kind{Kind: "thing"}
	terp.Kinds["backdrop"] = Kind{Kind: "thing"}
	terp.Kinds["device"] = Kind{Kind: "thing"}
	terp.Kinds["person"] = Kind{Kind: "thing"}
	terp.Kinds["man"] = Kind{Kind: "person"}
	terp.Kinds["woman"] = Kind{Kind: "person"}
	terp.Kinds["animal"] = Kind{Kind: "person"}

	terp.dict.AddStub("clear the screen")

	return terp
}

func (terp *Interp) evalSentence(s Sentence) {
	switch v := s.(type) {
	case Subheader:
	case Figure:
	case FileOf:
	case Understand:
	case DoesThePlayerMean:
	case RuleFor:
	case Table:
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
		name := strings.ToLower(v.Name)
		if _, exists := terp.Vars[name]; exists {
			panic(fmt.Sprintf("var exists: %v", v.Name))
		}
		terp.Vars[name] = PVar{
			Pos:   v.Pos,
			Kind:  strings.ToLower(v.Kind),
			Array: v.Array,
		}
	case Is:
		name := strings.ToLower(v.Object)
		if _, exists := terp.Vars[name]; exists {
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
		// panic(fmt.Sprintf("does not exist: %+#v", v))
	case Func:

	case Definition:
		// TODO
		terp.tmp.Defs = append(terp.tmp.Defs, v)
	case IsIn:
		// fmt.Fprintf(terp.tty, "%v: %v\n", v.Where, v.Objects)
	case ListedInRulebook:
		// TODO: priority
		// panic(fmt.Sprintf("listed: %v", v))
	case Rule:
		if v.Rulebook == "" {
			// TODO
			return
		}
		rb := strings.ToLower(v.Rulebook)
		rules := terp.tmp.Rulebooks[rb]
		rules = append(rules, v)
		terp.tmp.Rulebooks[rb] = rules
	case Action:
		name := strings.ToLower(v.Name)
		terp.Actions[name] = v
	case Prop:
		terp.tmp.Props = append(terp.tmp.Props, v)
	case PropVal:
	case PropEnum:
	case Kind:
		terp.Kinds[strings.ToLower(v.Name)] = v
	case ThereAre:
	case QuotedString:
	case Vector:
	case Relation:
	case Verb:
		terp.Verbs[v.Name] = v.Rel
		for _, a := range v.Alts {
			terp.Verbs[a] = v.Rel
		}
	default:
		panic(fmt.Sprintf("unknown sentence %T", v))
	}
}

func (terp *Interp) match(v string) interface{} {
	parts := strings.Split(v, " ")
	c := terp.dict.m
	for _, p := range parts {
		c = c.getNode(p)
		if c == nil {
			panic("match fail")
		}
	}

	return c.target
}

func (terp *Interp) call() {
}

type PVar struct {
	Pos   Pos
	Val   interface{}
	Kind  string
	ValA  []string
	Array bool
}

type PTable struct {
	Pos   Pos
	Kinds map[string]string
	Rows  []map[string]interface{}
}

type PRule struct {
	Pos     Pos
	Cond    *Expr
	Phrases []Phrase
}

type PDefinition struct {
	Pos     Pos
	Expr    Expr
	Phrases []Phrase
}

type PDict struct{}

type PStubFunc struct{}

func EvalExpr(e *Expr) interface{} {
	if e == nil {
		return true
	}

	return evalExpr(*e)
}

func evalExpr(e Expr) interface{} {
	// TODO: unary op
	v := evalExprE(e.E)
	for _, p := range e.Rest {
		v2 := evalExpr(p.Expr)
		v = evalExprOp(v, p.Op, v2)
	}
	return v
}

func evalExprE(e interface{}) interface{} {
	switch ee := e.(type) {
	case string:
		// TODO: variables, props etc
		n, err := strconv.Atoi(ee)
		if err == nil {
			return n
		}
		return ee
	case Expr:
		return evalExpr(ee)
	}

	panic(fmt.Sprintf("bug: unknown expr type: %T", e))
}

func evalExprOp(v1 interface{}, op Op, v2 interface{}) interface{} {
	switch op {
	case '+':
		switch v1.(type) {
		case int:
			return v1.(int) + v2.(int)
		}
	case '-':
		switch v1.(type) {
		case int:
			return v1.(int) - v2.(int)
		}
	case '*':
		switch v1.(type) {
		case int:
			return v1.(int) * v2.(int)
		}
	case '/':
		switch v1.(type) {
		case int:
			return v1.(int) / v2.(int)
		}
	case '=':
		switch v1.(type) {
		case int:
			return v1.(int) == v2.(int)
		}
	case '&':
		switch v1.(type) {
		case int:
			return v1.(int) == v2.(int)
		}
	case '|':
		switch v1.(type) {
		case int:
			return v1.(int) > 0 || v2.(int) > 0
		}
	}
	panic(fmt.Sprintf("bug: unknown expr: v1=%v, op=%c, v2=%v", v1, op, v2))
}
