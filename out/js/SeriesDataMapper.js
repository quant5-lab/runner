export class SeriesDataMapper {
  applyColorToData(data, color) {
    return data.map((point) => this.applyColorToPoint(point, color));
  }

  applyColorToPoint(point, defaultColor) {
    const existingColor = point.options?.color;
    const resolvedColor = this.resolvePointColor(existingColor, defaultColor);
    
    return {
      ...point,
      options: { ...point.options, color: resolvedColor },
    };
  }

  resolvePointColor(pointColor, seriesColor) {
    if (this.isExplicitGap(pointColor)) {
      return null;
    }
    return pointColor || seriesColor;
  }

  isExplicitGap(color) {
    return color === null;
  }
}
