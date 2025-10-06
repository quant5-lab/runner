class Logger {
  constructor() {
    this.verbose = process.env.VERBOSE === 'true';
  }

  log(message) {
    console.log(message);
  }

  debug(message) {
    if (this.verbose) {
      console.log(message);
    }
  }

  error(...args) {
    console.error(...args);
  }
}

export { Logger };
