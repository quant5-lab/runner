class Logger {
  constructor() {
    this.debugEnabled = process.env.DEBUG === 'true';
  }

  log(message) {
    console.log(message);
  }

  info(message) {
    console.log(message);
  }

  warn(message) {
    console.warn(message);
  }

  debug(message) {
    if (this.debugEnabled) {
      console.log(message);
    }
  }

  error(...args) {
    console.error(...args);
  }
}

export { Logger };
