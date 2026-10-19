package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

const unversionedSessionCallPrefix = "session.TimeFunc("

type sessionCodegenPath struct {
	name string
	emit func(pineVersion int, session SessionArgument) string
}

func allSessionCodegenPaths(tfArg ast.Expression) []sessionCodegenPath {
	return []sessionCodegenPath{
		{
			name: "TimeCodeGenerator/GenerateWithSession",
			emit: func(pv int, sa SessionArgument) string {
				return NewTimeCodeGeneratorWithVersion("\t", pv).GenerateWithSession("v", sa)
			},
		},
		{
			name: "TimeHandler/HandleVariableInit",
			emit: func(pv int, sa SessionArgument) string {
				var sessionExpr ast.Expression
				switch sa.Type {
				case ArgumentTypeLiteral:
					sessionExpr = &ast.Literal{Value: fmt.Sprintf("%q", sa.Value)}
				case ArgumentTypeIdentifier:
					sessionExpr = &ast.Identifier{Name: sa.Value}
				case ArgumentTypeWrappedIdentifier:
					sessionExpr = &ast.MemberExpression{
						Computed: true,
						Object:   &ast.Identifier{Name: sa.Value},
						Property: &ast.Literal{Value: 0},
					}
				default:
					sessionExpr = &ast.Literal{Value: 999}
				}
				return NewTimeHandlerWithVersion("\t", pv).HandleVariableInit("v",
					&ast.CallExpression{Arguments: []ast.Expression{tfArg, sessionExpr}})
			},
		},
		{
			name: "TimeHandler/HandleInlineExpression",
			emit: func(pv int, sa SessionArgument) string {
				var sessionExpr ast.Expression
				switch sa.Type {
				case ArgumentTypeLiteral:
					sessionExpr = &ast.Literal{Value: fmt.Sprintf("%q", sa.Value)}
				case ArgumentTypeIdentifier:
					sessionExpr = &ast.Identifier{Name: sa.Value}
				case ArgumentTypeWrappedIdentifier:
					sessionExpr = &ast.MemberExpression{
						Computed: true,
						Object:   &ast.Identifier{Name: sa.Value},
						Property: &ast.Literal{Value: 0},
					}
				default:
					sessionExpr = &ast.Literal{Value: 999}
				}
				return NewTimeHandlerWithVersion("\t", pv).HandleInlineExpression(
					[]ast.Expression{tfArg, sessionExpr})
			},
		},
	}
}

// Regression to the unversioned call silently applies the v5 all-7-days default
// regardless of the script's //@version, breaking the v4 Mon-Fri behaviour.
func TestTimeSession_VersionedCallContract(t *testing.T) {
	tfArg := &ast.Identifier{Name: "timeframe.period"}
	paths := allSessionCodegenPaths(tfArg)

	sessions := []SessionArgument{
		{Type: ArgumentTypeLiteral, Value: "0950-1645"},
		{Type: ArgumentTypeIdentifier, Value: "mySession"},
		{Type: ArgumentTypeWrappedIdentifier, Value: "sessionVar"},
	}

	for _, pv := range []int{4, 5} {
		for _, sa := range sessions {
			for _, path := range paths {
				name := fmt.Sprintf("v%d/%s/%d", pv, path.name, sa.Type)
				pv, sa, path := pv, sa, path
				t.Run(name, func(t *testing.T) {
					out := path.emit(pv, sa)
					if strings.Contains(out, unversionedSessionCallPrefix) {
						t.Errorf("unversioned %q found — must use session.TimeFuncWithVersion:\n%s",
							unversionedSessionCallPrefix, out)
					}
					if !strings.Contains(out, "session.TimeFuncWithVersion") {
						t.Errorf("session.TimeFuncWithVersion not found in output:\n%s", out)
					}
				})
			}
		}
	}
}

// A wrong version integer silently applies the wrong DAYS default for every bar.
func TestTimeSession_VersionIntegerEmbedded(t *testing.T) {
	tfArg := &ast.Identifier{Name: "timeframe.period"}
	paths := allSessionCodegenPaths(tfArg)

	for _, tc := range []struct {
		pineVersion int
		wantSuffix  string
	}{
		{4, ", 4)"},
		{5, ", 5)"},
	} {
		for _, path := range paths {
			name := fmt.Sprintf("v%d/%s", tc.pineVersion, path.name)
			tc, path := tc, path
			t.Run(name, func(t *testing.T) {
				out := path.emit(tc.pineVersion, SessionArgument{
					Type: ArgumentTypeLiteral, Value: "0930-1600",
				})
				if !strings.Contains(out, tc.wantSuffix) {
					t.Errorf("want version suffix %q in output:\n%s", tc.wantSuffix, out)
				}
			})
		}
	}
}

func TestTimeSession_InvalidSessionEmitsNaN(t *testing.T) {
	tfArg := &ast.Identifier{Name: "timeframe.period"}

	for _, pv := range []int{4, 5} {
		for _, path := range allSessionCodegenPaths(tfArg) {
			name := fmt.Sprintf("v%d/%s", pv, path.name)
			pv, path := pv, path
			t.Run(name, func(t *testing.T) {
				out := path.emit(pv, SessionArgument{Type: ArgumentTypeUnknown})
				if strings.Contains(out, "session.TimeFuncWithVersion") {
					t.Errorf("invalid session should not emit session.TimeFuncWithVersion:\n%s", out)
				}
				if !strings.Contains(out, "math.NaN()") {
					t.Errorf("invalid session should emit math.NaN(), got:\n%s", out)
				}
			})
		}
	}
}
