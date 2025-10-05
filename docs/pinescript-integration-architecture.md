# PineScript Integration Architecture

## Containerization

### Docker Container Configuration

This solution requires a multi-language environment to bridge PineScript parsing (Python) with Node.js execution.

#### Dockerfile
```dockerfile
FROM node:18-alpine

# Install Python and pip
RUN apk add --no-cache python3 py3-pip python3-dev build-base

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install pnpm and Node.js dependencies
RUN npm install -g pnpm
RUN pnpm install

# Install Python dependencies
COPY requirements.txt ./
RUN pip3 install -r requirements.txt

# Copy application code
COPY . .

# Expose port
EXPOSE 8080

# Default command
CMD ["pnpm", "run", "dev"]
```

#### requirements.txt
```
pynescript>=0.2.0
```

#### docker-compose.yml
```yaml
version: '3.8'

services:
  pinescript-runner:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
    command: pnpm run dev
```

### Container Build & Run
```bash
# Build container
docker build -t pinescript-runner .

# Run with volume mount for development
docker run -p 8080:8080 -v $(pwd):/app pinescript-runner

# Or use docker-compose
docker-compose up --build
```

## PineScript Import Architecture with pynescript

### Raw PineScript Input
```pinescript
//@version=5
strategy("RSI Strategy", overlay=true)
length = input( 14 )
overSold = input( 30 )
overBought = input( 70 )
price = close
vrsi = ta.rsi(price, length)
co = ta.crossover(vrsi, overSold)
cu = ta.crossunder(vrsi, overBought)
if (not na(vrsi))
	if (co)
		strategy.entry("RsiLE", strategy.long, comment="RsiLE")
	if (cu)
		strategy.entry("RsiSE", strategy.short, comment="RsiSE")
```

### Architecture Workflow

#### 1. Python Parser Service (`pinescript-parser.py`)
```python
import sys
import json
from pynescript import ast

def parse_pinescript(code):
    tree = ast.parse(code)
    
    # Extract indicators
    indicators = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            if node.func.attr == 'rsi':
                indicators.append({
                    'type': 'RSI',
                    'params': {'period': node.args[1].value}
                })
            elif node.func.attr == 'crossover':
                indicators.append({
                    'type': 'crossover',
                    'params': {'threshold': node.args[1].value}
                })
    
    return {'indicators': indicators}

if __name__ == '__main__':
    pinescript_code = sys.stdin.read()
    result = parse_pinescript(pinescript_code)
    print(json.dumps(result))
```

#### 2. Node.js Bridge (`pinescript-bridge.js`)
```javascript
const { spawn } = require('child_process');
const fs = require('fs').promises;

async function parsePineScript(pineScriptCode) {
    return new Promise((resolve, reject) => {
        const python = spawn('python3', ['pinescript-parser.py']);
        let output = '';
        
        python.stdout.on('data', (data) => {
            output += data.toString();
        });
        
        python.on('close', (code) => {
            if (code !== 0) reject(new Error('Parser failed'));
            resolve(JSON.parse(output));
        });
        
        python.stdin.write(pineScriptCode);
        python.stdin.end();
    });
}

module.exports = { parsePineScript };
```

#### 3. Updated `index.js` Integration
```javascript
const { parsePineScript } = require('./pinescript-bridge');
const PineTS = require('../PineTS');

class TechnicalAnalysisCalculator {
    async runPineScriptStrategy(pineScriptCode) {
        /* Parse PineScript to AST */
        const parsed = await parsePineScript(pineScriptCode);
        
        /* Convert to PineTS execution */
        const results = {};
        for (const indicator of parsed.indicators) {
            switch(indicator.type) {
                case 'RSI':
                    results.RSI = PineTS.Indicators.RSI(
                        this.ohlcData.map(d => d.close),
                        indicator.params.period
                    );
                    break;
                case 'crossover':
                    /* Calculate crossover signals */
                    results.signals = this.calculateCrossover(
                        results.RSI,
                        indicator.params.threshold
                    );
                    break;
            }
        }
        
        return results;
    }
}

async function main() {
    const pineScriptCode = await fs.readFile('strategy.pine', 'utf-8');
    const calculator = new TechnicalAnalysisCalculator(ohlcData);
    const results = await calculator.runPineScriptStrategy(pineScriptCode);
    
    /* Generate chart data */
    const chartData = {
        candlestick: ohlcData,
        plots: results
    };
    
    await fs.writeFile('chart-data.json', JSON.stringify(chartData));
}
```

### Execution Flow
```bash
# 1. Place PineScript strategy
echo '//@version=5...' > strategy.pine

# 2. Run analysis
pnpm run run

# 3. View chart
pnpm run dev
```

### Container-based Development Flow
```bash
# Build and run container
docker-compose up --build

# Access chart visualization
open http://localhost:8080/chart.html

# Container includes both Python (pynescript) and Node.js (PineTS)
```

### Result
Chart displays RSI indicator extracted from raw PineScript via containerized Python→Node.js bridge.

### Architecture Benefits
- **Language Isolation**: Python parsing isolated in container
- **Reproducible Environment**: Consistent Python + Node.js versions
- **Easy Deployment**: Single container with all dependencies
- **Development Flexibility**: Volume mounts for live code changes

### Next Steps
1. Create Python parser service
2. Build Node.js bridge
3. Integrate with existing PineTS workflow
4. Test with container environment