package parser

// Token represents a lexical token from C++ source
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenNumber
	TokenString
	TokenKeyword
	TokenOperator
	TokenPunctuation
	TokenNewline
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// Class represents a C++ class or struct
type Class struct {
	Name        string
	File        string
	StartLine   int
	EndLine     int
	Members     []Member
	Constructor *Function
	Destructor  *Function
	Methods     []Function
}

// Member represents a class member variable
type Member struct {
	Name      string
	Type      string
	IsPointer bool
	IsArray   bool
	Line      int
}

// Function represents a class method (constructor, destructor, or regular method)
type Function struct {
	Name          string
	IsDestructor  bool
	StartLine     int
	EndLine       int
	Allocations   []Allocation
	Deallocations []Deallocation
	MethodCalls   []string       // Methods called within this function
	Aliases       []PointerAlias // Pointer aliasing within this function
}

// Allocation represents a dynamic memory allocation
type Allocation struct {
	VarName string
	IsArray bool // true for new[], false for new
	Line    int
}

// Deallocation represents a dynamic memory deallocation
type Deallocation struct {
	VarName string
	IsArray bool // true for delete[], false for delete
	Line    int
}

// PointerAlias represents when one pointer is assigned to another
type PointerAlias struct {
	SourceVar string // original pointer (e.g., ptr1)
	TargetVar string // alias pointer (e.g., ptr2 = ptr1)
	Line      int
}

// Leak represents a detected memory leak
type Leak struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	ClassName string `json:"class"`
	VarName   string `json:"variable"`
	Reason    string `json:"reason"`
	Severity  string `json:"severity"` // "error", "warning"
}
