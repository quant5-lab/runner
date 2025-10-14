/* Pine Script v3/v4 to v5 auto-migrator
 * Transforms v3/v4 syntax to v5 before transpilation
 * Based on: https://www.tradingview.com/pine-script-docs/migration-guides/to-pine-version-5/ */

import TickeridMigrator from '../utils/tickeridMigrator.js';

class PineVersionMigrator {
  static V5_MAPPINGS = {
    // No namespace changes
    study: 'indicator',
    'tickerid()': 'ticker.new()',

    // Color constants in assignments (color=yellow → color=color.yellow)
    '=\\s*yellow\\b': '=color.yellow',
    '=\\s*green\\b': '=color.green',
    '=\\s*red\\b': '=color.red',
    '=\\s*blue\\b': '=color.blue',
    '=\\s*white\\b': '=color.white',
    '=\\s*black\\b': '=color.black',
    '=\\s*gray\\b': '=color.gray',
    '=\\s*orange\\b': '=color.orange',
    '=\\s*aqua\\b': '=color.aqua',
    '=\\s*fuchsia\\b': '=color.fuchsia',
    '=\\s*lime\\b': '=color.lime',
    '=\\s*maroon\\b': '=color.maroon',
    '=\\s*navy\\b': '=color.navy',
    '=\\s*olive\\b': '=color.olive',
    '=\\s*purple\\b': '=color.purple',
    '=\\s*silver\\b': '=color.silver',
    '=\\s*teal\\b': '=color.teal',

    // ta.* namespace
    accdist: 'ta.accdist',
    'alma(': 'ta.alma(',
    'atr(': 'ta.atr(',
    'bb(': 'ta.bb(',
    'bbw(': 'ta.bbw(',
    'cci(': 'ta.cci(',
    'cmo(': 'ta.cmo(',
    'cog(': 'ta.cog(',
    'dmi(': 'ta.dmi(',
    'ema(': 'ta.ema(',
    'hma(': 'ta.hma(',
    iii: 'ta.iii',
    'kc(': 'ta.kc(',
    'kcw(': 'ta.kcw(',
    'linreg(': 'ta.linreg(',
    'macd(': 'ta.macd(',
    'mfi(': 'ta.mfi(',
    'mom(': 'ta.mom(',
    nvi: 'ta.nvi',
    obv: 'ta.obv',
    pvi: 'ta.pvi',
    pvt: 'ta.pvt',
    'rma(': 'ta.rma(',
    'roc(': 'ta.roc(',
    'rsi(': 'ta.rsi(',
    'sar(': 'ta.sar(',
    'sma(': 'ta.sma(',
    'stoch(': 'ta.stoch(',
    'supertrend(': 'ta.supertrend(',
    'swma(': 'ta.swma(',
    'tr(': 'ta.tr(',
    'tsi(': 'ta.tsi(',
    vwap: 'ta.vwap',
    'vwma(': 'ta.vwma(',
    wad: 'ta.wad',
    'wma(': 'ta.wma(',
    'wpr(': 'ta.wpr(',
    wvad: 'ta.wvad',
    'barsince(': 'ta.barsince(',
    'change(': 'ta.change(',
    'correlation(': 'ta.correlation(',
    'cross(': 'ta.cross(',
    'crossover(': 'ta.crossover(',
    'crossunder(': 'ta.crossunder(',
    'cum(': 'ta.cum(',
    'dev(': 'ta.dev(',
    'falling(': 'ta.falling(',
    'highest(': 'ta.highest(',
    'highestbars(': 'ta.highestbars(',
    'lowest(': 'ta.lowest(',
    'lowestbars(': 'ta.lowestbars(',
    'median(': 'ta.median(',
    'mode(': 'ta.mode(',
    'percentile_linear_interpolation(': 'ta.percentile_linear_interpolation(',
    'percentile_nearest_rank(': 'ta.percentile_nearest_rank(',
    'percentrank(': 'ta.percentrank(',
    'pivothigh(': 'ta.pivothigh(',
    'pivotlow(': 'ta.pivotlow(',
    'range(': 'ta.range(',
    'rising(': 'ta.rising(',
    'stdev(': 'ta.stdev(',
    'valuewhen(': 'ta.valuewhen(',
    'variance(': 'ta.variance(',

    // math.* namespace
    'abs(': 'math.abs(',
    'acos(': 'math.acos(',
    'asin(': 'math.asin(',
    'atan(': 'math.atan(',
    'avg(': 'math.avg(',
    'ceil(': 'math.ceil(',
    'cos(': 'math.cos(',
    'exp(': 'math.exp(',
    'floor(': 'math.floor(',
    'log(': 'math.log(',
    'log10(': 'math.log10(',
    'max(': 'math.max(',
    'min(': 'math.min(',
    'pow(': 'math.pow(',
    'random(': 'math.random(',
    'round(': 'math.round(',
    'round_to_mintick(': 'math.round_to_mintick(',
    'sign(': 'math.sign(',
    'sin(': 'math.sin(',
    'sqrt(': 'math.sqrt(',
    'sum(': 'math.sum(',
    'tan(': 'math.tan(',
    'todegrees(': 'math.todegrees(',
    'toradians(': 'math.toradians(',

    // request.* namespace
    'financial(': 'request.financial(',
    'quandl(': 'request.quandl(',
    'security(': 'request.security(',
    'splits(': 'request.splits(',
    'dividends(': 'request.dividends(',
    'earnings(': 'request.earnings(',

    // ticker.* namespace
    'heikinashi(': 'ticker.heikinashi(',
    'kagi(': 'ticker.kagi(',
    'linebreak(': 'ticker.linebreak(',
    'pointfigure(': 'ticker.pointfigure(',
    'renko(': 'ticker.renko(',

    // str.* namespace
    'tostring(': 'str.tostring(',
    'tonumber(': 'str.tonumber(',
  };

