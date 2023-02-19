package assembler

import (
	"bufio"
	"io"
	"os"
	"strings"
	"unicode"
)

// Token Type
type Type uint8

const (
	DIRECTIVE Type = iota
	IDENTIFIER
	INSTRUCTION
	LABEL
	NUMBER
	STRING
)

// Test if char is a legal character for defining identifiers, directives and labels
func isIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$' || ch == '.'
}

// Test if  char a white space
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// Test if char octal/decimal/hex digit
func isDigit(ch rune, base int) bool {
	switch base {
	case 8: // octal
		return '0' <= ch && ch <= '7'
	case 10: // decimal
		return '0' <= ch && ch <= '9'
	case 16: // hex
		return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
	default:
		return false
	}
}

// Token
type Token struct {
	Type   Type
	Symbol string
	Line   int
}

// Return a new token
func newToken(tokenType Type, symbol string, line int) *Token {
	return &Token{Type: tokenType, Symbol: symbol, Line: line}
}

type Lexer struct {
	reader *bufio.Reader
	ch     rune
	line   int
}

// Return a new lexer
func NewLexer(file *os.File) *Lexer {
	lexer := new(Lexer)
	lexer.reader = bufio.NewReader(file)
	lexer.readRune()
	return lexer
}

// Reads a single UTF-8 character
func (l *Lexer) readRune() error {
	var err error
	l.ch, _, err = l.reader.ReadRune()
	if err != nil {
		l.ch = rune(0)
		return err
	}
	if l.ch == '\n' {
		l.line++
	}
	return nil
}

// Return the next token
func (l *Lexer) NextToken() (*Token, error) {
	var err error

	// Skip whitespaces
	if err = l.skipWhitespace(); err != nil {
		return nil, err
	}

	for l.ch == rune(';') || l.ch == rune('#') {
		// Skip comments
		if err = l.skipComment(); err != nil {
			return nil, err
		}
		// Skip whitespaces
		if err = l.skipWhitespace(); err != nil {
			return nil, err
		}
	}

	switch {
	case l.ch == rune(0): // EOF
		return nil, io.EOF

	case l.ch == '"': // String
		return l.readString()

	case l.ch == '.': // Directive
		return l.readDirective()

	case ('0' <= l.ch && l.ch <= '9') || l.ch == '-': // Number
		return l.readNumber()

	case isIdentifierChar(l.ch): // Identifier/Label/Instruction
		return l.readIdentifier()

	default:
		return nil, &UnexpectedToken{Token: string(l.ch), Line: l.line}
	}
}

// Skip whitespaces
func (l *Lexer) skipWhitespace() error {
	for isWhitespace(l.ch) {
		if err := l.readRune(); err != nil {
			return err
		}
	}
	return nil
}

// Skip comments
func (l *Lexer) skipComment() error {
	for l.ch != '\n' && l.ch != rune(0) {
		if err := l.readRune(); err != nil {
			return err
		}
	}
	return nil
}

// Read a directive
func (l *Lexer) readDirective() (*Token, error) {
	var buf strings.Builder
	line := l.line
	buf.WriteRune(l.ch)
	if err := l.readRune(); err != nil {
		return nil, err
	}
	for isIdentifierChar(l.ch) {
		buf.WriteRune(l.ch)
		if err := l.readRune(); err != nil {
			return nil, err
		}
	}
	return newToken(DIRECTIVE, strings.ToUpper(buf.String()), line), nil
}

// Read a number
func (l *Lexer) readNumber() (*Token, error) {
	var buf strings.Builder
	line := l.line
	base := 10 // default base (decimal)
	for i := 0; true; i++ {
		if i == 0 && l.ch == '-' { // negative numbers
			buf.WriteRune(l.ch)
		} else if i == 1 && l.ch == 'x' { // 0x
			base = 16 // hex
			buf.WriteRune(l.ch)
		} else if i == 1 && l.ch == 'o' { // 0o
			base = 8 // octal
			buf.WriteRune(l.ch)
		} else if isDigit(l.ch, base) {
			buf.WriteRune(l.ch)
		} else {
			break
		}
		if err := l.readRune(); err != nil {
			return nil, err
		}
	}
	return newToken(NUMBER, buf.String(), line), nil
}

// Read a quoted string
func (l *Lexer) readString() (*Token, error) {
	var buf strings.Builder
	line := l.line
	for {
		if err := l.readRune(); err != nil {
			return nil, err
		}
		if l.ch == rune(0) {
			return nil, UnmatechedDelimiter
		}
		if l.ch == '"' {
			break
		}
		if l.ch == '\\' { // escape
			if err := l.readRune(); err != nil {
				return nil, err
			}
			switch l.ch {
			case '\n':
				continue // escape new line
			case rune(0):
				return nil, UnmatechedDelimiter
			case '0':
				l.ch = rune(0)
			case 'n':
				l.ch = '\n'
			case 'r':
				l.ch = '\r'
			case 't':
				l.ch = '\t'
			}
		}
		buf.WriteRune(l.ch)
	}
	if err := l.readRune(); err != nil {
		return nil, err
	}
	return newToken(STRING, buf.String(), line), nil
}

// Read and Identifier/Label/Instruction
func (l *Lexer) readIdentifier() (*Token, error) {
	var buf strings.Builder
	line := l.line
	for isIdentifierChar(l.ch) {
		buf.WriteRune(l.ch)
		if err := l.readRune(); err != nil {
			return nil, err
		}
	}
	token := newToken(IDENTIFIER, strings.ToUpper(buf.String()), line)
	// Check if the symbol is a label
	if l.ch == ':' {
		token.Type = LABEL
		// Consume the current char (':')
		if err := l.readRune(); err != nil {
			return nil, err
		}
		return token, nil
	}
	// Check if the symbol is an instruction
	_, isInstruction := Instructions[token.Symbol]
	if isInstruction {
		token.Type = INSTRUCTION
	}
	return token, nil
}
