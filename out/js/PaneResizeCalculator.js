export const MIN_PANE_HEIGHT = 80;

export class PaneResizeCalculator {
  static calculate(aboveHeight, belowHeight, delta) {
    const lo = MIN_PANE_HEIGHT - aboveHeight;
    const hi = belowHeight    - MIN_PANE_HEIGHT;
    const clamped = Math.max(lo, Math.min(hi, delta));
    return { above: aboveHeight + clamped, below: belowHeight - clamped };
  }
}
