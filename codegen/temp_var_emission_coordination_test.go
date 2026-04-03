package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTempVarEmissionTracker_StateManagement(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*TempVarEmissionTracker)
		varName  string
		expected bool
	}{
		{
			name:     "unmarked_var_not_emitted",
			setup:    func(t *TempVarEmissionTracker) {},
			varName:  "ta_sma_50",
			expected: false,
		},
		{
			name: "marked_var_is_emitted",
			setup: func(t *TempVarEmissionTracker) {
				t.MarkAsEmitted("ta_sma_50")
			},
			varName:  "ta_sma_50",
			expected: true,
		},
		{
			name: "multiple_vars_tracked_independently",
			setup: func(t *TempVarEmissionTracker) {
				t.MarkAsEmitted("ta_sma_50")
				t.MarkAsEmitted("ta_ema_20")
			},
			varName:  "ta_rsi_14",
			expected: false,
		},
		{
			name: "reset_clears_all_state",
			setup: func(t *TempVarEmissionTracker) {
				t.MarkAsEmitted("ta_sma_50")
				t.MarkAsEmitted("ta_ema_20")
				t.Reset()
			},
			varName:  "ta_sma_50",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTempVarEmissionTracker()
			tt.setup(tracker)

			if got := tracker.WasEmitted(tt.varName); got != tt.expected {
				t.Errorf("WasEmitted(%q) = %v, want %v", tt.varName, got, tt.expected)
			}
		})
	}
}

func TestTempVarInlineDeduplicator_EmissionDecision(t *testing.T) {
	tests := []struct {
		name         string
		setupManager func(*TempVariableManager) *ast.CallExpression
		shouldEmit   bool
		description  string
	}{
		{
			name: "unregistered_call_should_emit",
			setupManager: func(mgr *TempVariableManager) *ast.CallExpression {
				return &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
				}
			},
			shouldEmit:  true,
			description: "Call not in temp var registry should emit",
		},
		{
			name: "registered_not_emitted_should_emit",
			setupManager: func(mgr *TempVariableManager) *ast.CallExpression {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
				}
				mgr.GetOrCreate(CallInfo{
					Call:     call,
					FuncName: "ta.sma",
					ArgHash:  "test",
				})
				return call
			},
			shouldEmit:  true,
			description: "Registered but not yet emitted should emit",
		},
		{
			name: "registered_already_emitted_should_skip",
			setupManager: func(mgr *TempVariableManager) *ast.CallExpression {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
				}
				info := CallInfo{
					Call:      call,
					FuncName:  "ta.sma",
					ArgHash:   "test",
					StmtIndex: 0,
				}
				varName := mgr.GetOrCreate(info)
				mgr.emissionTracker.MarkAsEmitted(varName)
				return call
			},
			shouldEmit:  false,
			description: "Already emitted should skip to prevent duplication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()
			mgr := gen.tempVarMgr
			call := tt.setupManager(mgr)

			deduplicator := NewTempVarInlineDeduplicator(mgr)
			callInfo := CallInfo{
				Call:     call,
				FuncName: "ta.sma",
				ArgHash:  "test",
			}

			if got := deduplicator.ShouldEmitCalculation(callInfo); got != tt.shouldEmit {
				t.Errorf("%s: ShouldEmitCalculation() = %v, want %v", tt.description, got, tt.shouldEmit)
			}
		})
	}
}

