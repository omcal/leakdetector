# LeakCheck - C++ Memory Leak Detector

A static analysis tool written in Go that detects potential memory leaks in C++ code by analyzing `new`/`delete` patterns in classes.

## Features

- 🔍 **Detects missing `delete`** - Finds allocations in constructors without matching deallocations in destructors
- ⚠️ **Array mismatch detection** - Flags `new[]` with `delete` instead of `delete[]`
- 🔄 **Reassignment leaks** - Detects pointer reassignment without prior delete
- 📁 **Recursive scanning** - Scans `.cpp`, `.h`, `.hpp` files recursively
- 🚫 **Folder exclusion** - Skip directories like `vendor`, `build`, `third_party`
- 📊 **JSON output** - Export results for CI/CD integration

## Installation

### From Source

```bash
go build -o leakcheck ./cmd/leakcheck
```

### Docker

```bash
docker build -t leakcheck .
```

## Usage

### Command Line

```bash
# Scan a directory
./leakcheck ./src

# Scan with exclusions
./leakcheck --exclude=vendor,build ./

# JSON output
./leakcheck --json ./src > report.json

# Show help
./leakcheck --help
```

### Docker

```bash
# Scan mounted source code
docker run --rm -v /path/to/your/code:/src leakcheck

# With exclusions
docker run --rm -v /path/to/your/code:/src leakcheck --exclude=vendor .

# Save JSON report
docker run --rm -v /path/to/your/code:/src leakcheck > report.json
```

## Output Examples

### Console Output

```
Scanning 15 file(s)...
Found 8 class(es) with pointer members

leak_sample.cpp:
  ❌ Line 14 [LeakyClass::name]: allocated with 'new' but not deleted in destructor
  ❌ Line 15 [LeakyClass::data]: allocated with 'new' but not deleted in destructor
  ❌ Line 33 [ArrayMismatch::arr]: allocated with 'new[]' but deleted with 'delete' instead of 'delete[]'
  ⚠️  Line 47 [ReassignmentLeak::ptr]: pointer reassigned with 'new' without deleting previous allocation

Summary: 3 error(s), 1 warning(s)
```

### JSON Output

```json
{
  "leaks": [
    {
      "file": "/path/to/leak_sample.cpp",
      "line": 14,
      "class": "LeakyClass",
      "variable": "name",
      "reason": "allocated with 'new' but not deleted in destructor",
      "severity": "error"
    }
  ],
  "summary": {
    "total_issues": 1,
    "errors": 1,
    "warnings": 0
  }
}
```

## Detection Rules

| Rule | Severity | Description |
|------|----------|-------------|
| Missing delete | Error | Variable allocated with `new` but not deleted in destructor |
| Array mismatch | Error | `new[]` paired with `delete` or vice versa |
| Reassignment leak | Warning | Pointer reassigned without deleting previous value |
| No destructor | Error | Class allocates memory but has no destructor |

## Limitations

- Static analysis only - cannot detect runtime-conditional leaks
- Does not track smart pointers (`std::unique_ptr`, `std::shared_ptr`)
- Does not analyze `malloc`/`free` (C-style allocations)
- Method call tracking limited to 1 level deep from destructor

## License

MIT
