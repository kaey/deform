package parse

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Parser struct {
	name   string
	s      []Sentence
	items  []item
	iti    int // index of current item
	err    error
	indent int // current indent level when parsing raw phrases
}

func Parse(name string, input string) ([]Sentence, error) {
	p := &Parser{
		name:  name,
		s:     make([]Sentence, 0, 100),
		items: lex(name, input),
		iti:   -1, // calling next() will point to first available item
	}

	p.parse()
	return p.s, p.err
}

func ParseFile(path string) ([]Sentence, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	input := new(strings.Builder)
	if _, err := io.Copy(input, f); err != nil {
		return nil, err
	}

	return Parse(path, input.String())
}

func (p *Parser) next() item {
	p.iti++
	it := p.items[p.iti]
	if it.typ == itemError {
		p.errorf("%s", it.val)
	}
	return it
}

func (p *Parser) backup() {
	p.iti--
}

func (p *Parser) peek() item {
	it := p.next()
	p.backup()
	return it
}

var errParse = errors.New("parser error")

func (p *Parser) errorf(format string, args ...interface{}) {
	it := p.items[p.iti]
	p.err = fmt.Errorf("%s:%d: %w", p.name, it.pos.Line, fmt.Errorf(format, args...))
	panic(errParse)
}

func (p *Parser) parse() {
	defer func() {
		if r := recover(); r != nil {
			/*if r != errParse {
				panic(r)
			}*/
			panic(p.err)
		}
	}()

	for {
		it := p.next()
		if it.typ == itemEOF {
			return
		} else if it.typ == itemComment || it.typ == itemSentenceEnd || it.typ == itemIndent {
			continue
		}

		s := p.parseSentence()
		p.s = append(p.s, s)
	}
}

func (p *Parser) parseSentence() Sentence {
	it := p.items[p.iti]
	if it.typ != itemWord {
		if it.typ == itemQuotedString {
			return RoomDescr(it.val)
		}
		p.errorf("expected word, got %v", it.val)
	}
	switch {
	case contains(it.val, "Volume", "Book", "Part", "Section", "Chapter"):
		return p.parseSubheader()
	case it.val == "Definition":
		return p.parseDefinition()
	case it.val == "To":
		return p.parseFunc()
	case it.val == "Report", it.val == "Check":
		return p.parseRule(it.val)
	case it.val == "Carry":
		p.mustParseWordOneOf("out", "Out")
		return p.parseRule("Carry out")
	case it.val == "Every":
		p.mustParseWordOneOf("turn")
		return p.parseRule("Every turn")
	case it.val == "This":
		p.mustParseWordOneOf("is")
		return p.parseRule("This is")
	}
	return p.parseUnknownSentence()
}

func (p *Parser) parseSubheader() Sentence {
	p.backup()

	acc := make([]string, 0, 3)
	for {
		it := p.next()
		if it.typ != itemWord && it.typ != itemLeftParen && it.typ != itemRightParen && it.typ != itemComma {
			p.backup()
			break
		}

		acc = append(acc, it.val)
	}

	p.mustParseSentenceEnd()

	return Subheader(strings.Join(acc, " "))
}

func (p *Parser) parseDefinition() Sentence {
	p.mustParseColon()

	var def Definition
	p.parseArticle()
	def.Object = p.parseWordsUntil("is", "are")
	def.Called = p.parseDefinitionCalled()
	p.mustParseWordOneOf("is", "are")
	def.Prop = p.parseWordsUntil("if", "when")
	if p.parseWordOneOf("if") {
		def.Cond = p.parseExpr()
		return def
	}
	if p.parseWordOneOf("when") {
		// TODO: differ between if and when
		def.Cond = p.parseExpr()
	}

	p.mustParseColon()
	p.parseComment()
	def.rawPhrases = p.parseRawPhrases()
	p.mustParseSentenceEnd()
	return def
}

func (p *Parser) parseDefinitionCalled() string {
	if !p.parseLeftParen() {
		return ""
	}

	p.mustParseWordOneOf("called")
	called := p.mustParseWord()
	p.mustParseRightParen()

	return called
}

func (p *Parser) parseFunc() Sentence {
	// TODO: decide if, decide whether, say
	var fn Func
	fn.Parts = p.parseFuncParts()
	p.mustParseColon()
	fn.Comment = p.parseComment()
	fn.rawPhrases = p.parseRawPhrases()
	p.mustParseSentenceEnd()
	return fn
}

func (p *Parser) parseFuncParts() []FuncPart {
	parts := make([]FuncPart, 0, 5)
	for {
		if w, ok := p.parseWord(); ok {
			parts = append(parts, FuncPart{Word: w})
			continue
		}

		if p.parseLeftParen() {
			part := FuncPart{}
			part.ArgName = p.parseWordsUntil("-")
			p.mustParseWordOneOf("-")
			p.parseArticle()
			part.ArgKind = p.parseWordsUntil()
			p.mustParseRightParen()
			parts = append(parts, part)
			continue
		}

		break
	}

	return parts
}

