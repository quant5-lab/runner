/* LineStyle converter for Lightweight Charts v4.1.1
 * Constants: 0=Solid, 1=Dotted, 2=Dashed, 3=LargeDashed, 4=SparseDotted
 */

export class LineStyleConverter {
  static SOLID = 0;
  static DOTTED = 1;
  static DASHED = 2;
  static LARGE_DASHED = 3;
  static SPARSE_DOTTED = 4;

  static toNumeric(lineStyle) {
    if (typeof lineStyle === 'number') {
      return this.validateNumeric(lineStyle);
    }

    if (typeof lineStyle === 'string') {
      return this.fromString(lineStyle);
    }

    return this.SOLID;
  }

  static fromString(styleString) {
    const normalized = styleString.toLowerCase().replace(/[-_]/g, '');
    
    switch (normalized) {
      case 'dotted':
        return this.DOTTED;
      case 'dashed':
        return this.DASHED;
      case 'largedashed':
        return this.LARGE_DASHED;
      case 'sparsedotted':
        return this.SPARSE_DOTTED;
      case 'solid':
      default:
        return this.SOLID;
    }
  }

  static validateNumeric(value) {
    return value >= 0 && value <= 4 ? value : this.SOLID;
  }
}
