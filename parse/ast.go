package parse

type Comment string

type Sentence interface{}

type RoomDescr string

type Subheader string

type Definition struct {
	Object     string
	Called     string
	Prop       string
	Cond       Expr
	rawPhrases []rawPhrase
	rawCond    rawExpr
}

type Func struct {
	Parts   []FuncPart
	Comment Comment

	rawPhrases []rawPhrase
}

type FuncPart struct {
	Word string

	// If Word is empty, this is arg.
	ArgName string
	ArgKind string
}

type Rule struct {
	Prefix   string // Before, instead, check, carry out, after, report, for.
	Rulebook string
	Name     string // This is foo rule.
	Cond     string // When.
	Phrases  []Phrase
	Comment  Comment

	rawPhrases []rawPhrase
	rawCond    rawExpr
}

type Table struct {
	Name        string
	Continued   bool
	ColNames    []string
	ColKinds    []string
	RowComments []Comment
	Rows        [][]string
}

type Figure struct {
	Name     string
	FilePath string
}

type RuleFor string // Used for input parser and status line

type Understand string // Used for input parser

type DoesThePlayerMean string // Used for input parser

type ThereAre struct {
	N    int
	Kind string
}

type FileOf string

type ListedInRulebook struct {
	Rule     string
	Rulebook string
	Listed   bool
	First    bool
	Last     bool
}

type Variable struct {
	Name  string
	Kind  string
	Array bool
}

type Kind struct {
	Name string
	Kind string
}

type Prop struct {
	Object string
	Kind   string
	Name   string
}

type Action struct {
	Name      string
	NThings   int
	Touchable bool
}

type PropVal struct {
	Prop    string
	Object  string
	Val     string
	Usually bool
}

type PropEnum struct {
	Object string
	Name   string
	Vals   []string
}

type Relation struct {
	Name      string
	Object    string
	NObjects  int
	Kind      string
	Object2   string
	NObjects2 int
	Kind2     string

	rawCond rawExpr
}

type Verb struct {
	Name string
	Alts []string // alternative verb forms
	Rel  string
}

type Vector struct {
	Pattern string
	Kind    string
	Parts   []string
}

type Is struct {
	Object    string
	Value     string
	EnumVal   []string
	Direction string // for rooms only
	Usually   bool
	Initially bool
	Negate    bool
}

type IsIn struct {
	Objects []string
	Where   string
}

type Phrase interface{}

type PhraseSay string

type PhraseDecide struct {
	Result string
}

type PhraseIf struct{}

type PhraseDoNothing struct{}

type Expr struct {
	E    interface{}
	Rest []ExprPart
}

type ExprPart struct {
	Op Op
	E  interface{}
}

type Op byte

type Ident string

type Number int

type rawPhrase struct {
	Comment  Comment
	Pos      Pos
	items    []item
	children []rawPhrase
}

type rawExpr struct {
	Pos   Pos
	items []item
}
