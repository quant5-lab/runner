package datafetcher

import (
	"github.com/borisquantlab/pinescript-go/runtime/context"
)

/* DataFetcher abstracts data source for multi-timeframe fetching */
type DataFetcher interface {
	/* Fetch retrieves OHLCV bars for symbol and timeframe */
	Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error)
}
