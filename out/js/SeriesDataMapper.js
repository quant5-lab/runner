/* Map indicator data to chart series format with color */
export class SeriesDataMapper {
  applyColorToData(data, color) {
    return data.map((point) => ({
      ...point,
      options: { ...point.options, color },
    }));
  }
}
