package analyzer

import (
	"leakcheck/internal/parser"
)

// MaxMethodDepth is the maximum depth to follow method calls
const MaxMethodDepth = 5

// Analyzer detects memory leaks in parsed C++ classes
type Analyzer struct {
	classes []parser.Class
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AddClasses adds parsed classes to analyze
func (a *Analyzer) AddClasses(classes []parser.Class) {
	a.classes = append(a.classes, classes...)
}

// Analyze performs leak detection and returns found issues
func (a *Analyzer) Analyze() []parser.Leak {
	var leaks []parser.Leak

	for _, class := range a.classes {
		classLeaks := a.analyzeClass(class)
		leaks = append(leaks, classLeaks...)
	}

	return leaks
}

func (a *Analyzer) analyzeClass(class parser.Class) []parser.Leak {
	var leaks []parser.Leak

	// Get all pointer members
	pointerMembers := make(map[string]parser.Member)
	for _, m := range class.Members {
		if m.IsPointer {
			pointerMembers[m.Name] = m
		}
	}

	if len(pointerMembers) == 0 {
		return nil
	}

	// Track allocations in constructor
	allocatedVars := make(map[string]parser.Allocation)
	if class.Constructor != nil {
		for _, alloc := range class.Constructor.Allocations {
			allocatedVars[alloc.VarName] = alloc
		}
	}

	// Build method map for quick lookup
	methodMap := make(map[string]*parser.Function)
	for i := range class.Methods {
		methodMap[class.Methods[i].Name] = &class.Methods[i]
	}

	// Track deallocations in destructor using MULTI-LEVEL method tracking
	deallocatedVars := make(map[string]parser.Deallocation)
	aliasMap := buildAliasMap(class) // Build pointer alias map

	if class.Destructor != nil {
		// Collect all deallocations recursively (multi-level)
		collectDeallocations(class.Destructor, methodMap, deallocatedVars, MaxMethodDepth, make(map[string]bool))
	}

	// Rule 1: Allocated in constructor but not deleted in destructor
	for varName, alloc := range allocatedVars {
		// Check direct delete or delete through alias
		deleted := isVarDeallocated(varName, deallocatedVars, aliasMap)

		if !deleted {
			leaks = append(leaks, parser.Leak{
				File:      class.File,
				Line:      alloc.Line,
				ClassName: class.Name,
				VarName:   varName,
				Reason:    "allocated with 'new' but not deleted in destructor",
				Severity:  "error",
			})
		} else {
			// Check for array mismatch
			dealloc := findDeallocation(varName, deallocatedVars, aliasMap)
			if dealloc != nil {
				if alloc.IsArray && !dealloc.IsArray {
					leaks = append(leaks, parser.Leak{
						File:      class.File,
						Line:      dealloc.Line,
						ClassName: class.Name,
						VarName:   varName,
						Reason:    "allocated with 'new[]' but deleted with 'delete' instead of 'delete[]'",
						Severity:  "error",
					})
				} else if !alloc.IsArray && dealloc.IsArray {
					leaks = append(leaks, parser.Leak{
						File:      class.File,
						Line:      dealloc.Line,
						ClassName: class.Name,
						VarName:   varName,
						Reason:    "allocated with 'new' but deleted with 'delete[]' instead of 'delete'",
						Severity:  "warning",
					})
				}
			}
		}
	}

	// Rule 2: Pointer reassignment without prior delete in methods
	for _, method := range class.Methods {
		for _, alloc := range method.Allocations {
			if _, exists := pointerMembers[alloc.VarName]; exists {
				// Check if this variable is deallocated before reassignment in the same method
				hasDeleteBeforeNew := false
				for _, dealloc := range method.Deallocations {
					if dealloc.VarName == alloc.VarName && dealloc.Line < alloc.Line {
						hasDeleteBeforeNew = true
						break
					}
				}

				if !hasDeleteBeforeNew {
					// Check if there's an existing allocation (reassignment without delete)
					if _, wasAllocatedInCtor := allocatedVars[alloc.VarName]; wasAllocatedInCtor {
						leaks = append(leaks, parser.Leak{
							File:      class.File,
							Line:      alloc.Line,
							ClassName: class.Name,
							VarName:   alloc.VarName,
							Reason:    "pointer reassigned with 'new' without deleting previous allocation (in " + method.Name + ")",
							Severity:  "warning",
						})
					}
				}
			}
		}
	}

	// Rule 3: Pointer aliasing - delete through alias is valid, but warn about potential issues
	for _, method := range class.Methods {
		for _, alias := range method.Aliases {
			if _, isPointerMember := pointerMembers[alias.SourceVar]; isPointerMember {
				// Check if target is later deleted but source is also deleted (double delete)
				sourceDeleted := false
				targetDeleted := false
				for _, dealloc := range method.Deallocations {
					if dealloc.VarName == alias.SourceVar {
						sourceDeleted = true
					}
					if dealloc.VarName == alias.TargetVar {
						targetDeleted = true
					}
				}
				if sourceDeleted && targetDeleted {
					leaks = append(leaks, parser.Leak{
						File:      class.File,
						Line:      alias.Line,
						ClassName: class.Name,
						VarName:   alias.SourceVar,
						Reason:    "pointer aliased to '" + alias.TargetVar + "' and both are deleted (potential double-free)",
						Severity:  "error",
					})
				}
			}
		}
	}

	// Rule 4: No destructor but has allocations
	if class.Destructor == nil {
		for _, member := range pointerMembers {
			if _, allocated := allocatedVars[member.Name]; allocated {
				leaks = append(leaks, parser.Leak{
					File:      class.File,
					Line:      member.Line,
					ClassName: class.Name,
					VarName:   member.Name,
					Reason:    "pointer member allocated but class has no destructor",
					Severity:  "error",
				})
			}
		}
	}

	return leaks
}

// collectDeallocations recursively collects deallocations from a function and its called methods
func collectDeallocations(fn *parser.Function, methodMap map[string]*parser.Function,
	result map[string]parser.Deallocation, depth int, visited map[string]bool) {

	if depth <= 0 || fn == nil {
		return
	}

	// Prevent infinite recursion
	if visited[fn.Name] {
		return
	}
	visited[fn.Name] = true

	// Add direct deallocations
	for _, dealloc := range fn.Deallocations {
		result[dealloc.VarName] = dealloc
	}

	// Recurse into called methods
	for _, methodName := range fn.MethodCalls {
		if calledMethod, exists := methodMap[methodName]; exists {
			collectDeallocations(calledMethod, methodMap, result, depth-1, visited)
		}
	}
}

// buildAliasMap creates a map of source -> targets for pointer aliases
func buildAliasMap(class parser.Class) map[string][]string {
	aliasMap := make(map[string][]string)

	// Collect aliases from all functions
	collectAliasesFromFunc := func(fn *parser.Function) {
		if fn == nil {
			return
		}
		for _, alias := range fn.Aliases {
			aliasMap[alias.SourceVar] = append(aliasMap[alias.SourceVar], alias.TargetVar)
			// Also reverse: if we delete target, it's like deleting source
			aliasMap[alias.TargetVar] = append(aliasMap[alias.TargetVar], alias.SourceVar)
		}
	}

	if class.Constructor != nil {
		collectAliasesFromFunc(class.Constructor)
	}
	if class.Destructor != nil {
		collectAliasesFromFunc(class.Destructor)
	}
	for i := range class.Methods {
		collectAliasesFromFunc(&class.Methods[i])
	}

	return aliasMap
}

// isVarDeallocated checks if a variable is deallocated directly or through an alias
func isVarDeallocated(varName string, deallocatedVars map[string]parser.Deallocation, aliasMap map[string][]string) bool {
	// Direct check
	if _, deleted := deallocatedVars[varName]; deleted {
		return true
	}

	// Check aliases
	if aliases, hasAliases := aliasMap[varName]; hasAliases {
		for _, aliasName := range aliases {
			if _, deleted := deallocatedVars[aliasName]; deleted {
				return true
			}
		}
	}

	return false
}

// findDeallocation finds the deallocation for a variable (direct or through alias)
func findDeallocation(varName string, deallocatedVars map[string]parser.Deallocation, aliasMap map[string][]string) *parser.Deallocation {
	// Direct check
	if dealloc, deleted := deallocatedVars[varName]; deleted {
		return &dealloc
	}

	// Check aliases
	if aliases, hasAliases := aliasMap[varName]; hasAliases {
		for _, aliasName := range aliases {
			if dealloc, deleted := deallocatedVars[aliasName]; deleted {
				return &dealloc
			}
		}
	}

	return nil
}

// AnalyzeClasses is a convenience function to analyze classes directly
func AnalyzeClasses(classes []parser.Class) []parser.Leak {
	analyzer := NewAnalyzer()
	analyzer.AddClasses(classes)
	return analyzer.Analyze()
}
