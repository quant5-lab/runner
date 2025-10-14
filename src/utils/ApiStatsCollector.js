/* API statistics collector - singleton pattern
 * Tracks API requests, cache hits/misses across strategy execution */

class ApiStatsCollector {
  static instance = null;
  static instanceCounter = 0;

  constructor() {
    if (ApiStatsCollector.instance) {
      return ApiStatsCollector.instance;
    }
    this.instanceId = ++ApiStatsCollector.instanceCounter;
    this.reset();
    ApiStatsCollector.instance = this;
  }

  reset() {
    this.stats = {
      totalRequests: 0,
      cacheHits: 0,
      cacheMisses: 0,
      byTimeframe: {},
      byProvider: {},
    };
  }

  recordRequest(provider, timeframe) {
    this.stats.totalRequests++;
    this.stats.byProvider[provider] = (this.stats.byProvider[provider] || 0) + 1;
    this.stats.byTimeframe[timeframe] = (this.stats.byTimeframe[timeframe] || 0) + 1;
  }

  recordCacheHit() {
    this.stats.cacheHits++;
  }

  recordCacheMiss() {
    this.stats.cacheMisses++;
  }

  getSummary() {
    const { totalRequests, cacheHits, cacheMisses, byTimeframe, byProvider } = this.stats;
    const totalCacheOps = cacheHits + cacheMisses;
    const cacheHitRate = totalCacheOps > 0 ? ((cacheHits / totalCacheOps) * 100).toFixed(1) : 0;

    return {
      totalRequests,
      cacheHits,
      cacheMisses,
      cacheHitRate: `${cacheHitRate}%`,
      byTimeframe,
      byProvider,
    };
  }

  logSummary(logger) {
    const { totalRequests, cacheHits, cacheMisses, cacheHitRate, byTimeframe, byProvider } =
      this.getSummary();

    const lines = [];
    lines.push('API Statistics:');
    lines.push(`Total Requests:\t${totalRequests}`);
    lines.push(`Cache Hits:\t${cacheHits}`);
    lines.push(`Cache Misses:\t${cacheMisses}`);
    lines.push(`Cache Hit Rate:\t${cacheHitRate}`);

    if (Object.keys(byTimeframe).length > 0) {
      lines.push('By Timeframe:');
      Object.entries(byTimeframe)
        .sort(([a], [b]) => a.localeCompare(b))
        .forEach(([tf, count]) => lines.push(`  ${tf}:\t${count}`));
    }

    if (Object.keys(byProvider).length > 0) {
      lines.push('By Provider:');
      Object.entries(byProvider)
        .sort(([a], [b]) => a.localeCompare(b))
        .forEach(([provider, count]) => lines.push(`  ${provider}:\t${count}`));
    }

    logger.debug(lines.join('\n'));
  }
}

export default ApiStatsCollector;
