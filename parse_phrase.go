package main

import (
	"strings"
)

func ParsePhrases(items []item) ([]Phrase, error) {
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
		if p.peek().typ == itemSentenceEnd {
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
	it := p.next()
	if it.typ == itemComment {
		p.backup()
		return p.parseComment()
	}

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
		if s, ok := p.parseExprQuotedString(); ok {
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
		ph.Value = p.mustParseWordsUntil()
		return ph
	case "while":
		ph := PhraseWhile{
			Pos: it.pos,
		}
		ph.Expr = p.parseExpr()
		p.mustParseColon()
		ph.Phrases = p.parsePhrases()
		return ph
	case "if":
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
	case "say":
		return p.parsePhraseSay()
	}
	// TODO: decide, decide on

	return p.parsePhraseFuncCall()
}

func (p *Parser) parsePhraseSay() PhraseSay {
	ph := PhraseSay{
		Pos: p.items[p.iti].pos,
	}

	if q, ok := p.parseQuotedString(); ok {
		ph.Say = q
		return ph
	}

	ph.Say = p.parsePhraseFuncCall()
	return ph
}

func (p *Parser) parsePhraseFuncCall() PhraseFuncCall {
	it := p.items[p.iti]
	ph := PhraseFuncCall{
		Pos: it.pos,
	}

	parts := []string{it.val}
	for {
		it = p.next()
		if it.typ != itemWord {
			p.backup()
			break
		}
		parts = append(parts, it.val)
	}

	ph.Func = strings.Join(parts, " ")
	return ph
}
