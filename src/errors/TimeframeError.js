/**
 * Custom error class for timeframe validation failures
 */
export class TimeframeError extends Error {
  constructor(timeframe, provider, supportedTimeframes) {
    const message = `${provider} does not support ${timeframe} timeframe. Supported: ${supportedTimeframes}`;
    super(message);
    this.name = 'TimeframeError';
    this.timeframe = timeframe;
    this.provider = provider;
    this.supportedTimeframes = supportedTimeframes;
  }
}
