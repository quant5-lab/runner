#!/usr/bin/env python3
"""Pine Script to JavaScript AST transpiler using pynescript"""
import sys
import json
from pynescript.ast import parse, dump
from pynescript.ast.grammar.asdl.generated.PinescriptASTNode import *
from input_function_transformer import InputFunctionTransformer
from scope_chain import ScopeChain


class Node:
    def __repr__(self):
        attrs = {k: v for k, v in self.__dict__.items() if not k.startswith('_')}
        return f"{self.__class__.__name__}({', '.join(f'{k}={v!r}' for k, v in attrs.items())})"


class Script(Node):
    def __init__(self, body, annotations):
        self.body = body
        self.annotations = annotations


class ReAssign(Node):
    def __init__(self, target, value):
        self.target = target
        self.value = value


class Assign(Node):
    def __init__(self, target, value, annotations, mode=None):
        self.target = target
        self.value = value
        self.annotations = annotations
        self.mode = mode


class Name(Node):
    def __init__(self, id, ctx):
        self.id = id
        self.ctx = ctx


class Constant(Node):
    def __init__(self, value, kind=None):
        self.value = value
        self.kind = kind


class BinOp(Node):
    def __init__(self, left, op, right):
        self.left = left
        self.op = op
        self.right = right


class Add: pass
class Sub: pass
class Mult: pass
class Div: pass
class Mod: pass
class Gt: pass
class Lt: pass
class Eq: pass
class GtE: pass
class LtE: pass
class NotEq: pass
class Or: pass
class And: pass
class Not: pass
class Store: pass
class Load: pass


class FunctionDef(Node):
    def __init__(self, name, args, body, method=None, export=None, annotations=[]):
        self.name = name
        self.args = args
        self.body = body
        self.method = method
        self.export = export
        self.annotations = annotations


class While(Node):
    def __init__(self, test, body=None):
        self.test = test
        self.body = body


class If(Node):
    def __init__(self, test, body, orelse):
        self.test = test
        self.body = body
        self.orelse = orelse


class ForTo(Node):
    def __init__(self, target, start, end, body):
        self.target = target
        self.start = start
        self.end = end
        self.body = body


class Param(Node):
    def __init__(self, name):
        self.name = name


class UnaryOp(Node):
    def __init__(self, op, operand):
        self.op = op
        self.operand = operand


class Call(Node):
    def __init__(self, func, args):
        self.func = func
        self.args = args


class Attribute(Node):
    def __init__(self, value, attr, ctx):
        self.value = value
        self.attr = attr
        self.ctx = ctx


class Arg(Node):
    def __init__(self, value, name=None):
        self.value = value
        self.name = name


class Expr(Node):
    def __init__(self, value):
        self.value = value


class Conditional(Node):
    def __init__(self, test, body, orelse):
        self.test = test
        self.body = body
        self.orelse = orelse


class BoolOp(Node):
    def __init__(self, op, values):
        self.op = op
        self.values = values


class Compare(Node):
    def __init__(self, left, ops, comparators):
        assert len(ops) == 1 and len(comparators) == 1
        self.left = left
        self.op = ops[0]
        self.right = comparators[0]


class Subscript(Node):
    def __init__(self, value, slice, ctx):
        self.value = value
        self.slice = slice
        self.ctx = ctx


def estree_node(type, **kwargs):
    """Create ESTree-compliant AST node"""
    node = {'type': type}
    node.update(kwargs)
    return node


