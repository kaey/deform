package parse

import "github.com/kr/pretty"

func (p *Parser) parseExpr() Expr {
	e := p.parseExprE()
	rest := p.parseExprRest()
	pretty.Println(e, rest)

	return Expr{
		E:    e,
		Rest: rest,
	}
}

func (p *Parser) parseExprE() interface{} {
	if p.parseLeftParen() {
		e := p.parseExpr()
		p.mustParseRightParen()
		return e
	}

	return p.parseWordsUntil("is", "and", "or", "+", "-", "*", "/") // TODO: Number / Ident
}

func (p *Parser) parseExprRest() []ExprPart {
	rest := make([]ExprPart, 0, 3)
	for {
		it := p.peek()
		if it.typ != itemWord && it.typ != itemLeftParen && it.typ != itemRightParen {
			break
		}
		op := p.parseOp()
		e := p.parseExprE()
		rest = append(rest, ExprPart{op, e})
	}

	return rest
}

func (p *Parser) parseOp() Op {
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
	case "not":
		return Op('!')
	}

	return 0
}
