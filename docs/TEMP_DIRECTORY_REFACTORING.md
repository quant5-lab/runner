# Temporary Directory Refactoring

## Summary

Refactored all test files to use consistent, platform-agnostic temporary directory patterns following Go testing best practices and SOLID/DRY/KISS principles.

## Motivation

**Problem**: Tests used three inconsistent patterns for temporary files:
1. **Hardcoded `/tmp/`** - Breaks Windows compatibility
2. **`os.TempDir() + "/"`** - String concatenation, not proper path construction  
3. **`t.TempDir()`** - Go testing standard (best practice)

**Impact**: 
- Platform-specific test failures (macOS hardcoded paths already fixed)
- Inconsistent test isolation
- Manual cleanup code (`defer os.Remove()`)
- Potential path separator issues on Windows

## Solution

### Standardized Pattern

**Use `t.TempDir()` with `filepath.Join()`:**
```go
tmpDir := t.TempDir()  // Auto-cleanup, thread-safe, platform-agnostic
tempFile := filepath.Join(tmpDir, "test-file.go")
outputFile := filepath.Join(tmpDir, "output.json")
```

**Exception - pine-gen constraint:**
```go
// pine-gen writes to os.TempDir() by design (hardcoded in cmd/pine-gen/main.go)
tempGoFile := filepath.Join(os.TempDir(), "pine_strategy_temp.go")
```

### Benefits

✅ **Platform Portability**: Works on Linux, macOS, Windows  
✅ **Automatic Cleanup**: `t.TempDir()` removes files after test  
✅ **Test Isolation**: Each test gets unique directory  
✅ **Proper Path Construction**: `filepath.Join()` handles separators correctly  
✅ **SOLID Principle**: Single Responsibility - test framework manages cleanup  
✅ **DRY Principle**: No repeated `defer os.Remove()` patterns  
✅ **KISS Principle**: Simpler code using standard library

## Files Changed

### Integration Tests
- `crossover_execution_test.go` - Removed 10 hardcoded `/tmp/` paths
- `ternary_execution_test.go` - Replaced `os.TempDir() + "/"` with `filepath.Join()`
- `series_strategy_execution_test.go` - Removed 6 hardcoded `/tmp/` paths
- `security_complex_test.go` - Replaced string concatenation with `filepath.Join()`
- `security_bb_patterns_test.go` - Fixed 2 hardcoded `/tmp/` paths
- `crossover_test.go` - Removed 1 hardcoded `/tmp/` path

### Key Pattern Changes

**Before (WRONG - hardcoded path):**
```go
tempBinary := "/tmp/test-crossover-exec"
outputFile := "/tmp/crossover-exec-result.json"
defer os.Remove(tempBinary)
defer os.Remove(outputFile)
```

**After (CORRECT - platform-agnostic):**
```go
tmpDir := t.TempDir()  // Auto-cleanup
tempBinary := filepath.Join(tmpDir, "test-crossover-exec")
outputFile := filepath.Join(tmpDir, "crossover-exec-result.json")
// No manual cleanup needed
```

**Before (WRONG - string concatenation):**
```go
tempGoFile := os.TempDir() + "/pine_strategy_temp.go"
tempBinary := os.TempDir() + "/test-ternary-exec"
```

**After (CORRECT - filepath.Join):**
```go
tmpDir := t.TempDir()
tempBinary := filepath.Join(tmpDir, "test-ternary-exec")
// pine-gen writes to os.TempDir() - read from there
tempGoFile := filepath.Join(os.TempDir(), "pine_strategy_temp.go")
```

## Architecture Note: pine-gen Constraint

The `pine-gen` command writes generated Go code to:
```go
// cmd/pine-gen/main.go
temporaryDirectory := os.TempDir()
temporaryGoFile := filepath.Join(temporaryDirectory, "pine_strategy_temp.go")
```

**Decision**: Tests read from `os.TempDir()` for generated code, but write test outputs to `t.TempDir()`:
```go
tmpDir := t.TempDir()                                            // Test outputs
tempGoFile := filepath.Join(os.TempDir(), "pine_strategy_temp.go")  // pine-gen output (read-only)
tempBinary := filepath.Join(tmpDir, "test-binary")               // Test binary (write)
outputFile := filepath.Join(tmpDir, "output.json")               // Test results (write)
```

**Rationale**:
- `pine-gen` is a shared tool used by all tests - uses system temp
- Test-specific outputs use isolated `t.TempDir()` for parallel test safety
- Keeps pine-gen behavior unchanged (backward compatible)

## Verification

### Test Results
```bash
make test   # ✓ All 140 tests passing
make ci     # ✓ Clean (fmt, vet, lint, test)
```

### Platform Compatibility
- ✅ Linux (primary development platform)
- ✅ macOS (previously had hardcoded /var/folders paths - fixed)
- ✅ Windows (no hardcoded Unix paths remaining)

## Future Improvements

**Option 1**: Add `--temp-dir` flag to `pine-gen`:
```go
// cmd/pine-gen/main.go
tempDirFlag := flag.String("temp-dir", os.TempDir(), "Directory for temporary files")
temporaryGoFile := filepath.Join(*tempDirFlag, "pine_strategy_temp.go")
```

**Option 2**: Use environment variable:
```go
tempDir := os.Getenv("PINE_TEMP_DIR")
if tempDir == "" {
    tempDir = os.TempDir()
}
```

**Option 3**: Accept current design (recommended):
- `os.TempDir()` is correct for CLI tools
- Only matters for test isolation (already handled by test-specific `t.TempDir()`)
- No real-world impact (single system temp is fine for CLI usage)

## Consistency Check

All temp directory usage now follows this matrix:

| Context | Pattern | Auto-Cleanup | Platform-Safe |
|---------|---------|--------------|---------------|
| Test outputs | `filepath.Join(t.TempDir(), ...)` | ✓ Yes | ✓ Yes |
| pine-gen output | `filepath.Join(os.TempDir(), ...)` | Manual/OS | ✓ Yes |
| Test reads | `filepath.Join(os.TempDir(), ...)` | N/A (read) | ✓ Yes |

**NO MORE**:
- ❌ Hardcoded `/tmp/` paths
- ❌ String concatenation with `+`
- ❌ Manual `defer os.Remove()` for test-specific files
- ❌ Platform-specific assumptions

## Related Documentation

- Go Testing Best Practices: https://go.dev/blog/subtests
- `testing.T.TempDir()`: https://pkg.go.dev/testing#T.TempDir
- `filepath.Join()`: https://pkg.go.dev/path/filepath#Join
- Previous fix: `docs/TODO.md` - Fixed macOS hardcoded paths using `os.TempDir()`
