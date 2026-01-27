package parser

import (
	"os"
	"path/filepath"
	"strings"
)

// Parser parses C++ source files and extracts class information
type Parser struct {
	tokens  []Token
	pos     int
	file    string
	classes []Class
}

// ParseFile parses a single C++ file
func ParseFile(filename string) ([]Class, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	absPath, _ := filepath.Abs(filename)
	lexer := NewLexer(string(content))
	tokens := lexer.Tokenize()

	parser := &Parser{
		tokens: tokens,
		pos:    0,
		file:   absPath,
	}

	return parser.parse(), nil
}

func (p *Parser) parse() []Class {
	// First pass: parse inline class definitions
	for !p.isAtEnd() {
		if p.matchKeyword("class") || p.matchKeyword("struct") {
			if class := p.parseClass(); class != nil {
				p.classes = append(p.classes, *class)
			}
		} else if p.isOutOfClassMethod() {
			// Parse out-of-class method definitions (ClassName::MethodName)
			p.parseOutOfClassMethod()
		} else {
			p.advance()
		}
	}
	return p.classes
}

// isOutOfClassMethod checks for pattern: Type ClassName::MethodName(
func (p *Parser) isOutOfClassMethod() bool {
	// Look for :: operator followed by ( within reasonable distance
	for i := 0; i < 10 && p.pos+i < len(p.tokens); i++ {
		tok := p.tokens[p.pos+i]
		if tok.Value == "::" {
			// Found ::, look for ( after it
			for j := i + 1; j < i+5 && p.pos+j < len(p.tokens); j++ {
				if p.tokens[p.pos+j].Value == "(" {
					return true
				}
				if p.tokens[p.pos+j].Value == ";" {
					return false
				}
			}
		}
		if tok.Value == ";" || tok.Value == "{" || tok.Value == "}" {
			return false
		}
	}
	return false
}

// parseOutOfClassMethod parses ClassName::MethodName() { ... } definitions
func (p *Parser) parseOutOfClassMethod() {
	startLine := p.current().Line

	// Collect tokens until we find ::
	var className string
	for !p.isAtEnd() && !p.checkValue("::") {
		if p.check(TokenIdent) {
			className = p.current().Value // Last ident before :: is class name
		}
		p.advance()
	}

	if className == "" || !p.matchValue("::") {
		return
	}

	// Check for destructor (~)
	isDestructor := p.checkValue("~")
	if isDestructor {
		p.advance()
	}

	// Get method name
	if !p.check(TokenIdent) {
		return
	}
	methodName := p.current().Value
	p.advance()

	// Skip parameters
	if !p.matchValue("(") {
		return
	}
	parenCount := 1
	for !p.isAtEnd() && parenCount > 0 {
		if p.checkValue("(") {
			parenCount++
		} else if p.checkValue(")") {
			parenCount--
		}
		p.advance()
	}

	// Skip initializer list for constructors
	if p.checkValue(":") && !isDestructor {
		p.advance()
		for !p.isAtEnd() && !p.checkValue("{") && !p.checkValue(";") {
			p.advance()
		}
	}

	// Parse body
	if !p.checkValue("{") {
		return
	}

	fn := &Function{
		Name:         methodName,
		IsDestructor: isDestructor,
		StartLine:    startLine,
	}

	p.parseFunctionBody(fn)

	// Find or create class to attach this method to
	var targetClass *Class
	for i := range p.classes {
		if p.classes[i].Name == className {
			targetClass = &p.classes[i]
			break
		}
	}

	if targetClass == nil {
		// Create a placeholder class for this method
		newClass := Class{
			Name:    className,
			File:    p.file,
			Methods: []Function{},
		}
		p.classes = append(p.classes, newClass)
		targetClass = &p.classes[len(p.classes)-1]
	}

	// Attach method to class
	if isDestructor {
		targetClass.Destructor = fn
	} else if methodName == className {
		targetClass.Constructor = fn
	} else {
		targetClass.Methods = append(targetClass.Methods, *fn)
	}
}

