package parse

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kr/pretty"
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
			panic(p.err) // TODO: remove
			/*if r != errParse {
				panic(r)
			}*/
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
	case it.val == "To", it.val == "to":
		return p.parseFunc()
	case it.val == "For":
		return p.parseRule("For")
	case it.val == "Report":
		return p.parseRule("Report")
	case it.val == "Check", it.val == "check":
		return p.parseRule("Check")
	case it.val == "Carry":
		p.mustParseWordOneOf("out", "Out")
		return p.parseRule("Carry out")
	case it.val == "After":
		return p.parseRule("After")
	case it.val == "Before":
		return p.parseRule("Before")
	case it.val == "Every":
		p.mustParseWordOneOf("turn")
		return p.parseRule("Every turn")
	case it.val == "When":
		p.mustParseWordOneOf("play")
		p.mustParseWordOneOf("begins")
		return p.parseRule("When play begins")
	case it.val == "This":
		p.mustParseWordOneOf("is")
		return p.parseRule("This is")
	case it.val == "Table":
		return p.parseTable()
	case it.val == "Figure":
		p.mustParseWordOneOf("of")
		return p.parseFigure()
	case it.val == "Rule":
		p.mustParseWordOneOf("for")
		return p.parseRuleFor()
	case it.val == "Understand":
		return p.parseUnderstand()
	case it.val == "Does":
		p.mustParseWordOneOf("the")
		p.mustParseWordOneOf("player")
		p.mustParseWordOneOf("mean")
		return p.parseDoesThePlayerMean()
	}

	p.backup()
	return p.parseDecl()
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
	def.Object = p.mustParseWordsUntil("is", "are")
	def.Called = p.parseDefinitionCalled()
	p.mustParseWordOneOf("is", "are")
	def.Prop = p.mustParseWordsUntil("if", "when")
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
			part.ArgName = p.mustParseWordsUntil("-")
			p.mustParseWordOneOf("-")
			p.parseArticle()
			part.ArgKind = p.mustParseWordsUntil()
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
	r.Prefix = when
	r.Rulebook = p.parseWordsUntil("when") // TODO: When play begins should be assigned here instead of prefix maybe
	if p.parseWordOneOf("when") {
		r.rawCond = p.parseRawExpr()
	}
	if p.parseLeftParen() {
		p.mustParseWordOneOf("this")
		p.mustParseWordOneOf("is")
		p.parseArticle()
		r.Name = p.mustParseWordsUntil("rule") // TODO: except for warp portals rule
		p.mustParseWordOneOf("rule")
		p.mustParseRightParen()
	}
	p.mustParseColon()
	r.Comment = p.parseComment()
	r.rawPhrases = p.parseRawPhrases()
	p.mustParseSentenceEnd()
	return r
}

func (p *Parser) parseTable() Sentence {
	return Table(p.parseUnknownSentence())
}

func (p *Parser) parseFigure() Sentence {
	return Figure(p.parseUnknownSentence())
}

func (p *Parser) parseRuleFor() Sentence {
	return RuleFor(p.parseUnknownSentence())
}

func (p *Parser) parseUnderstand() Sentence {
	for {
		it := p.next()
		if it.typ == itemSentenceEnd {
			break
		}
	}

	return Understand(p.parseUnknownSentence())
}

func (p *Parser) parseDoesThePlayerMean() Sentence {
	for {
		it := p.next()
		if it.typ == itemSentenceEnd {
			break
		}
	}

	return DoesThePlayerMean(p.parseUnknownSentence())
}

