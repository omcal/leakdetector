package parser

import (
	"strings"
	"unicode"
)

var keywords = map[string]bool{
	"class": true, "struct": true, "public": true, "private": true, "protected": true,
	"new": true, "delete": true, "virtual": true, "const": true, "static": true,
	"void": true, "int": true, "char": true, "float": true, "double": true,
	"bool": true, "long": true, "short": true, "unsigned": true, "signed": true,
	"if": true, "else": true, "for": true, "while": true, "do": true,
	"return": true, "nullptr": true, "NULL": true, "this": true,
	"template": true, "typename": true, "namespace": true, "using": true,
}

// Lexer tokenizes C++ source code
type Lexer struct {
	input  string
	pos    int
	line   int
	column int
	tokens []Token
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

// Tokenize processes the entire input and returns all tokens
func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		l.skipWhitespaceAndComments()
		if l.pos >= len(l.input) {
			break
		}

		ch := l.input[l.pos]

		// Check for :: scope operator before treating : as punctuation
		if ch == ':' && l.peek() == ':' {
			l.addToken(TokenOperator, "::")
			l.advance()
			l.advance()
			continue
		}

		switch {
		case ch == '"' || ch == '\'':
			l.readString(ch)
		case ch == '#':
			l.skipPreprocessor()
		case unicode.IsLetter(rune(ch)) || ch == '_':
			l.readIdentifier()
		case unicode.IsDigit(rune(ch)):
			l.readNumber()
		case l.isOperator(ch):
			l.readOperator()
		case l.isPunctuation(ch):
			l.addToken(TokenPunctuation, string(ch))
			l.advance()
		default:
			l.advance()
		}
	}

	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line, Column: l.column})
	return l.tokens
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.pos++
	}
}

func (l *Lexer) peek() byte {
	if l.pos+1 < len(l.input) {
		return l.input[l.pos+1]
	}
	return 0
}

func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advance()
		} else if ch == '/' && l.peek() == '/' {
			// Single-line comment
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.advance()
			}
		} else if ch == '/' && l.peek() == '*' {
			// Multi-line comment
			l.advance() // skip /
			l.advance() // skip *
			for l.pos < len(l.input)-1 {
				if l.input[l.pos] == '*' && l.peek() == '/' {
					l.advance() // skip *
					l.advance() // skip /
					break
				}
				l.advance()
			}
		} else {
			break
		}
	}
}

func (l *Lexer) skipPreprocessor() {
	// Skip preprocessor directives (lines starting with #)
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		// Handle line continuation
		if l.input[l.pos] == '\\' && l.peek() == '\n' {
			l.advance()
			l.advance()
			continue
		}
		l.advance()
	}
}

func (l *Lexer) readString(quote byte) {
	startLine := l.line
	startCol := l.column
	var sb strings.Builder
	sb.WriteByte(quote)
	l.advance() // skip opening quote

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\\' && l.pos+1 < len(l.input) {
			sb.WriteByte(ch)
			l.advance()
			if l.pos < len(l.input) {
				sb.WriteByte(l.input[l.pos])
				l.advance()
			}
		} else if ch == quote {
			sb.WriteByte(ch)
			l.advance()
			break
		} else if ch == '\n' {
			break // Unterminated string
		} else {
			sb.WriteByte(ch)
			l.advance()
		}
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenString,
		Value:  sb.String(),
		Line:   startLine,
		Column: startCol,
	})
}

func (l *Lexer) readIdentifier() {
	startLine := l.line
	startCol := l.column
	start := l.pos

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}

	value := l.input[start:l.pos]
	tokenType := TokenIdent
	if keywords[value] {
		tokenType = TokenKeyword
	}

	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   startLine,
		Column: startCol,
	})
}

func (l *Lexer) readNumber() {
	startLine := l.line
	startCol := l.column
	start := l.pos

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsDigit(rune(ch)) || ch == '.' || ch == 'x' || ch == 'X' ||
			(ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			l.advance()
		} else {
			break
		}
	}

	l.tokens = append(l.tokens, Token{
		Type:   TokenNumber,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	})
}

func (l *Lexer) isOperator(ch byte) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '=' ||
		ch == '<' || ch == '>' || ch == '!' || ch == '&' || ch == '|' ||
		ch == '^' || ch == '%' || ch == '~'
}

func (l *Lexer) readOperator() {
	startLine := l.line
	startCol := l.column
	start := l.pos

	// Handle multi-character operators
	if l.pos+1 < len(l.input) {
		two := l.input[l.pos : l.pos+2]
		if two == "::" || two == "->" || two == "==" || two == "!=" ||
			two == "<=" || two == ">=" || two == "&&" || two == "||" ||
			two == "++" || two == "--" || two == "+=" || two == "-=" ||
			two == "*=" || two == "/=" {
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, Token{
				Type:   TokenOperator,
				Value:  two,
				Line:   startLine,
				Column: startCol,
			})
			return
		}
	}

	l.advance()
	l.tokens = append(l.tokens, Token{
		Type:   TokenOperator,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	})
}

func (l *Lexer) isPunctuation(ch byte) bool {
	return ch == '{' || ch == '}' || ch == '(' || ch == ')' ||
		ch == '[' || ch == ']' || ch == ';' || ch == ',' ||
		ch == ':' || ch == '.'
}

func (l *Lexer) addToken(tokenType TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: l.column,
	})
}
