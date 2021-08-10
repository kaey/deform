package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemSentenceEnd
	itemIndent
	itemWord
	itemQuotedStringStart
	itemQuotedStringEnd
	itemQuotedStringText
	itemQuotedStringAction
	itemComma
	itemColon
	itemSemicolon
	itemLeftParen
	itemRightParen
	itemComment
	itemSpace
	itemNL
	itemTab
)

type item struct {
	typ itemType
	val string
	pos Pos
}

type Pos struct {
	Name string
	Pos  int
	Line int
}

func (p Pos) String() string {
	if p.Line == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", p.Name, p.Line)
}

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input string
	items []item
	width int // number of bytes in last scanned rune.
	pos   Pos
	start Pos
}

func lex(name, input string) []item {
	l := &lexer{
		input: input,
		items: make([]item, 0, 5000),
		pos:   Pos{Name: name, Line: 1},
		start: Pos{Name: name, Line: 1},
	}
	for state := lexSentence; state != nil; {
		state = state(l)
	}

	return l.items
}

func (l *lexer) next() rune {
	if l.pos.Pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos.Pos:])
	l.width = w
	l.pos.Pos += l.width
	if r == '\n' {
		l.pos.Line++
	}
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos.Pos -= l.width
	if l.width == 1 && l.input[l.pos.Pos] == '\n' {
		l.pos.Line--
	}
}

func (l *lexer) emit(typ itemType) {
	it := item{typ, l.input[l.start.Pos:l.pos.Pos], l.start}
	l.items = append(l.items, it)
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	it := item{itemError, fmt.Sprintf(format, args...), l.start}
	l.items = append(l.items, it)
	return nil
}

func lexSpace(l *lexer) stateFn {
	for {
		if r := l.next(); r != ' ' {
			l.backup()
			l.emit(itemSpace)
			return lexSentence
		}
	}
}

func lexNewline(l *lexer) stateFn {
	if r := l.peek(); r != '\n' {
		l.emit(itemNL)
		return lexIndent
	}

	for {
		if r := l.next(); r != '\n' {
			l.backup()
			l.emit(itemSentenceEnd)
			return lexIndent
		}
	}
}

func lexTab(l *lexer) stateFn {
	for {
		if r := l.next(); r != '\t' {
			l.backup()
			l.emit(itemTab)
			return lexSentence
		}
	}
}

func lexIndent(l *lexer) stateFn {
	for {
		if r := l.next(); r == '\t' {
			continue
		} else if r == ' ' {
			return l.errorf("spaces are not allowed for indent")
		}
		l.backup()
		break
	}

	if l.pos.Pos > l.start.Pos {
		l.emit(itemIndent)
	}
	return lexSentence
}

func lexWord(l *lexer) stateFn {
	for {
		if r := l.next(); !(isWordChar(r) || ((r == ',' || r == '.') && isWordChar(l.peek()))) {
			break
		}
	}
	l.backup()
	l.emit(itemWord)
	return lexSentence
}

func lexQuotedString(l *lexer) stateFn {
	l.emit(itemQuotedStringStart)
	for {
		if r := l.next(); r == '"' {
			l.backup()
			if l.pos.Pos > l.start.Pos {
				l.emit(itemQuotedStringText)
			}
			l.next()
			break
		} else if r == ']' {
			return l.errorf("closing bracket ] without matching opening bracket [")
		} else if r == '[' {
			l.backup()
			if l.pos.Pos > l.start.Pos {
				l.emit(itemQuotedStringText)
			}
			l.next()
			for {
				if r := l.next(); r == '"' {
					return l.errorf("unterminated action in quoted string")
				} else if r == ']' {
					break
				} else if r == '[' {
					return l.errorf("nested action in quoted string")
				} else if r == eof {
					return l.errorf("unterminated action in quoted string")
				}
			}
			l.emit(itemQuotedStringAction)
		} else if r == eof {
			return l.errorf("unterminated quoted string")
		}
	}

	l.emit(itemQuotedStringEnd)
	return lexSentence
}

func lexComment(l *lexer) stateFn {
	nbrackets := 1
	for {
		if r := l.next(); r == '[' {
			nbrackets++
		} else if r == ']' {
			nbrackets--
			if nbrackets == 0 {
				break
			}
		} else if r == eof {
			return l.errorf("unterminated comment")
		}
	}
	l.emit(itemComment)
	return lexSentence
}

func lexSentence(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		l.emit(itemEOF)
		return nil
	case r == ' ':
		return lexSpace
	case r == '\n':
		return lexNewline
	case r == '\t':
		return lexTab
	case r == '"':
		return lexQuotedString
	case r == '.':
		l.emit(itemSentenceEnd)
		return lexSentence
	case r == ',':
		l.emit(itemComma)
		return lexSentence
	case r == ':':
		l.emit(itemColon)
		return lexSentence
	case r == ';':
		l.emit(itemSemicolon)
		return lexSentence
	case r == '(':
		l.emit(itemLeftParen)
		return lexSentence
	case r == ')':
		l.emit(itemRightParen)
		return lexSentence
	case r == '[':
		return lexComment
	}

	return lexWord
}

func isWordChar(r rune) bool {
	switch r {
	case '<', '>', '=', '-', '+', '/', '*', '_', '\'':
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return unicode.IsLetter(r)
}
