package main

func ParsePhrases(items []item, dict *Dict) ([]Phrase, error) {
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
			panic(p.err)
		}
	}()

	phs := p.parsePhrases()
	return phs, p.err
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
		p.errorf("expected at least 1 phrase")
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
		ph.Expr = p.parseExpr()
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
			ph.Expr = p.parseExpr()
		}
		if p.peek().typ == itemWord {
			ph.Phrases = []Phrase{p.parsePhrase()}
			return ph
		}
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	}

	return p.parsePhraseFuncCall()
	/*
		switch it.val {
		case "now":
			ph := PhraseNow{
				Pos: it.pos,
			}
			p.parseArticle()
			ph.Object = p.mustParseWordsUntil("is")
			p.mustParseWordOneOf("is")
			ph.Expr = p.parseExpr()
			return ph
		case "add":
			ph := PhraseListAdd{
				Pos: it.pos,
			}
			if s, ok := p.parseQuotedString(); ok {
				ph.Value = s
			} else {
				ph.Value = Ident(p.mustParseWordsUntil("to"))
			}

			p.mustParseWordOneOf("to")
			p.parseArticle()
			ph.List = p.mustParseWordsUntil()
			return ph
		case "let":
			ph := PhraseLet{
				Pos: it.pos,
			}
			ph.Object = p.mustParseWordsUntil("be")
			p.mustParseWordOneOf("be")
			ph.Value = p.parseExpr()
			return ph

		case "say":
			return p.parsePhraseSay()
		}
		// TODO: decide, decide on

		return p.parsePhraseFuncCall()*/
}

/*func (p *Parser) parsePhraseSay() PhraseSay {
	ph := PhraseSay{
		Pos: p.items[p.iti].pos,
	}

	if q, ok := p.parseQuotedString(); ok {
		ph.Say = q
		return ph
	}

	ph.Say = p.parsePhraseFuncCall()
	return ph
}*/

func (p *Parser) parsePhraseFuncCall() Phrase {
	pos := p.peek().pos
	return PhraseFuncCall{
		Pos:  pos,
		Func: p.parseExpr(),
	}
}