func TestTempVarManager_EmissionCoordination(t *testing.T) {
	tests := []struct {
		name           string
		setupCalls     func(*TempVariableManager) []CallInfo
		stmtIndex      int
		verifyEmission func(t *testing.T, code string)
	}{
		{
			name: "single_indicator_emits_once",
			setupCalls: func(mgr *TempVariableManager) []CallInfo {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 20},
					},
				}
				info := CallInfo{
					Call:      call,
					FuncName:  "ta.sma",
					ArgHash:   "test",
					StmtIndex: 0,
				}
				mgr.GetOrCreate(info)
				return []CallInfo{info}
			},
			stmtIndex: 0,
			verifyEmission: func(t *testing.T, code string) {
				if !strings.Contains(code, "ta_sma_20") {
					t.Error("Expected ta_sma_20 emission")
				}
			},
		},
		{
			name: "different_statement_no_emission",
			setupCalls: func(mgr *TempVariableManager) []CallInfo {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 20},
					},
				}
				info := CallInfo{
					Call:      call,
					FuncName:  "ta.sma",
					ArgHash:   "test",
					StmtIndex: 0,
				}
				mgr.GetOrCreate(info)
				return []CallInfo{info}
			},
			stmtIndex: 1,
			verifyEmission: func(t *testing.T, code string) {
				if code != "" {
					t.Errorf("Expected no emission for different statement, got: %s", code)
				}
			},
		},
		{
			name: "mark_as_emitted_tracks_state",
			setupCalls: func(mgr *TempVariableManager) []CallInfo {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "ema"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 10},
					},
				}
				info := CallInfo{
					Call:      call,
					FuncName:  "ta.ema",
					ArgHash:   "test",
					StmtIndex: 0,
				}
				varName := mgr.GetOrCreate(info)

				if mgr.WasAlreadyEmitted(varName) {
					t.Error("Should not be marked as emitted before GenerateCalculationsForStatement")
				}
				return []CallInfo{info}
			},
			stmtIndex: 0,
			verifyEmission: func(t *testing.T, code string) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()
			mgr := gen.tempVarMgr

			infos := tt.setupCalls(mgr)

			code, err := mgr.GenerateCalculationsForStatement(tt.stmtIndex)
			if err != nil {
				t.Fatalf("GenerateCalculationsForStatement error: %v", err)
			}

			tt.verifyEmission(t, code)

			for _, info := range infos {
				if info.StmtIndex == tt.stmtIndex {
					varName := mgr.GetVarNameForCall(info.Call)
					if varName != "" && !mgr.WasAlreadyEmitted(varName) {
						t.Errorf("Var %s should be marked as emitted after GenerateCalculationsForStatement", varName)
					}
				}
			}
		})
	}
}

func TestGenerateVariableInit_DeduplicationIntegration(t *testing.T) {
	tests := []struct {
		name              string
		setupPreEmission  func(*generator) *ast.CallExpression
		varName           string
		initExpr          func(*ast.CallExpression) ast.Expression
		verifyNoDuplicate bool
	}{
		{
			name: "pre_emitted_ta_call_in_binary_expression",
			setupPreEmission: func(g *generator) *ast.CallExpression {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "sma"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 20},
					},
				}
				info := CallInfo{
					Call:      call,
					FuncName:  "ta.sma",
					ArgHash:   "test",
					StmtIndex: 0,
				}
				varName := g.tempVarMgr.GetOrCreate(info)
				g.tempVarMgr.emissionTracker.MarkAsEmitted(varName)
				return call
			},
			varName: "test",
			initExpr: func(call *ast.CallExpression) ast.Expression {
				return &ast.BinaryExpression{
					Left:     call,
					Operator: "+",
					Right:    &ast.Literal{Value: 10.0},
				}
			},
			verifyNoDuplicate: true,
		},
		{
			name: "pre_emitted_ta_call_in_ternary",
			setupPreEmission: func(g *generator) *ast.CallExpression {
				call := &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: "rsi"},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: 14},
					},
				}
				info := CallInfo{
					Call:      call,
					FuncName:  "ta.rsi",
					ArgHash:   "test",
					StmtIndex: 0,
				}
				varName := g.tempVarMgr.GetOrCreate(info)
				g.tempVarMgr.emissionTracker.MarkAsEmitted(varName)
				return call
			},
			varName: "signal",
			initExpr: func(call *ast.CallExpression) ast.Expression {
				return &ast.ConditionalExpression{
					Test: &ast.Identifier{Name: "condition"},
					Consequent: &ast.BinaryExpression{
						Left:     call,
						Operator: ">",
						Right:    &ast.Literal{Value: 50.0},
					},
					Alternate: &ast.Literal{Value: 0.0},
				}
			},
			verifyNoDuplicate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := createTestGenerator()
			gen.variables[tt.varName] = "float64"

			call := tt.setupPreEmission(gen)
			initExpr := tt.initExpr(call)

			code, err := gen.generateVariableInit(tt.varName, initExpr)
			if err != nil {
				t.Fatalf("generateVariableInit error: %v", err)
			}

			if tt.verifyNoDuplicate {
				inlineIndicatorCount := strings.Count(code, "/* Inline")
				if inlineIndicatorCount > 0 {
					t.Errorf("Found %d inline indicators - expected 0 (should use pre-emitted temp var)", inlineIndicatorCount)
				}

				if !strings.Contains(code, "Series") {
					t.Error("Expected reference to Series temp var, got none")
				}
			}
		})
	}
}
