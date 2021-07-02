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
	for {
		p.iti++
		it := p.items[p.iti]
		if it.typ == itemError {
			p.errorf("%s", it.val)
		} else if it.typ == itemSpace {
			continue
		}
		return it
	}
}

func (p *Parser) backup() {
	for {
		p.iti--
		if p.iti == -1 || p.items[p.iti].typ != itemSpace {
			break
		}
	}
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
			panic(p.err)
			/*if r != errParse {
				panic(r)
			}*/
		}
	}()

	for {
		it := p.next()
		if it.typ == itemEOF {
			return
		} else if it.typ == itemComment || it.typ == itemSentenceEnd || it.typ == itemIndent || it.typ == itemNL {
			// TODO: don't ignore comments
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
		p.mustParseWordOneOf("of")
		return p.parseTable()
	case it.val == "Figure":
		p.mustParseWordOneOf("of")
		return p.parseFigure()
	case it.val == "Rule":
		p.mustParseWordOneOf("for")
		return p.parseRuleFor()
	case it.val == "Understand", it.val == "understand":
		return p.parseUnderstand()
	case it.val == "Does":
		p.mustParseWordOneOf("the")
		p.mustParseWordOneOf("player")
		p.mustParseWordOneOf("mean")
		return p.parseDoesThePlayerMean()
	case it.val == "There":
		p.mustParseWordOneOf("is", "are")
		return p.parseThereAre()
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
	def.Called = p.parseCalled()
	p.mustParseWordOneOf("is", "are")
	def.Prop = p.mustParseWordsUntil("if", "when")
	if p.parseWordOneOf("if") {
		def.rawCond = p.parseRawExpr()
		return def
	}
	if p.parseWordOneOf("when") {
		// TODO: differ between if and when
		def.rawCond = p.parseRawExpr()
	}

	p.mustParseColon()
	p.parseComment()
	def.rawPhrases = p.parseRawPhrases()
	p.mustParseSentenceEnd()
	return def
}

func (p *Parser) parseCalled() string {
	if !p.parseLeftParen() {
		return ""
	}

	p.mustParseWordOneOf("called")
	p.parseArticle()
	called := p.mustParseWordsUntil()
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
	var t Table
	t.Name = p.mustParseWordsUntil()
	if p.parseLeftParen() {
		p.mustParseWordOneOf("continued")
		p.mustParseRightParen()
		t.Continued = true
	}
	p.mustParseNL()

	for {
		h := p.mustParseWordsUntil()
		t.ColNames = append(t.ColNames, h)
		if p.parseLeftParen() {
			k := p.mustParseWordsUntil()
			t.ColKinds = append(t.ColKinds, k)
			p.mustParseRightParen()
		} else {
			t.ColKinds = append(t.ColKinds, "")
		}
		if p.parseNL() {
			break
		}
		p.mustParseTab()
	}
	for {
		if p.parseWordOneOf("with") {
			p.parseNumber()
			p.mustParseWordOneOf("blank")
			p.mustParseWordOneOf("rows")
			p.mustParseSentenceEnd()
			return t
		}
		if p.parseComment() != "" {
			p.mustParseSentenceEnd()
			return t
		}
		row := make([]string, 0, len(t.ColNames))
		for i := 0; i < len(t.ColNames); i++ {
			var v string
			if p.peek().typ == itemQuotedString {
				v = p.next().val
			} else {
				v = p.mustParseWordsUntil()
			}
			row = append(row, v)
			if i < len(t.ColNames)-1 {
				p.mustParseTab()
			}
		}
		t.Rows = append(t.Rows, row)
		t.RowComments = append(t.RowComments, p.parseComment())
		if p.parseSentenceEnd() {
			return t
		}
		p.mustParseNL()
	}
}

func (p *Parser) parseFigure() Sentence {
	var f Figure
	f.Name = p.mustParseWordsUntil("is")
	p.mustParseWordOneOf("is")
	p.mustParseWordOneOf("the")
	p.mustParseWordOneOf("file")
	it := p.next()
	if it.typ != itemQuotedString {
		p.errorf("expected path to file in quotes, got %v", it)
	}
	f.FilePath = it.val
	p.mustParseSentenceEnd()
	return f
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

func (p *Parser) parseThereAre() Sentence {
	var t ThereAre
	n, ok := p.parseNumber()
	if !ok {
		p.parseArticle()
		n = 1
	}
	t.N = n
	t.Kind = p.mustParseWordsUntil()
	p.mustParseSentenceEnd()
	return t
}

func (p *Parser) parseVerb() Sentence {
	var v Verb
	v.Name = p.mustParseWordsUntil()
	if p.parseLeftParen() {
		for {
			p.mustParseWord() // just skip first word
			v.Alts = append(v.Alts, p.mustParseWordsUntil())
			if p.parseComma() {
				continue
			}
			break
		}
		p.mustParseRightParen()
	}
	p.mustParseWordOneOf("implies")
	p.parseArticle()
	v.Rel = p.mustParseWordsUntil("relation")
	p.mustParseWordOneOf("relation")
	p.mustParseSentenceEnd()
	return v
}

func (p *Parser) parseDecl() Sentence {
	p.parseArticleCapital()
	if p.parseWordOneOf("File") {
		p.mustParseWordOneOf("of")
		return FileOf(p.parseUnknownSentence())
	}

	if p.parseWordOneOf("verb") {
		p.mustParseWordOneOf("to")
		return p.parseVerb()
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

	if s, err := p.parseDeclRelation(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclVector(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclIsIn(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	if s, err := p.parseDeclIs(); err != nil {
		errs = append(errs, err)
	} else {
		return s
	}

	panic(errs)
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
	l.Rulebook = p.mustParseWordsUntil()
	p.mustParseSentenceEnd()
	return l, nil
}

func (p *Parser) parseDeclVariable() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var v Variable
	v.Name = p.mustParseWordsUntil("is")
	p.mustParseWordOneOf("is")
	p.parseArticle()
	p.parseWordOneOf("indexed") // just ignore this word
	if p.parseWordOneOf("list") {
		p.mustParseWordOneOf("of")
		v.Array = true
	}
	v.Kind = p.mustParseWordsUntil("that")
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
	pr.Object = p.mustParseWordsUntil("has", "have")
	p.mustParseWordOneOf("has", "have")
	p.parseArticle()
	pr.Kind = p.mustParseWordsUntil("called")
	if p.parseWordOneOf("called") {
		pr.Name = p.mustParseWordsUntil()
	}
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
		p.parseWordOneOf("or") // some props use comma followed by "or".
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

func (p *Parser) parseDeclRelation() (s Sentence, err error) {
	defer p.declBackup(&err)()

	num := func() int {
		switch p.items[p.iti].val {
		case "a", "one":
			return 1
		case "two":
			return 2
		case "various":
			return 3
		}
		panic("bug: unknown relation number")
	}

	var r Relation
	r.Name = p.mustParseWordsUntil("relates")
	p.mustParseWordOneOf("relates")
	p.mustParseWordOneOf("a", "one", "two", "various")
	r.NObjects = num()
	r.Kind = p.mustParseWordsUntil("to")
	r.Object = p.parseCalled()
	p.mustParseWordOneOf("to")
	p.mustParseWordOneOf("a", "one", "two", "various")
	r.NObjects2 = num()
	r.Kind2 = p.mustParseWordsUntil("when")
	r.Object2 = p.parseCalled()
	if p.parseWordOneOf("when") {
		r.rawCond = p.parseRawExpr()
	}
	p.mustParseSentenceEnd()
	return r, nil
}

func (p *Parser) parseDeclVector() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var v Vector
	v.Pattern = p.mustParseWord()
	p.mustParseWordOneOf("specifies")
	p.parseArticle()
	v.Kind = p.mustParseWordsUntil("with")
	p.mustParseWordOneOf("with")
	p.mustParseWordOneOf("parts")
	for {
		part := p.mustParseWordsUntil()
		v.Parts = append(v.Parts, part)
		if p.parseLeftParen() {
			p.mustParseWordsUntil() // without leading zeros
			p.mustParseRightParen()
		}
		if !p.parseComma() {
			break
		}
	}
	p.mustParseSentenceEnd()
	return v, nil
}

func (p *Parser) parseDeclIsIn() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var is IsIn
	n, ok := p.parseNumber()
	if ok {
		o := p.mustParseWordsUntil("is", "are")
		for i := 0; i < n; i++ {
			is.Objects = append(is.Objects, o)
		}
	} else {
		for {
			o := p.mustParseWordsUntil("is", "are")
			is.Objects = append(is.Objects, o)
			if !p.parseComma() {
				break
			}
		}
	}

	p.mustParseWordOneOf("is", "are")
	p.mustParseWordOneOf("in")
	p.parseArticle()
	is.Where = p.mustParseWordsUntil()
	p.mustParseSentenceEnd()
	return is, nil
}

func (p *Parser) parseDeclIs() (s Sentence, err error) {
	defer p.declBackup(&err)()

	var is Is
	is.Object = p.mustParseWordsUntil("is", "are")
	p.mustParseWordOneOf("is", "are")
	if p.peek().typ == itemQuotedString {
		is.Value = p.next().val
		p.mustParseSentenceEnd()
		return is, nil
	}
	is.Initially = p.parseWordOneOf("initially")
	is.Usually = p.parseWordOneOf("usually")
	if p.parseWordOneOf("east", "west", "south", "north", "above", "below") {
		is.Direction = p.items[p.iti].val
		if is.Direction != "above" && is.Direction != "below" {
			p.mustParseWordOneOf("of")
		}
	}
	is.Negate = p.parseWordOneOf("not")
	p.parseArticle()
	is.Value = p.mustParseWordsUntil("and")
	if p.parseComma() || p.parseWordOneOf("and") {
		is.EnumVal = append(is.EnumVal, is.Value)
		is.Value = ""
		for {
			v := p.mustParseWordsUntil("and")
			is.EnumVal = append(is.EnumVal, v)
			if !p.parseComma() && !p.parseWordOneOf("and") {
				break
			}
			p.parseWordOneOf("and")
		}
	}
	p.mustParseSentenceEnd()
	return is, nil
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

	p.parseNL()
	p.indent++
	phs := make([]rawPhrase, 0, 3)
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
		ph := p.parseRawPhrase()
		p.parseComment() /* TODO: assign comment to phrase */
		p.parseNL()
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
		case itemColon, itemSentenceEnd:
			p.backup()
			ex.items = p.items[start:p.iti]
			return ex
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
