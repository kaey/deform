package parse

type Comment string

type Sentence interface{}

type UnknownSentence string

type RoomDescr string

type Subheader string

type Definition struct {
	Object     string
	Called     string
	Prop       string
	Cond       Expr
	rawPhrases []rawPhrase
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
	When    string // Before, instead, check, carry out, after, report.
	Name    string // This is foo rule.
	Action  string
	Cond    string // When.
	Phrases []Phrase
	Comment Comment

	rawPhrases []rawPhrase
	rawCond    rawExpr
}

type Decl interface{}

type DeclDefVal struct {
	Object string
	Val    string
}

type DeclKind struct {
	Object string
	Kind   string
}

type DeclInstance struct {
	Object string
	Kind   string
}

type DeclProp struct {
	Object string
	Kind   string
	Prop   string
}

type DeclPropEnum struct {
	Object string
	Enum   []string
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
