package parser

import (
	"path/filepath"
	"strings"
)

// ClassRegistry holds all parsed classes for cross-file analysis
type ClassRegistry struct {
	// Classes by name (for matching header declarations with cpp implementations)
	classesByName map[string][]*Class
	// All classes
	allClasses []Class
}

// NewClassRegistry creates a new registry
func NewClassRegistry() *ClassRegistry {
	return &ClassRegistry{
		classesByName: make(map[string][]*Class),
	}
}

// AddClasses adds parsed classes to the registry
func (r *ClassRegistry) AddClasses(classes []Class) {
	for i := range classes {
		class := &classes[i]
		r.allClasses = append(r.allClasses, *class)
		r.classesByName[class.Name] = append(r.classesByName[class.Name], class)
	}
}

// MergeClasses merges class definitions split across header and implementation files
// Returns a list of fully merged classes
func (r *ClassRegistry) MergeClasses() []Class {
	merged := make(map[string]*Class)

	for _, class := range r.allClasses {
		existing, exists := merged[class.Name]
		if !exists {
			// First occurrence of this class
			classCopy := class
			merged[class.Name] = &classCopy
			continue
		}

		// Merge: combine information from multiple files
		r.mergeClassInto(existing, &class)
	}

	// Convert map to slice
	result := make([]Class, 0, len(merged))
	for _, class := range merged {
		result = append(result, *class)
	}
	return result
}

// mergeClassInto merges source class info into target
func (r *ClassRegistry) mergeClassInto(target, source *Class) {
	// Track which file is header vs implementation
	targetIsHeader := isHeaderFile(target.File)
	sourceIsHeader := isHeaderFile(source.File)

	// Merge members - always prefer header over implementation
	// Headers have the member declarations, cpp files typically don't repeat them
	if sourceIsHeader && !targetIsHeader {
		// Source is header, target is cpp - use source's members
		if len(source.Members) > 0 {
			target.Members = source.Members
		}
	} else if !sourceIsHeader && targetIsHeader {
		// Target is header, source is cpp - keep target's members (already in place)
	} else if len(target.Members) == 0 && len(source.Members) > 0 {
		// Both same type, take whichever has members
		target.Members = source.Members
	}

	// Merge constructor - prefer the one with actual function body (has allocations)
	if target.Constructor == nil && source.Constructor != nil {
		target.Constructor = source.Constructor
	} else if source.Constructor != nil && target.Constructor != nil {
		// Both have constructors - prefer the one with allocations (the implementation)
		if len(source.Constructor.Allocations) > 0 && len(target.Constructor.Allocations) == 0 {
			target.Constructor = source.Constructor
		}
	}

	// Merge destructor - prefer the one with actual function body
	if target.Destructor == nil && source.Destructor != nil {
		target.Destructor = source.Destructor
	} else if source.Destructor != nil && target.Destructor != nil {
		// Both have destructors - prefer the one with deallocations (the implementation)
		if len(source.Destructor.Deallocations) > 0 && len(target.Destructor.Deallocations) == 0 {
			target.Destructor = source.Destructor
		}
	}

	// Merge methods
	methodMap := make(map[string]*Function)
	for i := range target.Methods {
		methodMap[target.Methods[i].Name] = &target.Methods[i]
	}
	for _, method := range source.Methods {
		existing, exists := methodMap[method.Name]
		if !exists {
			target.Methods = append(target.Methods, method)
		} else if len(method.Allocations) > 0 || len(method.Deallocations) > 0 {
			// Source has more info, update
			*existing = method
		}
	}

	// Update file reference to include both
	if !strings.Contains(target.File, source.File) && target.File != source.File {
		target.File = target.File + ", " + filepath.Base(source.File)
	}
}

func isHeaderFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".h" || ext == ".hpp" || ext == ".hxx"
}
