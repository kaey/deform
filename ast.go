package main

type Sentence interface{}

type Comment struct {
	Pos Pos
	Str string
}

type QuotedString struct {
	Pos   Pos
	Parts []interface{} // contains string or RawFuncCall
}

type RawFuncCall string

type Subheader struct {
	Pos Pos
	Str string
}

type Definition struct {
	Pos        Pos
	Object     string
	Called     string
	Prop       string
	RawPhrases []item
	RawCond    []item
}

type Func struct {
	Pos        Pos
	Parts      []FuncPart
	Comment    Comment
	RawPhrases []item
}

type FuncPart struct {
	Word string

	// If Word is empty, this is arg.
	ArgName string
	ArgKind string
}

type Rule struct {
	Pos        Pos
	Rulebook   string
	Name       string // This is foo rule.
	Comment    Comment
	RawPhrases []item
	RawCond    []item
}

type Table struct {
	Pos         Pos
	Name        string
	Continued   bool
	ColNames    []string
	ColKinds    []string
	RowComments []Comment
	Rows        [][]interface{} // contains string or QuotedString
}

type Figure struct {
	Pos      Pos
	Name     string
	FilePath QuotedString
}

type RuleFor string // Used for input parser and status line

type Understand string // Used for input parser

type DoesThePlayerMean string // Used for input parser

type ThereAre struct {
	Pos  Pos
	N    int
	Kind string
}

type FileOf string

type ListedInRulebook struct {
	Pos      Pos
	Rule     string
	Rulebook string
	Listed   bool
	First    bool
	Last     bool
}

type Variable struct {
	Pos   Pos
	Name  string
	Kind  string
	Array bool
}

type Kind struct {
	Pos  Pos
	Name string
	Kind string
}

type Prop struct {
	Pos    Pos
	Object string
	Kind   string
	Array  bool
	Name   string
}

type Action struct {
	Pos       Pos
	Name      string
	NThings   int
	Touchable bool
}

type PropVal struct {
	Pos     Pos
	Prop    string
	Object  string
	Val     interface{} // string or QuotedString
	Usually bool
}

type PropEnum struct {
	Pos    Pos
	Object string
	Name   string
	Vals   []string
}

type Relation struct {
	Pos       Pos
	Name      string
	Object    string
	NObjects  int
	Kind      string
	Object2   string
	NObjects2 int
	Kind2     string
	RawCond   []item
}

type Verb struct {
	Pos  Pos
	Name string
	Alts []string // alternative verb forms
	Rel  string
}

type Vector struct {
	Pos     Pos
	Pattern string
	Kind    string
	Parts   []string
}

type Is struct {
	Pos       Pos
	Object    string
	Value     interface{} // string or QuotedString
	EnumVal   []string
	Direction string // for rooms only
	Usually   bool
	Initially bool
	Negate    bool
}

type IsIn struct {
	Pos     Pos
	Objects []string
	Where   string
}
