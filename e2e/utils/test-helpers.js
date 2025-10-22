/**
 * Shared constants and utilities for E2E tests
 */

/* Floating point comparison epsilon for TA function validation */
export const FLOAT_EPSILON = 0.00001;

/**
 * Assert that two floating point values are approximately equal
 * @param {number} actual - Actual value from test execution
 * @param {number} expected - Expected value from independent calculation
 * @param {number} epsilon - Maximum allowed difference (default: FLOAT_EPSILON)
 * @param {string} context - Optional context for error message
 * @throws {Error} If values differ by more than epsilon
 */
export function assertFloatEquals(actual, expected, epsilon = FLOAT_EPSILON, context = '') {
  const diff = Math.abs(actual - expected);
  if (diff > epsilon) {
    const msg = context 
      ? `${context}: Expected ${expected}, got ${actual} (diff: ${diff}, epsilon: ${epsilon})`
      : `Expected ${expected}, got ${actual} (diff: ${diff}, epsilon: ${epsilon})`;
    throw new Error(msg);
  }
}
