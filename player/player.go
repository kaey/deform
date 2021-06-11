package main

import (
	"fmt"
	"os"

	"github.com/kaey/deform/parse"
	"golang.org/x/term"
)

func main() {
	if err := Main(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func Main() error {
	tree, err := parse.ParseFile("../testdata/aaa.i7x")
	if err != nil {
		return err
	}

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

	interp, err := evalBook(tty, tree)
	if err != nil {
		return err
	}

	_ = interp

	return nil
}

type Interp struct {
	tty     *term.Terminal
	Kinds   map[string]Kind
	Objects map[string]Object
	Rules   map[string]Rule
}

func evalBook(tty *term.Terminal, tree interface{}) (*Interp, error) {
	book := tree.(parse.Book)
	if book.Header.Title != book.Footer.Title {
		return nil, fmt.Errorf("expected title %q in footer, got %q", book.Header.Title, book.Footer.Title)
	}

	in := &Interp{
		tty:     tty,
		Kinds:   make(map[string]Kind, 100),
		Objects: make(map[string]Object, 100),
		Rules:   make(map[string]Rule, 100),
	}

	for _, s := range book.Body.Sentences {
		in.eval(s)
	}

	for _, s := range in.Rules["When play begins"].What {
		in.eval(s)
	}

	return in, nil
}

func (in *Interp) eval(sentence parse.Sentence) {
	switch v := sentence.(type) {
	case parse.StmtNewKind:
	case parse.StmtNewInstance:
	case parse.StmtSay:
		fmt.Fprint(in.tty, v.Str)
	case parse.Rule:
		in.Rules[v.When] = Rule{
			When: v.When,
			What: v.What,
		}
	default:
		panic(fmt.Sprintf("unknown sentence %+#v", v))
	}
}

type Kind struct {
	Name   string
	Parent *Kind
	Props  map[string]Value
}

type Object struct {
	Name  string
	Props map[string]Value
}

type Value struct {
	Type ValueType
	Int  int
	Str  string
}

type ValueType int

const (
	ValueTypeInt = iota
	ValueTypeBool
	ValueTypeStr
)

type Rule struct {
	When string
	What []parse.Stmt
}
