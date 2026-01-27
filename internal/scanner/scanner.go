package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// Scanner recursively finds C++ files in directories
type Scanner struct {
	Excludes []string
}

// NewScanner creates a new file scanner with exclusion patterns
func NewScanner(excludes []string) *Scanner {
	return &Scanner{Excludes: excludes}
}

// ScanPath scans a file or directory for C++ files
func (s *Scanner) ScanPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if s.isCppFile(path) {
			return []string{path}, nil
		}
		return nil, nil
	}

	var files []string
	err = filepath.WalkDir(path, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files/dirs with errors
		}

		// Check if this directory should be excluded
		if d.IsDir() {
			if s.shouldExclude(filePath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a C++ file
		if s.isCppFile(filePath) && !s.shouldExclude(filePath) {
			files = append(files, filePath)
		}

		return nil
	})

	return files, err
}

// ScanPaths scans multiple paths for C++ files
func (s *Scanner) ScanPaths(paths []string) ([]string, error) {
	var allFiles []string
	seen := make(map[string]bool)

	for _, path := range paths {
		files, err := s.ScanPath(path)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			absPath, _ := filepath.Abs(f)
			if !seen[absPath] {
				seen[absPath] = true
				allFiles = append(allFiles, absPath)
			}
		}
	}

	return allFiles, nil
}

func (s *Scanner) isCppFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".cpp" || ext == ".h" || ext == ".hpp" ||
		ext == ".cc" || ext == ".cxx" || ext == ".hxx"
}

func (s *Scanner) shouldExclude(path string) bool {
	for _, exclude := range s.Excludes {
		// Match against directory name or path component
		base := filepath.Base(path)
		if base == exclude {
			return true
		}
		// Also check if exclude pattern is in the path
		if strings.Contains(path, string(filepath.Separator)+exclude+string(filepath.Separator)) {
			return true
		}
		if strings.HasSuffix(path, string(filepath.Separator)+exclude) {
			return true
		}
	}
	return false
}
