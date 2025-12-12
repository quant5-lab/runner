package datafetcher

import (
	"github.com/quant5-lab/runner/runtime/context"
)

/* DataFetcher abstracts data source for multi-timeframe fetching */
type DataFetcher interface {
	/* Fetch retrieves OHLCV bars for symbol and timeframe */
	Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error)
}
