"""
Input Function Parameter Transformer
Handles PineScript input.* defval extraction and positional argument mapping.
"""

class InputFunctionTransformer:
    INPUT_DEFVAL_FUNCTIONS = {
        'source', 'int', 'float', 'bool', 'string',
        'color', 'time', 'symbol', 'session', 'timeframe'
    }
    
    def __init__(self, estree_node_factory):
        self.estree_node = estree_node_factory
    
    def is_input_function_with_defval(self, node):
        try:
            func = node.func
            value = getattr(func, 'value', None)
            attr = getattr(func, 'attr', None)
            value_id = getattr(value, 'id', None)
            result = (value_id == 'input' and attr in self.INPUT_DEFVAL_FUNCTIONS)
        except Exception:
            result = False
        return result
    
    def transform_arguments(self, node, positional_args_js, named_args_props, visit_callback):
        if not self.is_input_function_with_defval(node):
            return positional_args_js, named_args_props, None
        
        defval_arg = None
        filtered_named_props = []
        
        for prop in named_args_props:
            if prop['key']['name'] == 'defval':
                defval_arg = prop['value']
            else:
                filtered_named_props.append(prop)
        
        final_positional_args = positional_args_js.copy()
        if defval_arg:
            final_positional_args.insert(0, defval_arg)

        return final_positional_args, filtered_named_props, defval_arg
    
    def extract_defval_from_arguments(self, args, visit_callback):
        defval_arg = None
        other_named_args = []
        
        for arg in args:
            arg_value_js = visit_callback(arg.value)
            
            if arg.name == 'defval':
                defval_arg = arg_value_js
            elif arg.name:
                prop = self.estree_node('Property',
                                       key=self.estree_node('Identifier', name=arg.name),
                                       value=arg_value_js,
                                       kind='init',
                                       method=False,
                                       shorthand=False,
                                       computed=False)
                other_named_args.append(prop)
        
        return defval_arg, other_named_args
