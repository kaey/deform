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
	Rows        [][]interface{} // int, string or QuotedString
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

// [Name] is a [list of] [Kind] that varies.
type Variable struct {
	Pos   Pos
	Name  string
	Kind  string
	Array bool
}

// A [Name] is a kind of [Kind].
type Kind struct {
	Pos  Pos
	Name string
	Kind string
}

// [Object] has a [Kind] called [Name].
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

// [Prop] of [Object] is [usually] [Val].
type PropVal struct {
	Pos     Pos
	Prop    string
	Object  string
	Val     interface{} // int, string or QuotedString
	Usually bool
}

// [Object] can be [Vals] (this is the [Name] property).
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
	Pos      Pos
	Name     string
	Alts     []string // alternative verb forms
	Reversed bool
	Rel      string
}

type Vector struct {
	Pos     Pos
	Pattern string
	Kind    string
	Parts   []string
}

// [Article] [Object] is [Direction] [usually|initially|always|never] [Value].
type Is struct {
	Pos       Pos
	Article   string
	Object    string
	Value     interface{} // bool, int, string or QuotedString
	EnumVal   []string
	Direction string // for rooms only
	Usually   bool
	Initially bool
	Always    bool
	Never     bool
	Negate    bool
}

// [N] [Objects] is in [Where].
type IsIn struct {
	Pos     Pos
	Objects []string
	Where   string
	Decl    bool // Implicit declaration of N objects.
}