func (p *Parser) parseClass() *Class {
	// Get class name
	if !p.check(TokenIdent) {
		return nil
	}

	className := p.current().Value
	startLine := p.current().Line
	p.advance()

	// Skip inheritance declaration
	for !p.isAtEnd() && !p.checkValue("{") && !p.checkValue(";") {
		p.advance()
	}

	// Forward declaration (ends with ;)
	if p.checkValue(";") {
		return nil
	}

	if !p.matchValue("{") {
		return nil
	}

	class := &Class{
		Name:      className,
		File:      p.file,
		StartLine: startLine,
		Members:   []Member{},
		Methods:   []Function{},
	}

	// Parse class body
	braceCount := 1
	for !p.isAtEnd() && braceCount > 0 {
		if p.checkValue("{") {
			braceCount++
			p.advance()
		} else if p.checkValue("}") {
			braceCount--
			if braceCount == 0 {
				class.EndLine = p.current().Line
			}
			p.advance()
		} else if p.checkKeyword("public") || p.checkKeyword("private") || p.checkKeyword("protected") {
			p.advance()
			p.matchValue(":") // skip the colon
		} else if p.isDestructorStart(className) {
			if fn := p.parseDestructor(className); fn != nil {
				class.Destructor = fn
			}
		} else if p.isConstructorStart(className) {
			if fn := p.parseConstructor(className); fn != nil {
				class.Constructor = fn
			}
		} else if p.isMemberDeclaration() {
			if member := p.parseMember(); member != nil {
				class.Members = append(class.Members, *member)
			}
		} else if p.isFunctionStart() {
			if fn := p.parseMethod(); fn != nil {
				class.Methods = append(class.Methods, *fn)
			}
		} else {
			p.advance()
		}
	}

	return class
}

func (p *Parser) isDestructorStart(className string) bool {
	if p.checkValue("~") {
		// Look ahead for class name
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == className {
			return true
		}
	}
	// virtual ~ClassName
	if p.checkKeyword("virtual") {
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == "~" {
			return true
		}
	}
	return false
}

func (p *Parser) isConstructorStart(className string) bool {
	if p.check(TokenIdent) && p.current().Value == className {
		// Look ahead for (
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == "(" {
			return true
		}
	}
	return false
}

func (p *Parser) parseDestructor(className string) *Function {
	startLine := p.current().Line

	// Skip virtual if present
	if p.checkKeyword("virtual") {
		p.advance()
	}

	// Skip ~
	p.matchValue("~")

	// Skip class name
	p.advance()

	// Skip parameters ()
	if !p.matchValue("(") {
		return nil
	}
	for !p.isAtEnd() && !p.checkValue(")") {
		p.advance()
	}
	p.matchValue(")")

	fn := &Function{
		Name:         "~" + className,
		IsDestructor: true,
		StartLine:    startLine,
	}

	// Parse body or skip declaration
	if p.checkValue(";") {
		p.advance()
		return fn
	}

	if p.checkValue("{") {
		p.parseFunctionBody(fn)
	}

	return fn
}

func (p *Parser) parseConstructor(className string) *Function {
	startLine := p.current().Line

	// Skip class name
	p.advance()

	// Parse parameters
	if !p.matchValue("(") {
		return nil
	}
	for !p.isAtEnd() && !p.checkValue(")") {
		p.advance()
	}
	p.matchValue(")")

	fn := &Function{
		Name:      className,
		StartLine: startLine,
	}

	// Skip initializer list
	if p.checkValue(":") {
		p.advance()
		for !p.isAtEnd() && !p.checkValue("{") && !p.checkValue(";") {
			p.advance()
		}
	}

	// Parse body or skip declaration
	if p.checkValue(";") {
		p.advance()
		return fn
	}

	if p.checkValue("{") {
		p.parseFunctionBody(fn)
	}

	return fn
}

