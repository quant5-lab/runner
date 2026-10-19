export class ConfigLoader {
  static async loadChartData(url = 'chart-data.json') {
    const response = await fetch(url + '?' + Date.now());
    return response.json();
  }
}
