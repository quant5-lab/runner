/* Migrates v3/v4 tickerid references to v5 syminfo.tickerid
 * Handles all valid PineScript v3/v4 usage patterns */

class TickeridMigrator {
  static migrate(code) {
    /* Pattern 1: standalone tickerid variable (not syminfo.tickerid, not tickerid()) */
    const standalonePattern = /(?<!syminfo\.)(?<!\.)\btickerid\b(?!\()/g;
    code = code.replace(standalonePattern, 'syminfo.tickerid');
    
    /* Pattern 2: tickerId (camelCase variant) */
    const camelCasePattern = /(?<!syminfo\.)(?<!\.)\btickerId\b(?!\()/g;
    code = code.replace(camelCasePattern, 'syminfo.tickerid');
    
    /* Pattern 3: tickerid() function call (preserves spaces inside parens) */
    code = code.replace(/\btickerid\((\s*)\)/g, 'ticker.new($1)');
    
    return code;
  }
}

export default TickeridMigrator;
