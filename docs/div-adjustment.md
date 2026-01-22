**YES.**

The adjustment factor is **cumulative** - each dividend/split multiplies the existing factor as you traverse backward.

```
Date        Close   Div    Factor          Adj Close
2024-12-01  300     -      1.0             300.00
2024-07-11  290     33.3   1.0 × (1-33.3/290) = 0.885   256.65
2024-01-15  280     25.0   0.885 × (1-25.0/280) = 0.806   225.68
```

Each older dividend **compounds** onto the previous factor, progressively lowering all historical prices before it.