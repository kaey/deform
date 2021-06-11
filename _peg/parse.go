package parse

import (
	"strconv"
)

//go:generate pigeon -o zparser.go -no-recover rules.peg

func sli(v interface{}) []interface{} {
	if v == nil {
		return nil
	}

	return v.([]interface{})
}

func str(v interface{}) string {
	if v == nil {
		return ""
	}

	return v.(string)
}

func slstmt(v interface{}) []Stmt {
	if v == nil {
		return nil
	}

	vv := v.([]interface{})
	r := make([]Stmt, len(vv))
	for i := range vv {
		r[i] = vv[i].(Stmt)
	}

	return r
}

func slstr(v interface{}) []string {
	if v == nil {
		return nil
	}

	vv := v.([]interface{})
	r := make([]string, len(vv))
	for i := range vv {
		r[i] = vv[i].(string)
	}

	return r
}

func slfnargs(v interface{}) []FuncArg {
	if v == nil {
		return nil
	}

	vv := v.([]interface{})
	r := make([]FuncArg, len(vv))
	for i := range vv {
		r[i] = vv[i].(FuncArg)
	}

	return r
}

type Extension struct {
	Header Header
	Body   Body
	Footer Footer
}

func newExtension(header, body, footer interface{}) (Extension, error) {
	return Extension{
		Header: header.(Header),
		Body:   body.(Body),
		Footer: footer.(Footer),
	}, nil
}

type Header struct {
	Title  string
	Author string
}

func newHeader(title, author interface{}) (Header, error) {
	return Header{
		Title:  str(title),
		Author: str(author),
	}, nil
}

type Footer struct {
	Title string
}

func newFooter(title interface{}) (Footer, error) {
	return Footer{
		Title: str(title),
	}, nil
}

type Body struct {
	Sentences []Sentence
}

func newBody(s interface{}) (Body, error) {
	ss := sli(s)
	rs := make([]Sentence, len(ss))
	for i := range ss {
		rs[i] = ss[i].(Sentence)
	}

	return Body{
		Sentences: rs,
	}, nil
}

type Sentence interface{}

type Subheader struct {
	S string
}

func newSubheader(s []byte) (Subheader, error) {
	return Subheader{S: string(s)}, nil
}

type Definition struct {
	Object string
	Called string
	Prop   string
	Stmts  []Stmt
}

func newDefinition(obj, called, prop, stmts interface{}) (Definition, error) {
	return Definition{
		Object: str(obj),
		Called: str(called),
		Prop:   str(prop),
		Stmts:  slstmt(stmts),
	}, nil
}

type Report struct {
	Action string
	Cond   string
	Stmts  []Stmt
}

func newReport(act, cond, stmts interface{}) (Report, error) {
	return Report{
		Action: act.(string),
		Cond:   cond.(string),
		Stmts:  slstmt(stmts),
	}, nil
}

type Rule struct {
	When  string
	Stmts []Stmt
}

func newRule(when, stmts interface{}) (Rule, error) {
	return Rule{
		When:  when.(string),
		Stmts: slstmt(stmts),
	}, nil
}

type Decl interface{}

type DeclDefVal struct {
	Object string
	Val    string
}

func newDeclDefVal(obj, val interface{}) (DeclDefVal, error) {
	return DeclDefVal{
		Object: obj.(string),
		Val:    val.(string),
	}, nil
}

type DeclKind struct {
	Object string
	Kind   string
}

func newDeclKind(obj, kind interface{}) (DeclKind, error) {
	return DeclKind{
		Object: obj.(string),
		Kind:   kind.(string),
	}, nil
}

type DeclInstance struct {
	Object string
	Kind   string
}

func newDeclInstance(obj, kind interface{}) (DeclInstance, error) {
	return DeclInstance{
		Object: obj.(string),
		Kind:   kind.(string),
	}, nil
}

type DeclProp struct {
	Object string
	Kind   string
	Prop   string
}

func newDeclProp(obj, kind, prop interface{}) (DeclProp, error) {
	return DeclProp{
		Object: obj.(string),
		Kind:   kind.(string),
		Prop:   prop.(string),
	}, nil
}

type DeclPropEnum struct {
	Object string
	Enum   []string
}

func newDeclPropEnum(obj, first interface{}, rest interface{}) (DeclPropEnum, error) {
	return DeclPropEnum{
		Object: obj.(string),
		Enum:   append([]string{first.(string)}, slstr(rest)...),
	}, nil
}

type Func interface{}

type FuncSay struct {
	Name  string
	Arg   *FuncArg
	Args  []FuncArg
	Stmts []Stmt
}

func newFuncSay(name, arg, args, stmts interface{}) (FuncSay, error) {
	var a FuncArg
	if arg != nil {
		a = arg.(FuncArg)
	}
	return FuncSay{
		Name:  name.(string),
		Arg:   &a,
		Args:  slfnargs(args),
		Stmts: slstmt(stmts),
	}, nil
}

type FuncOther struct {
	Name  string
	Arg   *FuncArg
	Args  []FuncArg
	Stmts []Stmt
}

func newFuncOther(name, arg, args, stmts interface{}) (FuncOther, error) {
	var a FuncArg
	if arg != nil {
		a = arg.(FuncArg)
	}
	return FuncOther{
		Name:  name.(string),
		Arg:   &a,
		Args:  slfnargs(args),
		Stmts: slstmt(stmts),
	}, nil
}

type FuncArg struct {
	Name string
	Kind string
}

func newFuncArg(name, kind interface{}) (FuncArg, error) {
	return FuncArg{
		Name: name.(string),
		Kind: kind.(string),
	}, nil
}

type Stmt interface{}

type StmtSay struct {
	Str string
}

func newStmtSay(str interface{}) (StmtSay, error) {
	return StmtSay{
		Str: str.(string),
	}, nil
}

type StmtDecide struct {
	Result string
}

func newStmtDecide(result interface{}) (StmtDecide, error) {
	return StmtDecide{
		Result: result.(string),
	}, nil
}

type StmtIf struct{}

type StmtDoNothing struct{}

func newStmtDoNothing() (StmtDoNothing, error) {
	return StmtDoNothing{}, nil
}

type Expr struct{}

func newExpr(e, rest interface{}) (Expr, error) {
	return Expr{}, nil
}

type Op byte

func newOp(v []byte) (Op, error) {
	return Op(v[0]), nil
}

type Ident string

func newIdent(v []byte) (Ident, error) {
	return Ident(v), nil
}

type Number int

func newNumber(v []byte) (Number, error) {
	s := string(v)
	if s == "true" {
		return 1, nil
	} else if s == "false" {
		return 0, nil
	}

	n, err := strconv.Atoi(s)
	return Number(n), err
}

type Comment string

func newComment(v []byte) (Comment, error) {
	return Comment(v), nil
}