func (p *Parser) parseRule(when string) Sentence {
	var r Rule
	r.When = when
	r.Action = p.parseWordsUntil("when")
	if p.parseWordOneOf("when") {
		r.rawCond = p.parseRawExpr()
	}
	if p.parseLeftParen() {
		p.mustParseWordOneOf("this")
		p.mustParseWordOneOf("is")
		p.parseArticle()
		r.Name = p.parseWordsUntil("rule") // TODO: except for warp portals rule
		p.mustParseWordOneOf("rule")
		p.mustParseRightParen()
	}
	p.mustParseColon()
	r.Comment = p.parseComment()
	r.rawPhrases = p.parseRawPhrases()
	p.mustParseSentenceEnd()
	return r
}

func (p *Parser) parseUnknownSentence() Sentence {
	p.backup()
	it := p.next()
	if it.typ != itemWord {
		p.errorf("parseUnknownSentence: expected word, got %v", it.val)
	}
	str := new(strings.Builder)
	str.WriteString(it.val)
	for {
		it := p.next()
		if it.typ == itemSentenceEnd {
			break
		}

		str.WriteByte(' ')
		str.WriteString(it.val)
	}

	return UnknownSentence(str.String())
}

func (p *Parser) parseRawPhrases() []rawPhrase {
	if it := p.peek(); it.typ == itemWord {
		return []rawPhrase{p.parseRawPhrase()}
	}

	p.indent++
	phs := make([]rawPhrase, 0, 3)
	for {
		if it := p.peek(); it.typ == itemSentenceEnd {
			break
		}

		in := p.parseIndent()
		if in < p.indent {
			p.backup()
			break
		} else if in > p.indent {
			p.errorf("expected indent %v, got %v", p.indent, in)
		}
		ph := p.parseRawPhrase()
		p.parseComment() /* TODO: assign comment to phrase */
		phs = append(phs, ph)
	}

	if len(phs) == 0 {
		p.errorf("expected at least 1 phrase")
	}

	p.indent--
	return phs
}

func (p *Parser) parseRawPhrase() rawPhrase {
	var ph rawPhrase
	it := p.next()
	ph.Pos = it.pos
	if it.typ == itemComment {
		ph.Comment = Comment(it.val)
		return ph
	}

	start := p.iti
	for {
		switch it.typ {
		case itemSemicolon:
			ph.items = p.items[start:p.iti] /* TODO: fix index? */
			return ph
		case itemColon:
			ph.items = p.items[start:p.iti]
			ph.Comment = p.parseComment()
			ph.children = p.parseRawPhrases()
			return ph
		case itemComma:
			ph.items = p.items[start:p.iti]
			ph.children = []rawPhrase{p.parseRawPhrase()}
			return ph
		case itemSentenceEnd:
			p.backup()
			ph.items = p.items[start:p.iti]
			return ph
		}
		it = p.next()
	}
}

func (p *Parser) parseRawExpr() rawExpr {
	var ex rawExpr
	it := p.next()
	ex.Pos = it.pos

	// TODO: avoid capturing rule name (check for left-paren, this, is)
	start := p.iti
	for {
		switch it.typ {
		case itemWord, itemLeftParen, itemRightParen:
		case itemColon:
			p.backup()
			ex.items = p.items[start:p.iti]
			return ex
		default:
		}
		it = p.next()
	}
}

/*
func (p *Parser) parsePhrase() Phrase {
	switch it.val {
	case "say":
		it := p.next()
		if it.typ != itemQuotedString {
			p.errorf("say expects quoted string, got %v", it.val)
		}
		return PhraseSay(it.val)
	// case "decide", "let", "do", "now":
	case "if":
		_ = p.parseExpr()
		if p.parseComma() {
			_ = p.parsePhrase()
			return nil
		}
		p.mustParseColon()
		_ = p.parseRawPhrases()
	case "otherwise":
		if it := p.peek(); it.typ == itemWord && it.val == "if" {
			p.next()
			_ = p.parseExpr()
		}
		p.mustParseColon()
		_ = p.parseRawPhrases()
	case "unless":
		_ = p.parseExpr()
		p.mustParseColon()
		_ = p.parseRawPhrases()
	case "repeat":
		_ = p.parseExpr()
		p.mustParseColon()
		_ = p.parseRawPhrases()
	case "while":
		_ = p.parseExpr()
		p.mustParseColon()
		_ = p.parseRawPhrases()
	}

	return nil
}
*/