class PyneToJsAstConverter:
    """Convert pynescript AST to ESTree JavaScript AST"""
    
    def __init__(self):
        self._scope_chain = ScopeChain()
        self._param_rename_stack = []

    def _is_shadowing_parameter(self, param_name):
        """Check if parameter shadows a variable in any parent scope"""
        level = self._scope_chain.get_declaration_scope_level(param_name)
        return level is not None and level < self._scope_chain.depth()

    def _rename_identifiers_in_ast(self, node, param_mapping):
        """Recursively rename identifiers in AST based on param_mapping"""
        if not param_mapping or not node:
            return node
        
        if isinstance(node, dict):
            if node.get('type') == 'Identifier' and node.get('name') in param_mapping:
                node['name'] = param_mapping[node['name']]
            
            for key, value in node.items():
                if isinstance(value, (dict, list)):
                    self._rename_identifiers_in_ast(value, param_mapping)
        
        elif isinstance(node, list):
            for item in node:
                self._rename_identifiers_in_ast(item, param_mapping)
        
        return node

    def _map_operator(self, op_node):
        if isinstance(op_node, Add): return '+'
        elif isinstance(op_node, Sub): return '-'
        elif isinstance(op_node, Mult): return '*'
        elif isinstance(op_node, Div): return '/'
        elif isinstance(op_node, Mod): return '%'
        raise NotImplementedError(f"Operator mapping not implemented for {type(op_node)}")

    def _map_comparison_operator(self, op_node):
        if isinstance(op_node, GtE): return '>='
        elif isinstance(op_node, Gt): return '>'
        elif isinstance(op_node, Lt): return '<'
        elif isinstance(op_node, LtE): return '<='
        elif isinstance(op_node, Eq): return '==='
        elif isinstance(op_node, NotEq): return '!=='
        raise NotImplementedError(f"Comparison operator mapping not implemented for {type(op_node)}")

    def _map_logical_operator(self, op_node):
        if isinstance(op_node, Or): return '||'
        elif isinstance(op_node, And): return '&&'
        raise NotImplementedError(f"Logical operator mapping not implemented for {type(op_node)}")

    def visit(self, node):
        """Visitor dispatch method"""
        method_name = 'visit_' + type(node).__name__
        visitor = getattr(self, method_name, self.generic_visit)
        return visitor(node)

    def generic_visit(self, node):
        raise NotImplementedError(f"No visit method implemented for {type(node)}")

    def visit_Script(self, node):
        body = [self.visit(stmt) for stmt in node.body]
        body = [stmt for stmt in body if stmt]
        return estree_node('Program', body=body, sourceType='module')

    def visit_Assign(self, node):
        js_value = self.visit(node.value)
        is_varip = hasattr(node, 'mode') and node.mode is not None
        
        if isinstance(node.target, Tuple):
            var_names = [elem.id for elem in node.target.elts]
            new_vars = [v for v in var_names 
                       if not self._scope_chain.is_declared_in_any_scope(v)]
            
            if new_vars:
                var_kind = 'let'
                for v in new_vars:
                    self._scope_chain.declare(v)
                declaration = estree_node('VariableDeclarator',
                                          id=self.visit(node.target),
                                          init=js_value)
                return estree_node('VariableDeclaration', declarations=[declaration], kind=var_kind)
            else:
                return estree_node('ExpressionStatement',
                                   expression=estree_node('AssignmentExpression',
                                                        operator='=',
                                                        left=self.visit(node.target),
                                                        right=js_value))
        else:
            var_name = node.target.id
            
            if not self._scope_chain.is_declared_in_any_scope(var_name):
                var_kind = 'let'
                self._scope_chain.declare(var_name)
                declaration = estree_node('VariableDeclarator',
                                          id=self.visit(node.target),
                                          init=js_value)
                return estree_node('VariableDeclaration', declarations=[declaration], kind=var_kind)
            else:
                return estree_node('ExpressionStatement',
                                   expression=estree_node('AssignmentExpression',
                                                        operator='=',
                                                        left=self.visit(node.target),
                                                        right=js_value))

    def visit_ReAssign(self, node):
        js_value = self.visit(node.value)
        
        if isinstance(node.target, Tuple):
            var_names = [elem.id for elem in node.target.elts]
            new_vars = [v for v in var_names 
                       if not self._scope_chain.is_declared_in_any_scope(v)]
            
            if new_vars:
                for v in new_vars:
                    self._scope_chain.declare(v)
                declaration = estree_node('VariableDeclarator',
                                          id=self.visit(node.target),
                                          init=js_value)
                return estree_node('VariableDeclaration', declarations=[declaration], kind='let')
            else:
                return estree_node('ExpressionStatement',
                                   expression=estree_node('AssignmentExpression',
                                                        operator='=',
                                                        left=self.visit(node.target),
                                                        right=js_value))
        else:
            var_name = node.target.id
            
            if not self._scope_chain.is_declared_in_any_scope(var_name):
                self._scope_chain.declare(var_name)
                declaration = estree_node('VariableDeclarator',
                                          id=self.visit(node.target),
                                          init=js_value)
                return estree_node('VariableDeclaration', declarations=[declaration], kind='let')
            else:
                return estree_node('ExpressionStatement',
                                   expression=estree_node('AssignmentExpression',
                                                        operator='=',
                                                        left=self.visit(node.target),
                                                        right=js_value))

    def visit_Name(self, node):
        var_name = node.id
        
        # Check parameter renaming first
        if self._param_rename_stack:
            current_mapping = self._param_rename_stack[-1]
            if var_name in current_mapping:
                return estree_node('Identifier', name=current_mapping[var_name])
        
        # Global wrapping logic: wrap globals accessed from nested scopes
        if self._scope_chain.depth() > 0:  # Inside function
            # Local variables (including renamed parameters) stay bare
            if not self._scope_chain.is_declared_in_current_scope(var_name):
                if self._scope_chain.is_global(var_name):
                    # Wrap as PineTS global: $.let.glb1_<var>
                    return estree_node('MemberExpression',
                        object=estree_node('MemberExpression',
                            object=estree_node('Identifier', name='$'),
                            property=estree_node('Identifier', name='let'),
                            computed=False
                        ),
                        property=estree_node('Identifier', name=f'glb1_{var_name}'),
                        computed=False
                    )
        
        # Bare identifier (local or at global scope)
        return estree_node('Identifier', name=var_name)

    def visit_Constant(self, node):
        return estree_node('Literal', value=node.value, raw=repr(node.value))

    def visit_BinOp(self, node):
        return estree_node('BinaryExpression',
                           operator=self._map_operator(node.op),
                           left=self.visit(node.left),
                           right=self.visit(node.right))

    def visit_UnaryOp(self, node):
        if isinstance(node.op, USub):
            operator = '-'
        elif isinstance(node.op, UAdd):
            operator = '+'
        elif isinstance(node.op, Not):
            operator = '!'
        else:
            raise NotImplementedError(f"Unary operator {type(node.op)} not implemented")
        
        return estree_node('UnaryExpression',
                           operator=operator,
                           prefix=True,
                           argument=self.visit(node.operand))

    def visit_Call(self, node):
        callee = self.visit(node.func)
        is_input_call = isinstance(node.func, Name) and node.func.id == 'input'
        
        transformer = InputFunctionTransformer(estree_node)
        is_input_with_defval = transformer.is_input_function_with_defval(node)

        positional_args_js = []
        named_args_props = []
        explicit_type_param = None

        for i, arg in enumerate(node.args):
            arg_value_js = self.visit(arg.value)
            
            # Extract explicit type parameter (e.g., type=input.float)
            if arg.name == 'type' and isinstance(arg.value, Attribute):
                if hasattr(arg.value.value, 'id') and arg.value.value.id == 'input':
                    explicit_type_param = arg.value.attr
                continue  # Skip type parameter from named args
            
            if arg.name:
                prop = estree_node('Property',
                                   key=estree_node('Identifier', name=arg.name),
                                   value=arg_value_js,
                                   kind='init',
                                   method=False,
                                   shorthand=False,
                                   computed=False)
                named_args_props.append(prop)
            else:
                positional_args_js.append(arg_value_js)

                # Type inference for input() - only if no explicit type
                if is_input_call and i == 0 and isinstance(arg.value, Constant) and not explicit_type_param:
                    first_arg_py_value = arg.value.value
                    if isinstance(first_arg_py_value, bool):
                        explicit_type_param = 'bool'
                    elif isinstance(first_arg_py_value, float):
                        explicit_type_param = 'float'
                    elif isinstance(first_arg_py_value, int):
                        explicit_type_param = 'int'

        # Transform input() to input.type() if type detected
        if is_input_call and explicit_type_param:
            callee = estree_node('MemberExpression',
                                 object=estree_node('Identifier', name='input'),
                                 property=estree_node('Identifier', name=explicit_type_param),
                                 computed=False)

        if is_input_with_defval:
            final_args_js, named_args_props, _ = transformer.transform_arguments(
                node, positional_args_js, named_args_props, self.visit
            )
        else:
            final_args_js = positional_args_js

        if named_args_props:
            options_object = estree_node('ObjectExpression', properties=named_args_props)
            final_args_js.append(options_object)

        return estree_node('CallExpression', callee=callee, arguments=final_args_js)

    def visit_Attribute(self, node):
        return estree_node('MemberExpression',
                           object=self.visit(node.value),
                           property=estree_node('Identifier', name=node.attr),
                           computed=False)

    def visit_Expr(self, node):
        if isinstance(node.value, (While, If, ForTo)):
            return self.visit(node.value)
        return estree_node('ExpressionStatement', expression=self.visit(node.value))

    def visit_Conditional(self, node):
        return estree_node('ConditionalExpression',
                           test=self.visit(node.test),
                           consequent=self.visit(node.body),
                           alternate=self.visit(node.orelse))

    def visit_BoolOp(self, node):
        if len(node.values) < 2:
            raise ValueError("BoolOp requires at least two values")

        expression = estree_node('LogicalExpression',
                                 operator=self._map_logical_operator(node.op),
                                 left=self.visit(node.values[0]),
                                 right=self.visit(node.values[1]))

        for i in range(2, len(node.values)):
            expression = estree_node('LogicalExpression',
                                     operator=self._map_logical_operator(node.op),
                                     left=expression,
                                     right=self.visit(node.values[i]))
        return expression

    def visit_Compare(self, node):
        return estree_node('BinaryExpression',
                           operator=self._map_comparison_operator(node.op),
                           left=self.visit(node.left),
                           right=self.visit(node.right))

    def visit_Subscript(self, node):
        obj = self.visit(node.value)
        prop = self.visit(node.slice)
        return estree_node('MemberExpression',
                           object=obj,
                           property=prop,
                           computed=True)

    def visit_Tuple(self, node):
        # Tuple used in assignments for array destructuring
        elements = [self.visit(elt) for elt in node.elts]
        return estree_node('ArrayPattern', elements=elements)

    def visit_FunctionDef(self, node):
        func_name = node.name
        
        # Push new function scope
        self._scope_chain.push_scope()
        
        # Build parameter mapping for shadowing parameters
        param_mapping = {}
        renamed_params = []
        
        for arg in node.args:
            original_name = arg.name
            
            if self._is_shadowing_parameter(original_name):
                # Rename shadowing parameter
                new_name = f"_param_{original_name}"
                param_mapping[original_name] = new_name
                renamed_params.append(estree_node('Identifier', name=new_name))
                self._scope_chain.declare(new_name)
            else:
                # Keep original parameter name
                renamed_params.append(self.visit(arg))
                self._scope_chain.declare(original_name)
        
        # Push param mapping for visit_Name() to use
        self._param_rename_stack.append(param_mapping)
        
        # Visit function body with param mapping active
        body_statements = [self.visit(stmt) for stmt in node.body]
        
        # Pop param mapping after body visited
        self._param_rename_stack.pop()
        
        body_block = estree_node('BlockStatement', body=body_statements)

        if body_statements and isinstance(node.body[-1], Expr):
            return_stmt = estree_node('ReturnStatement',
                                    argument=body_statements[-1].get('expression'))
            body_block['body'] = body_statements[:-1] + [return_stmt]

        # Pop function scope
        self._scope_chain.pop_scope()

        func_declaration = estree_node(
            'VariableDeclaration',
            declarations=[
                estree_node(
                    'VariableDeclarator',
                    id=estree_node('Identifier', name=func_name),
                    init=estree_node(
                        'ArrowFunctionExpression',
                        id=None,
                        params=renamed_params,
                        body=body_block,
                        expression=False,
                        generator=False,
                        **{"async": False}
                    )
                )
            ],
            kind='const'
        )

        self._scope_chain.declare(func_name, kind='const')
        return func_declaration

    def visit_Param(self, node):
        return estree_node('Identifier', name=node.name)

    def visit_While(self, node):
        test_js = self.visit(node.test)
        body_statements = [self.visit(stmt) for stmt in node.body]
        body_statements = [stmt for stmt in body_statements if stmt]
        body_block = estree_node('BlockStatement', body=body_statements)

        return {
            'type': 'WhileStatement',
            'test': test_js,
            'body': body_block
        }

    def visit_ForTo(self, node):
        var_name = node.target.id
        var_id = self.visit(node.target)
        start_js = self.visit(node.start)
        end_js = self.visit(node.end)

        init = estree_node('VariableDeclaration',
                          declarations=[
                              estree_node('VariableDeclarator',
                                         id=var_id,
                                         init=start_js)
                          ],
                          kind='let')

        self._scope_chain.declare(var_name)

        test = estree_node('BinaryExpression',
                          operator='<=',
                          left=var_id,
                          right=end_js)

        update = estree_node('UpdateExpression',
                            operator='++',
                            argument=var_id,
                            prefix=False)

        body_statements = [self.visit(stmt) for stmt in node.body]
        body_statements = [stmt for stmt in body_statements if stmt]
        body_block = estree_node('BlockStatement', body=body_statements)

        return {
            'type': 'ForStatement',
            'init': init,
            'test': test,
            'update': update,
            'body': body_block
        }

    def visit_If(self, node):
        test_js = self.visit(node.test)
        body_statements = [self.visit(stmt) for stmt in node.body]
        body_statements = [stmt for stmt in body_statements if stmt]
        consequent_block = estree_node('BlockStatement', body=body_statements)

        alternate_block = None
        if node.orelse:
            if isinstance(node.orelse, list):
                else_statements = [self.visit(stmt) for stmt in node.orelse]
                else_statements = [stmt for stmt in else_statements if stmt]
                alternate_block = estree_node('BlockStatement', body=else_statements)
            elif isinstance(node.orelse, If):
                alternate_block = self.visit(node.orelse)
            else:
                raise ValueError(f"Unexpected type for else branch: {type(node.orelse)}")

        return {
            'type': 'IfStatement',
            'test': test_js,
            'consequent': consequent_block,
            'alternate': alternate_block
        }


def main():
    """Main entry point"""
    if len(sys.argv) < 3:
        print(json.dumps({"error": "Usage: python parser.py <pine_script_file> <output_json_file>"}))
        sys.exit(1)

    filename = sys.argv[1]
    output_file = sys.argv[2]

    try:
        with open(filename, "r") as f:
            pine_code = f.read()

        tree = parse(pine_code)
        tree_dump = dump(tree, indent=2)
        
        converter = PyneToJsAstConverter()
        js_ast = converter.visit(eval(tree_dump))
        
        with open(output_file, "w") as f:
            json.dump(js_ast, f, indent=2)

    except FileNotFoundError:
        print(json.dumps({"error": f"File not found: {filename}"}))
        sys.exit(1)
    except Exception as e:
        print(json.dumps({"error": str(e), "type": type(e).__name__}))
        sys.exit(1)


if __name__ == "__main__":
    main()
