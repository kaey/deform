package main

type PVar struct {
	Pos   Pos
	Name  string
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
	Cond    interface{}
	Phrases []Phrase
}

type PDefinition struct {
	Pos     Pos
	Expr    interface{}
	Phrases []Phrase
}

type PFunc struct {
	Name     string
	NativeFn func([]interface{}) interface{}
}

type PFuncCall struct {
	Func *PFunc
	Args []interface{}
}

type PExprOp struct {
	Op    byte
	Left  interface{}
	Right interface{}
}

type PQuotedString struct {
	Parts []interface{}
}

type PObject struct{}

type Phrase interface{}

type PhraseListAdd struct {
	Pos   Pos
	Value interface{}
	List  string
}

type PhraseLet struct {
	Pos    Pos
	Object string
	Value  interface{}
}

type PhraseWhile struct {
	Pos     Pos
	Expr    interface{}
	Phrases []Phrase
}

type PhraseIf struct {
	Pos     Pos
	Expr    interface{}
	Phrases []Phrase
}

type PhraseOtherwiseIf struct {
	Pos     Pos
	Expr    interface{}
	Phrases []Phrase
}

type PhraseSay struct {
	Pos Pos
	Say interface{} // contains QuotedString or PhraseFuncCall
}

type PhraseDecide struct {
	Pos    Pos
	Result string
}

type PhraseFuncCall struct {
	Pos  Pos
	Func interface{}
}
