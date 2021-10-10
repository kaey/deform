package main

import (
	"log"
	"strconv"
	"strings"
)

func ParseExpr(items []item, dict *Dict) (e interface{}, err error) {
	p := &Parser{
		items: items,
		iti:   -1, // calling next() will point to first available item
		dict:  dict,
	}

	defer func() {
		if r := recover(); r != nil {
			if r != errParse {
				log.Println(p.fmtItems()) // TODO
				panic(r)
			}
			// panic(p.err)
			err = p.err
		}
		/*if p.iti < len(p.items) {
			// TODO
			panic(fmt.Errorf("%v: %v %v %#+v", p.items[0].pos, p.iti, len(p.items), e))
		}*/
	}()

	e = p.parseExpr1()
	return
}

func (p *Parser) parseExpr1() interface{} {
	e := p.parseExpr2()

	if e == nil {
		p.errorf("expression not understood: %v", p.fmtItems())
	}

	for {
		b := p.parseExprBinaryLogic()
		if b == nil {
			break
		}
		e2 := p.parseExpr2()
		if e2 == nil {
			p.errorf("expression not understood: %v", p.fmtItems())
		}
		b.Left = e
		b.Right = e2
		e = b
	}

	return e
}

func (p *Parser) parseExpr2() interface{} {
	e := p.parseExpr3(false)
	if e == nil {
		p.errorf("expression not understood: %v", p.fmtItems())
	}

	for {
		b := p.parseExprBinaryRelation()
		if b == nil {
			break
		}
		e2 := p.parseExpr3(false)
		if e2 == nil {
			p.errorf("expression not understood: %v", p.fmtItems())
		}
		b.Left = e
		b.Right = e2
		e = b
	}

	return e
}

func (p *Parser) parseExpr3(fnarg bool) interface{} {
	e := p.parseExprIdent(fnarg)
	for {
		ee := p.parseExprIdent(fnarg)
		if ee == nil {
			break
		}
		e = ee // TODO: preserve all prefixes
	}
	if e == nil {
		if fnarg {
			return nil
		}
		p.errorf("expression not understood: %v", p.fmtItems())
	}
	for {
		b := p.parseExprBinaryOp()
		if b == nil {
			break
		}
		e2 := p.parseExprIdent(false)
		if e2 == nil {
			if fnarg {
				return nil
			}
			p.errorf("expression not understood: %v", p.fmtItems())
		}
		b.Left = e
		b.Right = e2
		e = b
	}

	return e
}

func (p *Parser) parseExprBinaryRelation() *PExprBinary {
	iti := p.iti
	for _, row := range p.dict.binary {
		if e := p.parseExprTryMatch(row, false); e != nil {
			return &PExprBinary{Op: e}
		}

		p.iti = iti
	}

	return nil
}

func (p *Parser) parseExprBinaryOp() *PExprBinary {
	w, ok := p.parseWord()
	if !ok {
		return nil
	}
	if w != "+" && w != "-" && w != "*" && w != "/" {
		p.backup()
		return nil
	}

	return &PExprBinary{Op: w}
}

func (p *Parser) parseExprBinaryLogic() *PExprBinary {
	w, ok := p.parseWord()
	if !ok {
		return nil
	}
	if w != "and" && w != "or" {
		p.backup()
		return nil
	}

	return &PExprBinary{Op: w}
}

func (p *Parser) parseExprIdent(fnarg bool) interface{} {
	if p.parseLeftParen() {
		e := p.parseExpr1()
		p.mustParseRightParen()
		return e
	}

	if q, ok := p.parseQuotedString(); ok {
		return q
		// TODO: parse strings
		/*args := make([]interface{}, 0, 3)
		for _, pt := range q.Parts {
			switch pt := pt.(type) {
			case string:
				args = append(args, pt)
			case RawFuncCall:
				e, err := ParseExpr(lex(q.Pos.String(), string(pt)), p.dict)
				if err != nil {
					p.errorf("%v", err)
				}
				args = append(args, e)
			default:
				p.errorf("invalid item in quoted string: %T", pt)
			}
		}

		return PQuotedString{
			Parts: args,
		}*/
	}

	if w, ok := p.parseWord(); ok {
		if w == "true" {
			return true
		} else if w == "false" {
			return false
		} else if w == "zero" {
			return 0
		} else if w == "not" {
			return "not" // TODO: proper negate op
		}
		if len(w) > 2 && w[0] == '<' {
			return w // TODO: vector
		}
		n, err := strconv.Atoi(w)
		if err == nil {
			return n
		}
		f, err := strconv.ParseFloat(w, 64)
		if err == nil {
			return f
		}
		p.backup()
	}

	iti := p.iti

	for _, row := range p.dict.ident {
		if e := p.parseExprTryMatch(row, fnarg); e != nil {
			return e
		}
		p.iti = iti
	}

	return nil
}

func (p *Parser) parseExprTryMatch(row mrow, fnarg bool) interface{} {
	fn, isfn := row.target.(*PFunc)
	var fnargs []interface{}

	p.parseArticle()
	for i, m := range row.match {
		if i == 0 && fnarg && m == "@" {
			return nil
		}
		if m == "@" {
			arg := p.parseExpr3(true)
			if arg == nil {
				return nil
			}
			fnargs = append(fnargs, arg)
			continue
		}

		if w, ok := p.parseWord(); !ok || !strings.EqualFold(m, w) {
			return nil
		}
	}

	if isfn { // TODO: Actions can have arguments as well.
		return &PFuncCall{
			Func: fn,
			Args: fnargs,
		}
	}
	return row.target
}
