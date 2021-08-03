package main

import (
	"strconv"
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
				panic(r)
			}
			err = p.err
			// panic(p.err)
		}
	}()

	e = p.parseExpr()
	return
}

func (p *Parser) parseExpr() interface{} {
	e := p.parseExprOne()
	for {
		ee := p.parseExprBinary(e)
		if ee == nil {
			break
		}
		e = ee
	}

	return e
}

func (p *Parser) parseExprOne() interface{} {
	if p.parseLeftParen() {
		e := p.parseExpr()
		p.mustParseRightParen()
		return e
	}

	iti := p.iti

	if q, ok := p.parseQuotedString(); ok {
		args := make([]interface{}, 0, 3)
		for _, pt := range q.Parts {
			switch pt := pt.(type) {
			case string:
				args = append(args, pt)
			case RawFuncCall:
				e, err := ParseExpr(lex("string", string(pt)), p.dict)
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
		}
	}

	if w, ok := p.parseWord(); ok {
		// TODO: parse float
		n, err := strconv.Atoi(w)
		if err == nil {
			return n
		}
		p.backup()
	}

	for _, row := range p.dict.rows {
		e, ok := p.parseExprRow(row)
		if !ok {
			p.iti = iti
			continue
		}
		return e
	}

	it := p.next()
	p.errorf("expression not understood: %s", it.val)
	panic("unreachable")
}

func (p *Parser) parseExprRow(row mrow) (interface{}, bool) {
	e := row.target
	fne := PFuncCall{}
	fn, isfn := e.(*PFunc)

	for _, n := range row.n {
		if n.kind {
			if !isfn {
				p.errorf("match for kind in non-function")
			}
			arg := p.parseExpr() // TODO: check kind
			fne.Args = append(fne.Args, arg)
			continue
		}

		if w, ok := p.parseWord(); !ok || n.match != w {
			return nil, false
		}
	}

	if isfn {
		fne.Func = fn
		return fne, true
	}
	return e, true
}

func (p *Parser) parseExprBinary(e interface{}) interface{} {
	it := p.next()
	if it.typ != itemWord {
		p.backup()
		return nil
	}

	switch it.val {
	case "+":
		e2 := p.parseExprOne()
		return PExprOp{
			Op:    '+',
			Left:  e,
			Right: e2,
		}
	case "-":
		e2 := p.parseExprOne()
		return PExprOp{
			Op:    '-',
			Left:  e,
			Right: e2,
		}
	case "*":
		e2 := p.parseExprOne()
		return PExprOp{
			Op:    '*',
			Left:  e,
			Right: e2,
		}
	case "/":
		e2 := p.parseExprOne()
		return PExprOp{
			Op:    '/',
			Left:  e,
			Right: e2,
		}
	case ">": // TODO: func calls shouldn't consume this shit.
		e2 := p.parseExpr()
		return PExprOp{
			Op:    '>',
			Left:  e,
			Right: e2,
		}
	case "<":
		e2 := p.parseExpr()
		return PExprOp{
			Op:    '<',
			Left:  e,
			Right: e2,
		}
	case "of":
	}

	p.backup()
	return nil
}
