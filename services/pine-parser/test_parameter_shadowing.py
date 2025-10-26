#!/usr/bin/env python3
"""Unit tests for parameter shadowing in PyneToJsAstConverter"""
import sys
sys.path.insert(0, '/app/services/pine-parser')

from parser import PyneToJsAstConverter
from scope_chain import ScopeChain


# Parameter renaming prefix used by parser for shadowing parameters
PARAM_PREFIX = "_param_"


def test_is_shadowing_parameter_detects_global_shadowing():
    """Test _is_shadowing_parameter detects parameter shadowing global"""
    converter = PyneToJsAstConverter()
    
    # Declare global variable
    converter._scope_chain.declare("LWdilength")
    
    # Enter function scope
    converter._scope_chain.push_scope()
    
    # Check if parameter shadows global
    assert converter._is_shadowing_parameter("LWdilength")
    assert not converter._is_shadowing_parameter("other_param")
    
    converter._scope_chain.pop_scope()
    print("✅ test_is_shadowing_parameter_detects_global_shadowing")


def test_is_shadowing_parameter_no_shadowing_in_global_scope():
    """Test _is_shadowing_parameter returns False at global scope"""
    converter = PyneToJsAstConverter()
    
    # Declare at global scope
    converter._scope_chain.declare("global_var")
    
    # At global scope, no shadowing possible
    assert not converter._is_shadowing_parameter("global_var")
    
    print("✅ test_is_shadowing_parameter_no_shadowing_in_global_scope")


def test_is_shadowing_parameter_detects_parent_scope_shadowing():
    """Test _is_shadowing_parameter detects shadowing from parent function scope"""
    converter = PyneToJsAstConverter()
    
    # Global scope
    converter._scope_chain.declare("global_var")
    
    # First function scope
    converter._scope_chain.push_scope()
    converter._scope_chain.declare("func1_var")
    
    # Second function scope (nested)
    converter._scope_chain.push_scope()
    
    # Should detect both global and parent function scope
    assert converter._is_shadowing_parameter("global_var")
    assert converter._is_shadowing_parameter("func1_var")
    assert not converter._is_shadowing_parameter("new_param")
    
    converter._scope_chain.pop_scope()
    converter._scope_chain.pop_scope()
    print("✅ test_is_shadowing_parameter_detects_parent_scope_shadowing")


def test_rename_identifiers_in_ast_simple_identifier():
    """Test _rename_identifiers_in_ast renames simple identifier node"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'Identifier',
        'name': 'LWdilength'
    }
    
    renamed_param = f"{PARAM_PREFIX}LWdilength"
    param_mapping = {'LWdilength': renamed_param}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['name'] == renamed_param
    print("✅ test_rename_identifiers_in_ast_simple_identifier")


def test_rename_identifiers_in_ast_nested_structure():
    """Test _rename_identifiers_in_ast renames identifiers in nested structures"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'BinaryExpression',
        'operator': '*',
        'left': {'type': 'Identifier', 'name': 'value'},
        'right': {'type': 'Literal', 'value': 2}
    }
    
    renamed_param = f"{PARAM_PREFIX}value"
    param_mapping = {'value': renamed_param}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['left']['name'] == renamed_param
    assert node['right']['value'] == 2  # Unchanged
    print("✅ test_rename_identifiers_in_ast_nested_structure")


def test_rename_identifiers_in_ast_preserves_non_mapped_names():
    """Test _rename_identifiers_in_ast preserves identifiers not in mapping"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'BinaryExpression',
        'operator': '+',
        'left': {'type': 'Identifier', 'name': 'param1'},
        'right': {'type': 'Identifier', 'name': 'param2'}
    }
    
    renamed_param = f"{PARAM_PREFIX}param1"
    param_mapping = {'param1': renamed_param}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['left']['name'] == renamed_param
    assert node['right']['name'] == 'param2'  # Not in mapping
    print("✅ test_rename_identifiers_in_ast_preserves_non_mapped_names")


def test_rename_identifiers_in_ast_handles_arrays():
    """Test _rename_identifiers_in_ast renames identifiers in array structures"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'ArrayExpression',
        'elements': [
            {'type': 'Identifier', 'name': 'value'},
            {'type': 'Identifier', 'name': 'temp'},
            {'type': 'Identifier', 'name': 'result'}
        ]
    }
    
    renamed_value = f"{PARAM_PREFIX}value"
    renamed_temp = f"{PARAM_PREFIX}temp"
    param_mapping = {'value': renamed_value, 'temp': renamed_temp}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['elements'][0]['name'] == renamed_value
    assert node['elements'][1]['name'] == renamed_temp
    assert node['elements'][2]['name'] == 'result'  # Not in mapping
    print("✅ test_rename_identifiers_in_ast_handles_arrays")


