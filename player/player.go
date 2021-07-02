package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaey/deform/parse"
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
		ss, err := parse.ParseFile(f)
		if err != nil {
			return err
		}
		for _, s := range ss {
			terp.evalSentence(s)
		}
	}

	for {
		line, err := tty.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch strings.TrimSpace(line) {
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
				for p := v.Parent; p != ""; {
					fmt.Fprintf(tty, " -> %v", p)
					p = terp.Kinds[p].Parent
				}
				fmt.Fprintf(tty, "\n")
			}
		case "t", "tables":
			for n, t := range terp.Tables {
				for i, row := range t.Rows {
					fmt.Fprintf(tty, "%v[%v]: %v\n", n, i, row)
				}
			}
		}
	}
}

type Interp struct {
	tty *term.Terminal

	Vars      map[string]Var
	Tables    map[string]Table
	Rulebooks map[string]Rulebook
	Actions   map[string]Action
	Kinds     map[string]Kind
}

func newInterp(tty *term.Terminal) *Interp {
	terp := &Interp{
		tty:       tty,
		Vars:      make(map[string]Var),
		Tables:    make(map[string]Table),
		Rulebooks: make(map[string]Rulebook),
		Actions:   make(map[string]Action),
		Kinds:     make(map[string]Kind),
	}

	terp.Vars["maximum score"] = Var{Kind: "number"}
	terp.Kinds["object"] = Kind{}
	terp.Kinds["direction"] = Kind{Parent: "object"}
	terp.Kinds["room"] = Kind{Parent: "object"}
	terp.Kinds["region"] = Kind{Parent: "object"}
	terp.Kinds["thing"] = Kind{Parent: "object"}
	terp.Kinds["door"] = Kind{Parent: "thing"}
	terp.Kinds["container"] = Kind{Parent: "thing"}
	terp.Kinds["player's holdall"] = Kind{Parent: "container"}
	terp.Kinds["supporter"] = Kind{Parent: "thing"}
	terp.Kinds["backdrop"] = Kind{Parent: "thing"}
	terp.Kinds["device"] = Kind{Parent: "thing"}
	terp.Kinds["person"] = Kind{Parent: "thing"}
	terp.Kinds["man"] = Kind{Parent: "person"}
	terp.Kinds["woman"] = Kind{Parent: "person"}
	terp.Kinds["animal"] = Kind{Parent: "person"}
	return terp
}

func (terp *Interp) evalSentence(s parse.Sentence) {
	switch v := s.(type) {
	case parse.Subheader:
	case parse.Figure:
	case parse.FileOf:
	case parse.Understand:
	case parse.DoesThePlayerMean:
	case parse.RuleFor:
	case parse.Func:
	case parse.Table:
		name := strings.ToLower(v.Name)
		t := terp.Tables[name]
		// TODO: col kinds
		for _, row := range v.Rows {
			m := make(map[string]string, len(v.ColNames))
			for i, colname := range v.ColNames {
				m[colname] = row[i]
			}
			t.Rows = append(t.Rows, m)
		}
		terp.Tables[name] = t
	case parse.Variable:
		name := strings.ToLower(v.Name)
		if _, exists := terp.Vars[name]; exists {
			panic(fmt.Sprintf("var exists: %v", v.Name))
		}
		terp.Vars[name] = Var{
			Kind:  strings.ToLower(v.Kind),
			Array: v.Array,
		}
	case parse.Definition:
	case parse.Is:
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
			terp.Rulebooks[name] = Rulebook{}
			return
		}

		// panic(fmt.Sprintf("does not exist: %+#v", v))
	case parse.IsIn:
	case parse.ListedInRulebook:
		// panic(fmt.Sprintf("listed: %v", v))
	case parse.Rule:
	case parse.Action:
		name := strings.ToLower(v.Name)
		terp.Actions[name] = Action{
			NThings:   v.NThings,
			Touchable: v.Touchable,
		}
	case parse.Prop:
	case parse.PropVal:
	case parse.PropEnum:
	case parse.Kind:
		terp.Kinds[strings.ToLower(v.Name)] = Kind{Parent: strings.ToLower(v.Kind)}
	case parse.ThereAre:
	case parse.RoomDescr:
	case parse.Vector:
	case parse.Relation:
	case parse.Verb:
	default:
		panic(fmt.Sprintf("unknown sentence %T", v))
	}
}

type Var struct {
	Val   string
	Kind  string
	ValA  []string
	Array bool
}

type Table struct {
	Kinds map[string]string
	Rows  []map[string]string
}

type Rulebook struct {
	Rules []string
}

type Action struct {
	NThings   int
	Touchable bool
}

type Kind struct {
	Parent string
}
