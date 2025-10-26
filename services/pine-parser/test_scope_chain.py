#!/usr/bin/env python3
"""Unit tests for ScopeChain"""
import sys
sys.path.insert(0, '/app/services/pine-parser')

from scope_chain import ScopeChain


def test_initial_state():
    """Test initial global scope"""
    sc = ScopeChain()
    assert sc.depth() == 0
    assert sc.current_scope_size() == 0
    assert sc.total_variables() == 0
    print("✅ test_initial_state")


def test_declare_global():
    """Test declaring global variables"""
    sc = ScopeChain()
    sc.declare("global_var")
    assert sc.is_declared_in_current_scope("global_var")
    assert sc.is_declared_in_any_scope("global_var")
    assert sc.get_declaration_scope_level("global_var") == 0
    assert not sc.is_global("global_var")  # Not global when at global scope
    print("✅ test_declare_global")


def test_push_pop_scope():
    """Test scope stack operations"""
    sc = ScopeChain()
    sc.push_scope()
    assert sc.depth() == 1
    sc.push_scope()
    assert sc.depth() == 2
    sc.pop_scope()
    assert sc.depth() == 1
    sc.pop_scope()
    assert sc.depth() == 0
    print("✅ test_push_pop_scope")


def test_cannot_pop_global():
    """Test that global scope cannot be popped"""
    sc = ScopeChain()
    try:
        sc.pop_scope()
        assert False, "Should raise RuntimeError"
    except RuntimeError as e:
        assert "Cannot pop global scope" in str(e)
    print("✅ test_cannot_pop_global")


def test_variable_inheritance():
    """Test variable lookup with inheritance"""
    sc = ScopeChain()
    sc.declare("global_var")
    
    sc.push_scope()  # Enter function
    sc.declare("local_var")
    
    # Both variables visible
    assert sc.is_declared_in_any_scope("global_var")
    assert sc.is_declared_in_any_scope("local_var")
    
    # Only local in current scope
    assert not sc.is_declared_in_current_scope("global_var")
    assert sc.is_declared_in_current_scope("local_var")
    
    print("✅ test_variable_inheritance")


def test_global_detection():
    """Test is_global() detection"""
    sc = ScopeChain()
    sc.declare("global_var")
    
    # Not global when at global scope
    assert not sc.is_global("global_var")
    
    sc.push_scope()  # Enter function
    
    # Now it's global (in scope 0, accessed from scope 1)
    assert sc.is_global("global_var")
    
    sc.declare("local_var")
    assert not sc.is_global("local_var")
    
    print("✅ test_global_detection")


def test_scope_levels():
    """Test get_declaration_scope_level()"""
    sc = ScopeChain()
    sc.declare("global_var")
    
    sc.push_scope()  # Scope 1
    sc.declare("func1_var")
    
    sc.push_scope()  # Scope 2
    sc.declare("func2_var")
    
    assert sc.get_declaration_scope_level("global_var") == 0
    assert sc.get_declaration_scope_level("func1_var") == 1
    assert sc.get_declaration_scope_level("func2_var") == 2
    assert sc.get_declaration_scope_level("nonexistent") is None
    
    print("✅ test_scope_levels")


def test_nested_functions():
    """Test nested function scopes (realistic scenario)"""
    sc = ScopeChain()
    sc.declare("global_var")
    
    sc.push_scope()  # outer_func
    sc.declare("x")  # parameter
    sc.declare("local_outer")
    
    sc.push_scope()  # inner_func
    sc.declare("y")  # parameter
    sc.declare("result")
    
    # All variables accessible
    assert sc.is_declared_in_any_scope("global_var")
    assert sc.is_declared_in_any_scope("x")
    assert sc.is_declared_in_any_scope("local_outer")
    assert sc.is_declared_in_any_scope("y")
    assert sc.is_declared_in_any_scope("result")
    
    # Global detection
    assert sc.is_global("global_var")
    assert not sc.is_global("x")
    assert not sc.is_global("local_outer")
    assert not sc.is_global("y")
    assert not sc.is_global("result")
    
    # Scope levels
    assert sc.get_declaration_scope_level("global_var") == 0
    assert sc.get_declaration_scope_level("x") == 1
    assert sc.get_declaration_scope_level("local_outer") == 1
    assert sc.get_declaration_scope_level("y") == 2
    assert sc.get_declaration_scope_level("result") == 2
    
    sc.pop_scope()  # Exit inner_func
    sc.pop_scope()  # Exit outer_func
    
    assert sc.depth() == 0
    print("✅ test_nested_functions")


def test_variable_shadowing():
    """Test variable shadowing across scopes"""
    sc = ScopeChain()
    sc.declare("x")
    
    sc.push_scope()
    sc.declare("x")  # Shadow global x
    
    # Both exist in different scopes
    assert sc.get_declaration_scope_level("x") == 0  # Returns first found
    assert sc.is_declared_in_current_scope("x")
    assert sc.is_declared_in_any_scope("x")
    
    print("✅ test_variable_shadowing")


def test_total_variables():
    """Test total_variables() count"""
    sc = ScopeChain()
    sc.declare("a")
    sc.declare("b")
    
    sc.push_scope()
    sc.declare("c")
    sc.declare("d")
    
    assert sc.total_variables() == 4
    assert sc.current_scope_size() == 2
    
    print("✅ test_total_variables")


if __name__ == "__main__":
    test_initial_state()
    test_declare_global()
    test_push_pop_scope()
    test_cannot_pop_global()
    test_variable_inheritance()
    test_global_detection()
    test_scope_levels()
    test_nested_functions()
    test_variable_shadowing()
    test_total_variables()
    
    print("\n🎉 All 10 tests passed")
