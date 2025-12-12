/* Config file loader for optional explicit pane overrides (SRP) */
export class ConfigLoader {
  static async loadStrategyConfig(strategyName) {
    try {
      const configUrl = `${strategyName}.config`;
      const response = await fetch(configUrl + '?' + Date.now());
      
      if (!response.ok) {
        return null;
      }

      const config = await response.json();
      return config.indicators || null;
    } catch (error) {
      return null;
    }
  }

  static async loadChartData(url = 'chart-data.json') {
    const response = await fetch(url + '?' + Date.now());
    return await response.json();
  }
}