func (p *Parser) parseMethod() *Function {
	startLine := p.current().Line

	// Skip return type and modifiers
	for !p.isAtEnd() && !p.checkValue("(") && !p.checkValue(";") && !p.checkValue("{") {
		p.advance()
	}

	if p.checkValue(";") {
		p.advance()
		return nil
	}

	// Get function name (token before '(')
	funcName := ""
	if p.pos > 0 {
		funcName = p.tokens[p.pos-1].Value
	}

	if !p.matchValue("(") {
		return nil
	}

	// Skip parameters
	parenCount := 1
	for !p.isAtEnd() && parenCount > 0 {
		if p.checkValue("(") {
			parenCount++
		} else if p.checkValue(")") {
			parenCount--
		}
		p.advance()
	}

	fn := &Function{
		Name:      funcName,
		StartLine: startLine,
	}

	// Skip const, noexcept, etc.
	for p.checkKeyword("const") || p.check(TokenIdent) {
		if p.checkValue("{") || p.checkValue(";") {
			break
		}
		p.advance()
	}

	if p.checkValue(";") {
		p.advance()
		return fn
	}

	if p.checkValue("{") {
		p.parseFunctionBody(fn)
	}

	return fn
}

func (p *Parser) parseFunctionBody(fn *Function) {
	if !p.matchValue("{") {
		return
	}

	braceCount := 1
	for !p.isAtEnd() && braceCount > 0 {
		if p.checkValue("{") {
			braceCount++
			p.advance()
		} else if p.checkValue("}") {
			braceCount--
			if braceCount == 0 {
				fn.EndLine = p.current().Line
			}
			p.advance()
		} else if p.checkKeyword("new") {
			alloc := p.parseAllocation()
			if alloc != nil {
				fn.Allocations = append(fn.Allocations, *alloc)
			}
		} else if p.checkKeyword("delete") {
			dealloc := p.parseDeallocation()
			if dealloc != nil {
				fn.Deallocations = append(fn.Deallocations, *dealloc)
			}
		} else if p.check(TokenIdent) {
			identName := p.current().Value
			identLine := p.current().Line

			// Check for method calls
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == "(" {
				fn.MethodCalls = append(fn.MethodCalls, identName)
			}

			// Check for pointer aliasing: ptr2 = ptr1 (where both are identifiers, no 'new')
			// Pattern: ident = ident ; (without 'new' keyword in between)
			if alias := p.checkPointerAlias(identName, identLine); alias != nil {
				fn.Aliases = append(fn.Aliases, *alias)
			}

			p.advance()
		} else {
			p.advance()
		}
	}
}

// checkPointerAlias checks if current position is a pointer alias assignment
// Pattern: target = source; (where source is an identifier, not 'new')
func (p *Parser) checkPointerAlias(targetName string, line int) *PointerAlias {
	// Look ahead: ident = ident ;
	if p.pos+3 >= len(p.tokens) {
		return nil
	}

	// Check pattern: current(ident) = next(ident) ; (or other terminator)
	if p.tokens[p.pos+1].Value != "=" {
		return nil
	}

	nextTok := p.tokens[p.pos+2]
	if nextTok.Type != TokenIdent {
		return nil
	}

	// Make sure it's not: ident = new ...
	if nextTok.Value == "new" {
		return nil
	}

	// Check it ends with ; or is followed by something reasonable
	if p.pos+3 < len(p.tokens) {
		afterSource := p.tokens[p.pos+3]
		if afterSource.Value == ";" || afterSource.Value == "}" || afterSource.Value == "," {
			return &PointerAlias{
				TargetVar: targetName,
				SourceVar: nextTok.Value,
				Line:      line,
			}
		}
	}

	return nil
}

func (p *Parser) parseAllocation() *Allocation {
	line := p.current().Line
	p.advance() // skip 'new'

	isArray := false
	if p.checkValue("[") {
		isArray = true
	}

	// Look for variable being assigned
	// Pattern: varName = new Type or this->varName = new Type
	// We need to look backwards for the variable name
	varName := p.findAssignmentTarget()

	if varName == "" {
		return nil
	}

	// Skip to end of statement
	for !p.isAtEnd() && !p.checkValue(";") && !p.checkValue("{") {
		if p.checkValue("[") {
			isArray = true
		}
		p.advance()
	}

	return &Allocation{
		VarName: varName,
		IsArray: isArray,
		Line:    line,
	}
}

