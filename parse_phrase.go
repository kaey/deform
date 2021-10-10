package main

import "log"

func ParsePhrases(items []item, dict *Dict) (phs []Phrase, err error) {
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
	}()

	phs = p.parsePhrases()
	return
}

func (p *Parser) parsePhrases() []Phrase {
	if it := p.peek(); it.typ == itemWord {
		return []Phrase{p.parsePhrase()}
	}

	p.parseComment()
	p.parseNL()
	p.indent++
	phs := make([]Phrase, 0, 3)
	for {
		if it := p.peek(); it.typ == itemSentenceEnd || it.typ == itemEOF {
			break
		}
		in := p.parseIndent()
		if in < p.indent {
			p.backup()
			break
		} else if in > p.indent {
			p.errorf("expected indent %v, got %v", p.indent, in)
		}
		ph := p.parsePhrase()
		p.parseSemicolon()
		p.parseNL()
		if ph == nil {
			continue
		}
		phs = append(phs, ph)
	}

	if len(phs) == 0 {
		p.errorf("expected at least 1 phrase: %v", p.fmtItems())
	}
	p.indent--
	return phs
}

func (p *Parser) parsePhrase() Phrase {
	if p.peek().typ == itemComment {
		return p.parseComment()
	}

	switch p.peek().val {
	case "while":
		it := p.next()
		ph := PhraseWhile{
			Pos: it.pos,
		}
		ph.Expr = p.parsePhraseFuncCall()
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	case "if":
		it := p.next()
		ph := PhraseIf{
			Pos: it.pos,
		}
		ph.Expr = p.parseExpr1()
		if p.parseComma() {
			ph.Phrases = []Phrase{p.parsePhrase()}
			return ph
		}
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	case "unless":
		it := p.next()
		ph := PhraseIf{
			Pos: it.pos,
			Neg: true,
		}
		ph.Expr = p.parseExpr1()
		if p.parseComma() {
			ph.Phrases = []Phrase{p.parsePhrase()}
			return ph
		}
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	case "otherwise":
		it := p.next()
		ph := PhraseOtherwiseIf{
			Pos: it.pos,
		}
		if p.parseWordOneOf("if") {
			ph.Expr = p.parseExpr1()
		}
		if p.peek().typ == itemWord {
			ph.Phrases = []Phrase{p.parsePhrase()}
			return ph
		}
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	case "repeat":
		it := p.next()
		ph := PhraseRepeat{
			Pos: it.pos,
		}
		p.mustParseWordOneOf("with")
		ph.Local = p.mustParseWord()
		p.mustParseWordOneOf("running")
		p.mustParseWordOneOf("through")
		p.dict.Add("local var", ph.Local)
		p.dict.Sort()
		ph.List = p.parseExpr1()
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	case "let":
		it := p.next()
		ph := PhraseLet{
			Pos: it.pos,
		}
		ph.Local = p.mustParseWordsUntil("be")
		p.dict.Add("local var", ph.Local)
		p.dict.Sort()
		p.mustParseWordOneOf("be")
		ph.Value = p.parseExpr1()
		return ph
	case "follow":
		it := p.next()
		_ = it
		p.parseExpr1()
		return nil // TODO
	case "try":
		it := p.next()
		_ = it
		_ = p.parseWordOneOf("silently")
		p.parseExpr1()
		return nil // TODO
	}

	return p.parsePhraseFuncCall()
}

func (p *Parser) parsePhraseFuncCall() Phrase {
	pos := p.peek().pos
	return PhraseFuncCall{
		Pos:  pos,
		Func: p.parseExpr1(),
	}
}
