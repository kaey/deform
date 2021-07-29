package main

import "strconv"

func ParseExpr(items []item, dict map[string]string) (*Expr, error) {
	if len(items) == 0 {
		return nil, nil
	}

	p := &Parser{
		items: items,
		iti:   -1, // calling next() will point to first available item
	}

	defer func() {
		if r := recover(); r != nil {
			if r != errParse {
				panic(r)
			}
			panic(p.err)
		}
	}()

	e := p.parseExpr()
	return &e, p.err
}

func (p *Parser) parseExpr() Expr {
	u := p.parseExprUnaryOp()
	e := p.parseExprE()
	rest := p.parseExprRest()

	return Expr{
		Unary: u,
		E:     e,
		Rest:  rest,
	}
}

func (p *Parser) parseExprE() interface{} {
	if p.parseLeftParen() {
		e := p.parseExpr()
		p.mustParseRightParen()
		return e
	}

	if s, ok := p.parseExprQuotedString(); ok {
		return s
	}

	ident := p.mustParseWordsUntil("+", "-", "*", "/", "<", ">", "is", "and", "or", "not")
	n, err := strconv.Atoi(ident)
	if err == nil {
		return Number(n)
	}

	return Ident(ident)
}

func (p *Parser) parseExprRest() []ExprPart {
	rest := make([]ExprPart, 0, 3)
	for {
		it := p.peek()
		if it.typ != itemWord {
			break
		}
		op := p.parseExprOp()
		e := p.parseExpr()
		rest = append(rest, ExprPart{op, e})
	}

	return rest
}

func (p *Parser) parseExprOp() Op {
	op := p.mustParseWord()
	switch op {
	case "+", "-", "*", "/", "<", ">":
		return Op(op[0])
	case "is":
		return Op('=')
	case "and":
		return Op('&')
	case "or":
		return Op('|')
	}

	return 0
}

func (p *Parser) parseExprUnaryOp() Op {
	op, ok := p.parseWord()
	if !ok {
		return 0
	}

	switch op {
	case "not":
		return Op('!')
	}

	p.backup()
	return 0
}
