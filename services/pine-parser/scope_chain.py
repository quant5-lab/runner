"""Scope chain with variable inheritance support for Pine Script parser"""


class ScopeChain:
    """
    Scope chain for tracking variable declarations across nested scopes.
    Supports variable lookup with inheritance from parent scopes.
    """
    
    def __init__(self):
        """Initialize with global scope"""
        self._scopes = [set()]  # Stack: [{globals}, {func1}, {func2}]
    
    def push_scope(self):
        """Enter new scope (e.g., function body)"""
        self._scopes.append(set())
    
    def pop_scope(self):
        """Exit current scope and return to parent"""
        if len(self._scopes) > 1:
            self._scopes.pop()
        else:
            raise RuntimeError("Cannot pop global scope")
    
    def declare(self, var_name):
        """Declare variable in current scope"""
        self._scopes[-1].add(var_name)
    
    def is_declared_in_current_scope(self, var_name):
        """Check if variable is declared in current (innermost) scope only"""
        return var_name in self._scopes[-1]
    
    def is_declared_in_any_scope(self, var_name):
        """Check if variable is declared in any scope (with inheritance)"""
        return any(var_name in scope for scope in self._scopes)
    
    def get_declaration_scope_level(self, var_name):
        """
        Get scope level where variable was declared.
        
        Returns:
            int: Scope level (0 = global, 1 = first function, etc.)
            None: Variable not declared in any scope
        """
        for i, scope in enumerate(self._scopes):
            if var_name in scope:
                return i
        return None
    
    def is_global(self, var_name):
        """
        Check if variable is global (declared in scope 0, accessed from nested scope).
        
        Returns True only when:
        - Variable is declared in global scope (level 0)
        - Currently in a nested scope (depth > 0)
        """
        return (var_name in self._scopes[0] and 
                len(self._scopes) > 1)
    
    def depth(self):
        """
        Get current scope depth.
        
        Returns:
            0: Global scope
            1: First function level
            2: Nested function level
            etc.
        """
        return len(self._scopes) - 1
    
    def current_scope_size(self):
        """Get number of variables in current scope"""
        return len(self._scopes[-1])
    
    def total_variables(self):
        """Get total number of unique variables across all scopes"""
        all_vars = set()
        for scope in self._scopes:
            all_vars.update(scope)
        return len(all_vars)
