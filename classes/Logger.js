class Logger {
    log(message) {
        console.log(message);
    }

    error(...args) {
        console.error(...args);
    }
}

export { Logger };