# Price feeder documentation

## Description
A side-car process that Juno validators must run in order to provide Juno's on-chain price oracle with price information. 

The ```price-feeder``` tool are based on Terra's [x/oracle](https://github.com/terra-money/classic-core/tree/main/x/oracle) module and [oracle-feeder](https://github.com/terra-money/oracle-feeder).

## Overview: 

Juno ```price-feeder``` includes 2 components:
- PriceFeeder: responsible for obtaining price information from various reliable data sources, e.g. exchanges, and exposing this data via an API.
- PriceOracle: consumes this data and periodically submits vote and prevote messages following the oracle voting procedure.

When ``PriceOracle`` start, it loops a process called ``Tick`` with 1000 milisecond interval. 

## ```Tick``` process in detail: 
1. GetParamCache: Get the latest parameters in x/oracle
2. SetPrices: Retrieve all the prices and candles from our set of providers as determined in the config. Calculate TVWAP or VWAP prices, warns missing prices, and filters out any faulty providers.    
3. Vote/Prevote exchangeRate.

## ```SetPrices``` in detail: 
1. Get all price providers
2. Collect prices from providers and set prices to local var.
3. Compute the prices and save it to ``oracle.prices``.
    - Convert any non-USD denominated candles into USD
    - Filter out any erroneous candles

    - ComputeTvwapsByProvider and ComputeTVWAP. 
    - Check if tvwapPrices available, else use recent prices & VWAP instead. 
    - ConvertTickersToUSD and FilterTickerDeviations. 
    - ComputeVWAP. 

### TVWAP calculate process: 
TVWAP is time volume weighted average price. 

**Input**: Map(provider => Map(base => CandlePrice))

**Output**: Map(base => price)

**Process**: 

For each candle within 5 timePeriod (5 mins), we calculate:

    volume = candle.Volume * (weightUnit * (period - timeDiff) + minimumTimeWeight)

    volumeSum[base] = volumeSum[base] + volume

    weightedPrices[base] = weightedPrices[base] + candle.Price * volume

where:

- weightUnit = (1 - minimumTimeWeight) / period
- timeDiff = now - candle.TimeStamp
- minimumTimeWeight = 0.2

Then use VWAP formula for `volumeSum` and `weightedPrices`

**VWAP formula**:
VWAP is volume weighted average price. 

    vwap[base] = weightedPrices[base] / volumeSum[base]

### Explain TVWAP

TWAP is the average price of a financial asset over a certain period of time. The time period is chosen by the trader based on the market analysis and trading strategy adopted. TWAPs are normally used to execute large-volume trades in smaller chunks without disturbing the market. Large-scale institutional traders track TWAP values and trade by dividing their orders into smaller parts to try and keep their orders as close to TWAP values as possible.

TWAP benefits: 
- Lower likelihood of causing asset price volatility when placing large orders
- Ability to hide your market strategy from other large-volume traders
- Good strategy for those who prefer trading by placing frequent daily orders
- Applicability to algorithmic trading strategies

TWAP limitations: 
- The TWAP formula concentrates on asset prices only and fails to take into account trading volumes.
- Limited applicability to the needs of smaller-scale traders

VWAP is a mechanism used to calculate the price of an asset by taking price data from multiple trading environments and weighting each price by the amount of volume on each liquid market an asset is trading on. The VWAP calculation methodology is also used more broadly across finance as a technical indicator for traders, an order option offered by brokers or exchanges, and a benchmark. 

VWAP benefits:
- Market Coverage.
- Highly Accurate and Fresh Data
- Manipulation-Resistant

VWAP limitations: 
- inaccurate or misleading for large orders that require many days to fill.
- can be used to manipulate trading by placing trades only when market prices are at levels favorable with VWAP
- not account for opportunity cost.

TVWAP formula is a new mechanism used to weight volume based on timestamp within the 5 min period. It's a variant of VWAP formula. The main difference between TVWAP and VWAP is that TVWAP shrinks the volume based on candle's timestamp, the further the timestamp count, the more volume will be shrink. VWAP is a measurement that shows the average price of an asset in a period of time, while TVWAP is a measurement that shows the average price of an asset in a period of time with a favor based on timestamp of data. 