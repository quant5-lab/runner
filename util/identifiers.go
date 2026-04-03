package util

import "strings"

/* Go reserved keywords and built-in functions that conflict with variable names */
var goReservedWords = map[string]bool{
	/* Keywords */
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,

	/* Built-in functions */
	"append": true, "cap": true, "close": true, "complex": true, "copy": true,
	"delete": true, "imag": true, "len": true, "make": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true, "recover": true,

	/* Built-in types */
	"bool": true, "byte": true, "complex64": true, "complex128": true,
	"error": true, "float32": true, "float64": true, "int": true, "int8": true,
	"int16": true, "int32": true, "int64": true, "rune": true, "string": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,

	/* Predeclared constants */
	"true": true, "false": true, "iota": true, "nil": true,
}

/*
SanitizeGoIdentifier converts Pine variable names to valid Go identifiers,
escaping reserved words that would cause compile errors.

Strategy: Append underscore suffix (_len, _type, _map) to avoid collisions.
*/
func SanitizeGoIdentifier(name string) string {
	/* Empty name - should not happen, but handle defensively */
	if name == "" {
		return "_unnamed"
	}

	/* Check if name conflicts with Go reserved word */
	if goReservedWords[strings.ToLower(name)] {
		return name + "_"
	}

	return name
}
