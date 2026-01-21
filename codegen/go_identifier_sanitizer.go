package codegen

import "github.com/quant5-lab/runner/util"

/*
SanitizeGoIdentifier is deprecated. Use util.SanitizeGoIdentifier instead.
Kept for backward compatibility with existing codegen code.
*/
func SanitizeGoIdentifier(name string) string {
	return util.SanitizeGoIdentifier(name)
}