  static needsMigration(pineCode, version) {
    return version === null || version < 5;
  }

  static hasV3V4Syntax(pineCode) {
    /* Detect v3/v4 syntax patterns that need migration */
    return /\b(study|(?<!ta\.|request\.|math\.|ticker\.|str\.)(?:sma|ema|rsi|security))\s*\(/.test(pineCode);
  }

  static migrate(pineCode, version) {
    if (!this.needsMigration(pineCode, version)) {
      return pineCode;
    }

    let migrated = pineCode;

    /* Migrate tickerid references first (handles all variants) */
    migrated = TickeridMigrator.migrate(migrated);

    /* Apply function patterns first (longer patterns), then simple identifiers */
    const functionPatterns = [];
    const identifierPatterns = [];

    for (const [v4Pattern, v5Replacement] of Object.entries(this.V5_MAPPINGS)) {
      if (v4Pattern.includes('(')) {
        functionPatterns.push([v4Pattern, v5Replacement]);
      } else {
        identifierPatterns.push([v4Pattern, v5Replacement]);
      }
    }

    /* Process function calls first to avoid partial matches */
    for (const [v4Pattern, v5Replacement] of functionPatterns) {
      const regex = new RegExp(this.escapeRegex(v4Pattern), 'g');
      migrated = migrated.replace(regex, v5Replacement);
    }

    /* Then process identifiers and regex patterns */
    for (const [v4Pattern, v5Replacement] of identifierPatterns) {
      const isRegexPattern = v4Pattern.includes('\\');

      if (isRegexPattern) {
        const regex = new RegExp(v4Pattern, 'g');
        migrated = migrated.replace(regex, v5Replacement);
      } else {
        const regex = new RegExp(`\\b${this.escapeRegex(v4Pattern)}\\b`, 'g');
        migrated = migrated.replace(regex, v5Replacement);
      }
    }

    return migrated;
  }

  static escapeRegex(str) {
    return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }
}

export default PineVersionMigrator;
