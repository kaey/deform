package main

import (
	"strconv"
	"strings"
)

func stripArticle(s string) string {
	ss := strings.Split(s, " ")
	if len(ss) < 2 {
		return s
	}
	if contains(ss[0], "A", "An", "The", "a", "an", "the") {
		return strings.Join(ss[1:], " ")
	}
	return s
}

func (p *Parser) parseArticle() string {
	it := p.next()
	if it.typ != itemWord || it.val != "A" && it.val != "An" && it.val != "The" && it.val != "a" && it.val != "an" && it.val != "the" {
		p.backup()
		return ""
	}

	return it.val
}

func (p *Parser) parseWord() (string, bool) {
	it := p.next()
	if it.typ != itemWord {
		p.backup()
		return "", false
	}

	return it.val, true
}

func (p *Parser) mustParseWord() string {
	it := p.next()
	if it.typ != itemWord {
		p.errorf("expected word, got %v", it.val)
	}

	return it.val
}

func (p *Parser) parseWordOneOf(v ...string) bool {
	if it := p.next(); it.typ != itemWord || !contains(it.val, v...) {
		p.backup()
		return false
	}

	return true
}

func (p *Parser) mustParseWordOneOf(v ...string) {
	if it := p.next(); it.typ != itemWord || !contains(it.val, v...) {
		p.errorf("expected %v, got %v", concat(v...), it.val)
	}
}

func (p *Parser) parseWordsUntil(v ...string) string {
	acc := make([]string, 0, 3)
	for {
		it := p.next()
		if it.typ != itemWord || contains(it.val, v...) {
			p.backup()
			break
		}

		acc = append(acc, it.val)
	}

	return strings.Join(acc, " ")
}

func (p *Parser) mustParseWordsUntil(v ...string) string {
	r := p.parseWordsUntil(v...)
	if r == "" {
		p.errorf("expected at least 1 word, got %v", p.items[p.iti+1].val)
	}

	return r
}

func (p *Parser) mustParseColon() {
	if it := p.next(); it.typ != itemColon {
		p.errorf("expected colon, got %q", it.val)
	}
}

func (p *Parser) parseSemicolon() bool {
	if it := p.next(); it.typ != itemSemicolon {
		p.backup()
		return false
	}

	return true
}

func (p *Parser) parseSentenceEnd() bool {
	if it := p.next(); it.typ != itemSentenceEnd {
		p.backup()
		return false
	}

	return true
}

func (p *Parser) mustParseSentenceEnd() {
	if it := p.next(); it.typ != itemSentenceEnd {
		p.errorf("expected sentence end, got %q", it.val)
	}
}

func (p *Parser) parseLeftParen() bool {
	if it := p.next(); it.typ != itemLeftParen {
		p.backup()
		return false
	}

	return true
}

func (p *Parser) mustParseRightParen() {
	if it := p.next(); it.typ != itemRightParen {
		p.errorf("expected right paren, got %q", it.val)
	}
}

func (p *Parser) parseComma() bool {
	if it := p.next(); it.typ != itemComma {
		p.backup()
		return false
	}

	return true
}

func (p *Parser) parseComment() Comment {
	s := make([]string, 0, 1)
	for {
		it := p.next()
		if it.typ != itemComment {
			p.backup()
			return Comment{
				Pos: it.pos,
				Str: strings.Join(s, ". "),
			}
		}

		s = append(s, it.val)
	}
}

func (p *Parser) parseIndent() int {
	it := p.next()
	if it.typ != itemIndent {
		p.backup()
		return 0
	}

	return len(it.val)
}

func (p *Parser) parseNumber() (int, bool) {
	it := p.next()
	if it.typ != itemWord {
		p.backup()
		return 0, false
	}

	n, err := strconv.Atoi(it.val)
	if err != nil {
		p.backup()
		return 0, false
	}

	return n, true
}

func (p *Parser) parseNL() bool {
	if it := p.next(); it.typ != itemNL {
		p.backup()
		return false
	}

	return true
}

func (p *Parser) mustParseNL() {
	if it := p.next(); it.typ != itemNL {
		p.errorf("expected newline, got %q", it.val)
	}
}

func (p *Parser) mustParseTab() {
	if it := p.next(); it.typ != itemTab {
		p.errorf("expected tab, got %q", it.val)
	}
}

func (p *Parser) parseQuotedString() (QuotedString, bool) {
	it := p.next()
	if it.typ != itemQuotedStringStart {
		p.backup()
		return QuotedString{}, false
	}

	q := QuotedString{
		Pos: it.pos,
	}

	for {
		if it := p.next(); it.typ == itemQuotedStringText {
			q.Parts = append(q.Parts, it.val)
		} else if it.typ == itemQuotedStringAction {
			q.Parts = append(q.Parts, RawFuncCall(strings.Trim(it.val, "[]")))
		} else if it.typ == itemQuotedStringEnd {
			break
		} else {
			p.errorf("bug: invalid item inside quoted string: %v", it)
		}
	}

	return q, true
}
