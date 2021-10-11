package main

import (
	"log"
	"strconv"

	"go4.org/intern"
)

type Memo struct {
	Target interface{}
	Idx    int
}

func ParseExpr(items []item, dict *Dict) (e interface{}, err error) {
	p := &Parser{
		items: items,
		iti:   -1, // calling next() will point to first available item
		dict:  dict,
		memo:  make(map[int]Memo, 5),
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
		e2 := p.parseExprIdent(fnarg)
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
patloop:
	for _, pat := range p.dict.binary {
		p.iti = iti

		for _, part := range pat.iparts {
			if it := p.next(); it.typ != itemWord || part != it.ival {
				continue patloop
			}
		}

		return &PExprBinary{Op: pat.target}
	}

	p.iti = iti
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

var ifnarg = intern.GetByString("@")

func (p *Parser) parseExprIdent(fnarg bool) interface{} {
	p.parseArticle()
	iti := p.iti

	if fnarg {
		if m, ok := p.memo[p.iti]; ok {
			p.iti = m.Idx
			return m.Target
		}
	}

	if p.parseLeftParen() {
		e := p.parseExpr1()
		p.mustParseRightParen()
		return p.memoize(e, fnarg, iti)
	}

	if q, ok := p.parseQuotedString(); ok {
		return p.memoize(q, fnarg, iti)
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
			return p.memoize(true, fnarg, iti)
		} else if w == "false" {
			return p.memoize(false, fnarg, iti)
		} else if w == "zero" {
			return p.memoize(0, fnarg, iti)
		} else if w == "not" {
			return p.memoize("not", fnarg, iti) // TODO: proper negate op
		}
		if len(w) > 2 && w[0] == '<' {
			return p.memoize(w, fnarg, iti) // TODO: vector
		}
		n, err := strconv.Atoi(w)
		if err == nil {
			return p.memoize(n, fnarg, iti)
		}
		f, err := strconv.ParseFloat(w, 64)
		if err == nil {
			return p.memoize(f, fnarg, iti)
		}
		p.backup()
	}

	fnargs := make([]interface{}, 0, 2)
patloop:
	for _, pat := range p.dict.ident {
		p.iti = iti
		fnargs = fnargs[:0]

		for i, part := range pat.iparts {
			if i == 0 && fnarg && part == ifnarg {
				continue patloop
			}
			if part == ifnarg {
				arg := p.parseExpr3(true)
				if arg == nil {
					continue patloop
				}
				fnargs = append(fnargs, arg)
				continue
			}

			if it := p.next(); it.typ != itemWord || part != it.ival {
				continue patloop
			}
		}

		if len(fnargs) > 0 {
			return p.memoize(&PCall{
				Target: pat.target,
				Args:   fnargs,
			}, fnarg, iti)
		}
		return p.memoize(pat.target, fnarg, iti)
	}

	p.iti = iti
	return nil
}

func (p *Parser) memoize(e interface{}, fnarg bool, iti int) interface{} {
	if fnarg {
		p.memo[iti] = Memo{
			Target: e,
			Idx:    p.iti,
		}
	}

	return e
}