func (p *Parser) parseDecl() Sentence {
	p.parseArticleCapital() // TODO
	if p.parseWordOneOf("File") {
		p.mustParseWordOneOf("of")
		return FileOf(p.parseUnknownSentence())
	}

	errs := make([]error, 0, 10)

	if s, err := p.parseDeclRule(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclListedInRulebook(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclVariable(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclKind(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclProp(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclAction(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclPropVal(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclPropEnum(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	pretty.Println("ERRS", errs)

	pretty.Println("FIXME:", p.parseUnknownSentence())
	return nil
}

func (p *Parser) declBackup(err *error) func() {
	inititi := p.iti
	return func() {
		if r := recover(); r != nil {
			if r != errParse {
				panic(r)
			}

			*err = p.err
			p.err = nil
			p.iti = inititi
		}
	}
}

func (p *Parser) parseDeclRule() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var r Rule
	r.Rulebook = p.mustParseWordsUntil("rule")
	p.mustParseWordOneOf("rule")
	if p.parseLeftParen() {
		p.mustParseWordOneOf("this")
		p.mustParseWordOneOf("is")
		p.parseArticle()
		r.Name = p.mustParseWordsUntil("rule")
		p.mustParseWordOneOf("rule")
		p.mustParseRightParen()
	}
	p.mustParseColon()
	p.parseComment()
	p.parseRawPhrases()
	p.mustParseSentenceEnd()
	return r, nil
}

func (p *Parser) parseDeclListedInRulebook() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var l ListedInRulebook
	l.Rule = p.mustParseWordsUntil("rule")
	p.mustParseWordOneOf("rule")
	p.mustParseWordOneOf("is")
	l.Listed = !p.parseWordOneOf("not")
	p.mustParseWordOneOf("listed")
	l.First = p.parseWordOneOf("first")
	l.Last = p.parseWordOneOf("last")
	p.mustParseWordOneOf("in")
	p.parseArticle()
	l.Rulebook = p.mustParseWordsUntil() // TODO: is not listed in any rulebook
	p.mustParseSentenceEnd()
	return l, nil
}

func (p *Parser) parseDeclVariable() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var v Variable
	v.Name = p.mustParseWordsUntil("is")
	p.mustParseWordOneOf("is")
	p.parseArticle()
	p.parseWordOneOf("indexed")            // just ignore this word
	v.Kind = p.mustParseWordsUntil("that") // TODO: list of clothing
	p.mustParseWordOneOf("that")
	p.mustParseWordOneOf("varies")
	p.mustParseSentenceEnd()
	return v, nil
}

func (p *Parser) parseDeclKind() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var k Kind
	k.Name = p.mustParseWordsUntil("is", "are")
	p.mustParseWordOneOf("is", "are")
	p.mustParseWordOneOf("a")
	p.mustParseWordOneOf("kind")
	p.mustParseWordOneOf("of")
	k.Kind = p.mustParseWordsUntil()
	p.mustParseSentenceEnd()
	return k, nil
}

func (p *Parser) parseDeclProp() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var pr Prop
	pr.Object = p.mustParseWordsUntil("has")
	p.mustParseWordOneOf("has")
	p.mustParseWordOneOf("a")
	pr.Kind = p.mustParseWordsUntil("called")
	p.mustParseWordOneOf("called")
	pr.Name = p.mustParseWordsUntil()
	p.mustParseSentenceEnd()
	return pr, nil
}

func (p *Parser) parseDeclAction() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var a Action
	a.Name = p.mustParseWordsUntil("is")
	p.mustParseWordOneOf("is")
	p.mustParseWordOneOf("an")
	p.mustParseWordOneOf("action")
	p.mustParseWordOneOf("applying")
	p.mustParseWordOneOf("to")
	if p.parseWordOneOf("nothing") {
		a.NThings = 0
	} else if p.parseWordOneOf("one") {
		a.NThings = 1
		a.Touchable = p.parseWordOneOf("touchable")
		p.mustParseWordOneOf("thing")
	} else if p.parseWordOneOf("two") {
		a.NThings = 2
		p.mustParseWordOneOf("things", "objects")
	}
	p.mustParseSentenceEnd()
	return a, nil
}

func (p *Parser) parseDeclPropVal() (s Sentence, err error) {
	defer p.declBackup(&err)()
	// TODO: pack of cards, chest of drawers, belt of sturdiness etc

	var v PropVal
	v.Prop = p.mustParseWordsUntil("of")
	p.mustParseWordOneOf("of")
	p.parseArticle()
	v.Object = p.mustParseWordsUntil("is")
	p.mustParseWordOneOf("is")
	v.Usually = p.parseWordOneOf("usually")
	p.parseArticle()
	if p.peek().typ == itemQuotedString {
		v.Val = p.next().val
	} else {
		v.Val = p.mustParseWordsUntil()
	}
	p.mustParseSentenceEnd()
	return v, nil
}

func (p *Parser) parseDeclPropEnum() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var e PropEnum
	e.Object = p.mustParseWordsUntil("can")
	p.mustParseWordOneOf("can")
	p.mustParseWordOneOf("be")
	for {
		v := p.mustParseWordsUntil("or")
		e.Vals = append(e.Vals, v)
		if !p.parseComma() && !p.parseWordOneOf("or") {
			break
		}
	}
	if p.parseLeftParen() {
		p.mustParseWordOneOf("this")
		p.mustParseWordOneOf("is")
		p.parseArticle()
		e.Name = p.mustParseWordsUntil("property")
		p.mustParseWordOneOf("property")
		p.mustParseRightParen()
	}
	p.mustParseSentenceEnd()
	return e, nil
}

func (p *Parser) parseUnknownSentence() string {
	s := new(strings.Builder)
	for {
		it := p.next()
		if it.typ == itemSentenceEnd {
			break
		}
		s.WriteString(" ")
		s.WriteString(it.val)
	}

	return s.String()
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