func (p *Parser) findAssignmentTarget() string {
	// Look backwards for pattern: varName = or this->varName =
	for i := p.pos - 1; i >= 0 && i > p.pos-10; i-- {
		if p.tokens[i].Value == "=" {
			// Found assignment, look for variable before it
			for j := i - 1; j >= 0 && j > i-5; j-- {
				if p.tokens[j].Type == TokenIdent && p.tokens[j].Value != "this" {
					return p.tokens[j].Value
				}
			}
		}
	}
	return ""
}

func (p *Parser) parseDeallocation() *Deallocation {
	line := p.current().Line
	p.advance() // skip 'delete'

	isArray := false
	if p.checkValue("[") {
		isArray = true
		p.advance()       // skip [
		p.matchValue("]") // skip ]
	}

	// Get the variable being deleted
	varName := ""

	// Check for this-> prefix (this is a KEYWORD, not ident)
	if p.checkKeyword("this") {
		p.advance() // skip 'this'
		if p.checkValue("->") {
			p.advance() // skip '->'
			if p.check(TokenIdent) {
				varName = p.current().Value
			}
		}
	} else if p.check(TokenIdent) {
		varName = p.current().Value
	}

	if varName == "" {
		return nil
	}

	return &Deallocation{
		VarName: varName,
		IsArray: isArray,
		Line:    line,
	}
}

func (p *Parser) isMemberDeclaration() bool {
	// Look for pattern: Type* varName; or Type *varName;
	// Must contain a pointer indicator
	savedPos := p.pos
	hasPointer := false
	hasIdent := false

	for i := 0; i < 10 && savedPos+i < len(p.tokens); i++ {
		tok := p.tokens[savedPos+i]
		if tok.Value == ";" {
			break
		}
		if tok.Value == "(" || tok.Value == "{" {
			return false // It's a function
		}
		if tok.Value == "*" {
			hasPointer = true
		}
		if tok.Type == TokenIdent {
			hasIdent = true
		}
	}

	return hasPointer && hasIdent
}

func (p *Parser) parseMember() *Member {
	startLine := p.current().Line
	var tokens []Token

	// Collect tokens until semicolon
	for !p.isAtEnd() && !p.checkValue(";") {
		tokens = append(tokens, p.current())
		p.advance()
	}
	p.matchValue(";")

	if len(tokens) < 2 {
		return nil
	}

	// Find pointer and variable name
	isPointer := false
	isArray := false
	varName := ""
	var typeTokens []string

	for i, tok := range tokens {
		if tok.Value == "*" {
			isPointer = true
		} else if tok.Value == "[" {
			isArray = true
		} else if tok.Type == TokenIdent {
			// Last identifier before ; is the variable name
			if i == len(tokens)-1 || tokens[i+1].Value == "[" || tokens[i+1].Value == "=" {
				varName = tok.Value
			} else {
				typeTokens = append(typeTokens, tok.Value)
			}
		}
	}

	if !isPointer || varName == "" {
		return nil
	}

	return &Member{
		Name:      varName,
		Type:      strings.Join(typeTokens, " "),
		IsPointer: isPointer,
		IsArray:   isArray,
		Line:      startLine,
	}
}

func (p *Parser) isFunctionStart() bool {
	// Look ahead for pattern: ... name(...)
	savedPos := p.pos
	for i := 0; i < 15 && savedPos+i < len(p.tokens); i++ {
		tok := p.tokens[savedPos+i]
		if tok.Value == ";" {
			return false
		}
		if tok.Value == "(" {
			return true
		}
		if tok.Value == "{" || tok.Value == "}" {
			return false
		}
	}
	return false
}

// Token navigation helpers
func (p *Parser) current() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == TokenEOF
}

func (p *Parser) check(tokenType TokenType) bool {
	return !p.isAtEnd() && p.current().Type == tokenType
}

func (p *Parser) checkValue(value string) bool {
	return !p.isAtEnd() && p.current().Value == value
}

func (p *Parser) checkKeyword(keyword string) bool {
	return p.check(TokenKeyword) && p.current().Value == keyword
}

func (p *Parser) matchValue(value string) bool {
	if p.checkValue(value) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) matchKeyword(keyword string) bool {
	if p.checkKeyword(keyword) {
		p.advance()
		return true
	}
	return false
}