def test_rename_identifiers_in_ast_deeply_nested():
    """Test _rename_identifiers_in_ast handles deeply nested structures"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'CallExpression',
        'callee': {'type': 'Identifier', 'name': 'ta.rma'},
        'arguments': [
            {'type': 'Identifier', 'name': 'up'},
            {
                'type': 'BinaryExpression',
                'operator': '*',
                'left': {'type': 'Identifier', 'name': 'length'},
                'right': {'type': 'Literal', 'value': 2}
            }
        ]
    }
    
    renamed_length = f"{PARAM_PREFIX}length"
    param_mapping = {'length': renamed_length}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['arguments'][0]['name'] == 'up'  # Not in mapping
    assert node['arguments'][1]['left']['name'] == renamed_length
    assert node['arguments'][1]['right']['value'] == 2
    print("✅ test_rename_identifiers_in_ast_deeply_nested")


def test_rename_identifiers_in_ast_conditional_expression():
    """Test _rename_identifiers_in_ast handles ternary/conditional expressions"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'ConditionalExpression',
        'test': {
            'type': 'BinaryExpression',
            'operator': '>',
            'left': {'type': 'Identifier', 'name': 'index'},
            'right': {'type': 'Literal', 'value': 5}
        },
        'consequent': {'type': 'Literal', 'value': 5},
        'alternate': {'type': 'Identifier', 'name': 'index'}
    }
    
    renamed_index = f"{PARAM_PREFIX}index"
    param_mapping = {'index': renamed_index}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['test']['left']['name'] == renamed_index
    assert node['alternate']['name'] == renamed_index
    assert node['consequent']['value'] == 5
    print("✅ test_rename_identifiers_in_ast_conditional_expression")


def test_rename_identifiers_in_ast_multiple_occurrences():
    """Test _rename_identifiers_in_ast renames all occurrences"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'BlockStatement',
        'body': [
            {
                'type': 'VariableDeclaration',
                'declarations': [{
                    'type': 'VariableDeclarator',
                    'id': {'type': 'Identifier', 'name': 'temp'},
                    'init': {
                        'type': 'BinaryExpression',
                        'operator': '*',
                        'left': {'type': 'Identifier', 'name': 'value'},
                        'right': {'type': 'Literal', 'value': 2}
                    }
                }]
            },
            {
                'type': 'BinaryExpression',
                'operator': '+',
                'left': {'type': 'Identifier', 'name': 'temp'},
                'right': {'type': 'Identifier', 'name': 'value'}
            }
        ]
    }
    
    renamed_value = f"{PARAM_PREFIX}value"
    param_mapping = {'value': renamed_value}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    # Check first occurrence in declaration
    assert node['body'][0]['declarations'][0]['init']['left']['name'] == renamed_value
    # Check second occurrence in expression
    assert node['body'][1]['right']['name'] == renamed_value
    # temp should be unchanged
    assert node['body'][0]['declarations'][0]['id']['name'] == 'temp'
    assert node['body'][1]['left']['name'] == 'temp'
    print("✅ test_rename_identifiers_in_ast_multiple_occurrences")


def test_rename_identifiers_in_ast_empty_mapping():
    """Test _rename_identifiers_in_ast handles empty mapping gracefully"""
    converter = PyneToJsAstConverter()
    
    node = {
        'type': 'Identifier',
        'name': 'unchanged'
    }
    
    param_mapping = {}
    
    converter._rename_identifiers_in_ast(node, param_mapping)
    
    assert node['name'] == 'unchanged'
    print("✅ test_rename_identifiers_in_ast_empty_mapping")


if __name__ == "__main__":
    test_is_shadowing_parameter_detects_global_shadowing()
    test_is_shadowing_parameter_no_shadowing_in_global_scope()
    test_is_shadowing_parameter_detects_parent_scope_shadowing()
    test_rename_identifiers_in_ast_simple_identifier()
    test_rename_identifiers_in_ast_nested_structure()
    test_rename_identifiers_in_ast_preserves_non_mapped_names()
    test_rename_identifiers_in_ast_handles_arrays()
    test_rename_identifiers_in_ast_deeply_nested()
    test_rename_identifiers_in_ast_conditional_expression()
    test_rename_identifiers_in_ast_multiple_occurrences()
    test_rename_identifiers_in_ast_empty_mapping()
    
    print("\n🎉 All 11 parameter shadowing tests passed")
