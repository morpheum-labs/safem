# Scaled Converter API Documentation

## Overview

The `scaled_converter.go` file provides high-performance, memory-safe number conversions for orderbook operations, risk calculations, and oracle price aggregation. It converts u256 (*big.Int) to u64 (scaled) for fast sorting and comparisons while preserving precision for exact calculations.

## Purpose

- **Orderbook Performance**: Convert prices to uint64 keys for fast sorting and map lookups
- **Risk Engine**: Convert position values to keys for aggregation and liquidation checks
- **Oracle Aggregation**: Convert price feeds to keys for efficient aggregation
- **ADL Ranking**: Convert ADL scores to keys for fast sorting
- **Overflow Protection**: Comprehensive bounds checking to prevent silent failures

## Key Concepts

### Scaled Keys

Instead of using float64 or *big.Int directly for sorting and comparisons, scaled converters use uint64 keys:

- **Original Value**: `50000.0` USD (float64 or *big.Int)
- **Scaled Key**: `5000000000000` (uint64) = 50000.0 * 1e8
- **Benefits**: Fast integer comparison, no floating-point precision issues, efficient map keys

### Scale Factors

Different scale factors for different use cases:

```go
PriceScale    = 1e8  // Prices: 8 decimal places
ValueScale    = 1e8  // Position values: 8 decimal places
QuantityScale = 1e8  // Order quantities: 8 decimal places
RatioScale    = 1e8  // Ratios: 8 decimal places (margin ratios, risk factors) - CRITICAL: Changed from 1e6 to 1e8 for unified precision
ScoreScale    = 1e8  // ADL scores: 8 decimal places
```

**IMPORTANT**: `RatioScale` was updated from `1e6` to `1e8` to align with satoshi precision across all scales. This provides 100x more precision (0.00000001 vs 0.000001) and prevents conversion errors between different scales.

### Why Scaled Keys?

1. **Performance**: Integer comparison is faster than float64
2. **Precision**: Avoids floating-point precision loss
3. **Map Keys**: uint64 is efficient for map operations
4. **Sorting**: Fast integer sorting for orderbooks
5. **Memory**: Smaller than *big.Int for hot paths

## API Reference

### Constants

```go
const (
    PriceScale    = 1e8  // 8 decimal places for prices
    ValueScale    = 1e8  // 8 decimal places for values
    QuantityScale = 1e8  // 8 decimal places for quantities
    RatioScale    = 1e6  // 6 decimal places for ratios
    ScoreScale    = 1e8  // 8 decimal places for scores
)

var (
    MaxSafePrice    = float64(math.MaxUint64) / PriceScale
    MaxSafeValue    = float64(math.MaxUint64) / ValueScale
    MaxSafeQuantity = float64(math.MaxUint64) / QuantityScale
    MaxSafeRatio    = float64(math.MaxUint64) / RatioScale
)
```

### Core Conversion Functions

#### BigIntToScaledUint64

```go
func BigIntToScaledUint64(value *big.Int, scale uint64) (uint64, error)
```

**Purpose**: Primary conversion function for u256 → u64 (scaled)

**Use When**:
- Converting order prices from EIP-712 payloads
- Converting position values for aggregation
- Converting quantities for orderbook operations

**Example**:
```go
priceBig := big.NewInt(5000000000000) // 50000.0 in satoshi
priceKey, err := BigIntToScaledUint64(priceBig, PriceScale)
if err != nil {
    // Handle overflow
}
// Returns: 5000000000000, nil
```

#### ScaledUint64ToBigInt

```go
func ScaledUint64ToBigInt(value uint64, scale uint64) *big.Int
```

**Purpose**: Reverse conversion for exact calculations

**Use When**:
- Converting keys back to exact prices for PnL calculations
- Converting keys back for precise arithmetic
- Maintaining precision in calculations

**Example**:
```go
priceKey := uint64(5000000000000)
priceBig := ScaledUint64ToBigInt(priceKey, PriceScale)
// Returns: *big.Int representing 50000.0
```

### Price Key Functions

#### BigIntToPriceKey

```go
func BigIntToPriceKey(priceBig *big.Int) (uint64, error)
```

**Purpose**: Convert order prices to keys for orderbook operations

**Use When**:
- Orderbook price level keys
- Order matching
- Depth queries

**Example**:
```go
orderPrice := big.NewInt(5000000000000) // 50000.0 USD
priceKey, err := BigIntToPriceKey(orderPrice)
if err != nil {
    // Handle error
}
// Use priceKey as map key in orderbook
```

#### Float64ToPriceKey

```go
func Float64ToPriceKey(price float64) (uint64, error)
```

**Purpose**: Convert display prices to keys

**Use When**:
- User input processing
- API requests
- Market data conversion

**Example**:
```go
userPrice := 50000.0
priceKey, err := Float64ToPriceKey(userPrice)
if err != nil {
    // Handle bounds error
}
```

#### PriceKeyToFloat64

```go
func PriceKeyToFloat64(priceKey uint64) float64
```

**Purpose**: Convert keys to display prices

**Use When**:
- API responses
- UI display
- Logging

**Warning**: May lose precision for very large values

**Example**:
```go
priceKey := uint64(5000000000000)
price := PriceKeyToFloat64(priceKey)
// Returns: 50000.0
```

#### StringToPriceKey

```go
func StringToPriceKey(priceStr string) (uint64, error)
```

**Purpose**: Convert u256 strings from EIP-712 to keys

**Use When**:
- Order submission
- Signature verification
- API parsing

**Example**:
```go
priceStr := "5000000000000" // From EIP-712 payload
priceKey, err := StringToPriceKey(priceStr)
if err != nil {
    // Handle error
}
```

### Value Key Functions

#### BigIntToValueKey

```go
func BigIntToValueKey(valueBig *big.Int) (uint64, error)
```

**Purpose**: Convert position values to keys for aggregation

**Use When**:
- ADL position calculator
- Portfolio aggregation
- Market totals

**Example**:
```go
positionValue := big.NewInt(100000000000) // $1000
valueKey, err := BigIntToValueKey(positionValue)
if err != nil {
    // Handle error
}
```

### Safe Arithmetic Operations

#### AddPriceKeys

```go
func AddPriceKeys(a, b uint64) (uint64, error)
```

**Purpose**: Safely add two price keys with overflow protection

**Use When**:
- Aggregating order quantities at price levels
- Calculating total value
- Summing quantities

**Example**:
```go
qty1 := uint64(1000000000)  // 0.01 BTC
qty2 := uint64(2000000000)  // 0.02 BTC
total, err := AddPriceKeys(qty1, qty2)
if err != nil {
    // Handle overflow
}
// Returns: 3000000000 (0.03 BTC), nil
```

#### SubtractPriceKeys

```go
func SubtractPriceKeys(a, b uint64) (uint64, error)
```

**Purpose**: Safely subtract with underflow protection

**Use When**:
- Order cancellation
- Position updates
- Delta calculations

**Example**:
```go
totalQty := uint64(100000000)   // 1.0 BTC
cancelQty := uint64(25000000)   // 0.25 BTC
remaining, err := SubtractPriceKeys(totalQty, cancelQty)
if err != nil {
    // Handle underflow
}
// Returns: 75000000 (0.75 BTC), nil
```

#### MultiplyPriceKeys

```go
func MultiplyPriceKeys(a, b uint64) (uint64, error)
```

**Purpose**: Safely multiply with overflow protection

**Use When**:
- Position value calculation (size * price)
- Margin calculations
- Value aggregations

**Example**:
```go
sizeKey, _ := Float64ToQuantityKey(0.5)  // 0.5 BTC
priceKey, _ := Float64ToPriceKey(50000.0) // $50,000
valueKey, err := MultiplyPriceKeys(sizeKey, priceKey)
if err != nil {
    // Handle overflow
}
// Returns: 2500000000000 ($25,000), nil
```

#### ComparePriceKeys

```go
func ComparePriceKeys(a, b uint64) int
```

**Purpose**: Fast integer comparison for sorting

**Returns**: -1 (a < b), 0 (a == b), 1 (a > b)

**Use When**:
- Orderbook sorting
- Price level ordering
- Ranking operations

**Example**:
```go
priceKey1 := uint64(5000000000000) // 50000.0
priceKey2 := uint64(5000100000000) // 50001.0
result := ComparePriceKeys(priceKey1, priceKey2)
// Returns: -1 (priceKey1 < priceKey2)
```

### Risk Engine Functions

#### CalculateMarginRatioKey

```go
func CalculateMarginRatioKey(equityKey, marginRequirementKey uint64) (uint64, error)
```

**Purpose**: Fast margin ratio calculation for liquidation checks

**Use When**:
- Cross-margin portfolio
- Liquidation engine
- Risk factor calculations

**Example**:
```go
equityKey, _ := Float64ToValueKey(10000.0)      // $10,000
marginKey, _ := Float64ToValueKey(5000.0)     // $5,000
ratioKey, err := CalculateMarginRatioKey(equityKey, marginKey)
if err != nil {
    // Handle error
}
ratio := RatioKeyToFloat64(ratioKey)
// Returns: 2.0 (200% margin ratio)
```

#### IsLiquidatableKey

```go
func IsLiquidatableKey(marginRatioKey uint64, liquidationThresholdKey uint64) bool
```

**Purpose**: Fast liquidation check without float64 comparisons

**Use When**:
- Liquidation engine
- Risk monitoring
- High-frequency liquidation checks

**Example**:
```go
marginRatioKey, _ := Float64ToRatioKey(1.05)    // 105%
thresholdKey, _ := Float64ToRatioKey(1.1)      // 110%
isLiquidatable := IsLiquidatableKey(marginRatioKey, thresholdKey)
// Returns: true (ratio < threshold)
```

### ADL Functions

#### Float64ToScoreKey

```go
func Float64ToScoreKey(score float64) (uint64, error)
```

**Purpose**: Convert ADL scores to keys for fast sorting

**Use When**:
- ADL ranking module
- Position prioritization
- High-frequency sorting operations

**Example**:
```go
adlScore := 1.75
scoreKey, err := Float64ToScoreKey(adlScore)
if err != nil {
    // Handle error
}
// Use scoreKey for sorting positions by ADL priority
```

### Batch Operations

#### BatchBigIntToPriceKeys

```go
func BatchBigIntToPriceKeys(prices []*big.Int) ([]uint64, error)
```

**Purpose**: Efficient bulk conversion for orderbook snapshots

**Use When**:
- Orderbook depth queries
- Market data aggregation
- Batch processing

**Example**:
```go
snapshotPrices := []*big.Int{
    big.NewInt(5000000000000),
    big.NewInt(5000100000000),
    big.NewInt(5000200000000),
}
priceKeys, err := BatchBigIntToPriceKeys(snapshotPrices)
if err != nil {
    // Handle error
}
```

## Usage Patterns

### Pattern 1: Orderbook Price Levels

```go
// Convert order prices to keys for orderbook
orderPrice := big.NewInt(5000000000000) // 50000.0 USD
priceKey, err := BigIntToPriceKey(orderPrice)
if err != nil {
    return err
}

// Use as map key in orderbook
orderbook[priceKey] = append(orderbook[priceKey], order)
```

### Pattern 2: Fast Price Sorting

```go
// Convert prices to keys
priceKeys := make([]uint64, len(prices))
for i, price := range prices {
    key, _ := Float64ToPriceKey(price)
    priceKeys[i] = key
}

// Sort using integer comparison (fast!)
sort.Slice(priceKeys, func(i, j int) bool {
    return priceKeys[i] < priceKeys[j]
})
```

### Pattern 3: Liquidation Check

```go
// Calculate margin ratio using keys
equityKey, _ := Float64ToValueKey(userEquity)
marginKey, _ := Float64ToValueKey(marginRequirement)
ratioKey, _ := CalculateMarginRatioKey(equityKey, marginKey)

// Check if liquidatable
thresholdKey, _ := Float64ToRatioKey(1.1) // 110%
if IsLiquidatableKey(ratioKey, thresholdKey) {
    triggerLiquidation(userID)
}
```

### Pattern 4: Position Value Calculation

```go
// Calculate position value: size * price
sizeKey, _ := Float64ToQuantityKey(0.5)  // 0.5 BTC
priceKey, _ := Float64ToPriceKey(50000.0) // $50,000
valueKey, err := MultiplyPriceKeys(sizeKey, priceKey)
if err != nil {
    return fmt.Errorf("overflow: %w", err)
}

// Convert back for display
positionValue := ValueKeyToFloat64(valueKey)
// Returns: 25000.0 ($25,000)
```

### Pattern 5: Order Aggregation

```go
// Aggregate orders at same price level
priceKey, _ := Float64ToPriceKey(50000.0)
totalQuantity := uint64(0)

for _, order := range ordersAtPrice {
    qtyKey, _ := Float64ToQuantityKey(order.Quantity)
    totalQuantity, _ = AddPriceKeys(totalQuantity, qtyKey)
}

// Display aggregated quantity
totalQty := ScaledUint64ToFloat64(totalQuantity, QuantityScale)
fmt.Printf("Total at $%.2f: %.8f BTC\n", PriceKeyToFloat64(priceKey), totalQty)
```

## Performance Characteristics

### Why Scaled Keys are Fast

1. **Integer Comparison**: O(1) integer comparison vs float64 precision issues
2. **Map Efficiency**: uint64 keys are efficient for Go maps
3. **Cache Locality**: Smaller values fit better in CPU cache
4. **No Allocations**: Direct uint64 operations (no heap allocations)

### Performance Comparison

| Operation | Float64 | *big.Int | Scaled uint64 |
|-----------|---------|----------|---------------|
| Comparison | Slow (precision issues) | Slow (method call) | Fast (direct) |
| Map Key | Inefficient | Inefficient | Efficient |
| Sorting | Slow (precision) | Slow (method call) | Fast (integer) |
| Memory | 8 bytes | Variable (large) | 8 bytes |

## Error Handling

### Common Errors

1. **Overflow**: `"value overflow: exceeds uint64 range"`
   - **Cause**: Value * scale exceeds uint64 max
   - **Solution**: Check bounds before conversion, use MaxSafePrice constants

2. **Underflow**: `"value underflow: negative value for unsigned type"`
   - **Cause**: Negative values not allowed for uint64
   - **Solution**: Validate input is non-negative

3. **Invalid Scale**: `"invalid scale factor"`
   - **Cause**: Scale is 0
   - **Solution**: Use predefined scale constants

4. **Out of Bounds**: `"value out of safe bounds"`
   - **Cause**: Value exceeds MaxSafePrice/MaxSafeValue
   - **Solution**: Validate before conversion

### Error Handling Best Practices

```go
// Always check bounds before conversion
if price > MaxSafePrice {
    return fmt.Errorf("price %f exceeds maximum %f", price, MaxSafePrice)
}

priceKey, err := Float64ToPriceKey(price)
if err != nil {
    return fmt.Errorf("conversion failed: %w", err)
}

// Validate after conversion
if err := ValidatePriceKey(priceKey); err != nil {
    return fmt.Errorf("invalid price key: %w", err)
}
```

## When to Use vs Other Packages

### vs satoshi.go

**Use `scaled_converter.go` when**:
- ✅ Need uint64 keys for fast sorting/comparison
- ✅ Orderbook operations (map keys, sorting)
- ✅ Risk engine calculations (liquidation checks)
- ✅ Performance-critical paths

**Use `satoshi.go` when**:
- ✅ String-based conversions (EIP-712 payloads)
- ✅ Cross-chain conversions (Wei ↔ Satoshi)
- ✅ User input/output formatting
- ✅ Payload processing

### vs safemath.go

**Use `scaled_converter.go` when**:
- ✅ High-performance orderbook operations
- ✅ Need fast sorting and comparisons
- ✅ Risk engine calculations
- ✅ Performance-critical paths

**Use `safemath.go` when**:
- ✅ Multi-token operations
- ✅ Need flexible decimal precision
- ✅ APR calculations
- ✅ String formatting

## Best Practices

### 1. Use Appropriate Scale

```go
// ❌ WRONG: Using wrong scale
ratioKey, _ := Float64ToScaledUint64(ratio, PriceScale) // Wrong scale!

// ✅ CORRECT: Use RatioScale for ratios
ratioKey, _ := Float64ToRatioKey(ratio) // Correct scale
```

### 2. Validate Bounds Before Conversion

```go
// ❌ WRONG: No bounds checking
priceKey, _ := Float64ToPriceKey(price) // May overflow

// ✅ CORRECT: Check bounds first
if price > MaxSafePrice {
    return fmt.Errorf("price too large")
}
priceKey, err := Float64ToPriceKey(price)
```

### 3. Use Batch Operations

```go
// ❌ WRONG: Individual conversions
for _, price := range prices {
    key, _ := Float64ToPriceKey(price) // Multiple validations
}

// ✅ CORRECT: Batch conversion
priceKeys, err := BatchFloat64ToPriceKeys(prices) // Single validation
```

### 4. Preserve Original *big.Int for Exact Calculations

```go
// Store both key and original value
type Order struct {
    PriceKey    uint64    // For sorting/comparison
    PriceBig    *big.Int  // For exact calculations
}

// Use key for sorting
sort.Slice(orders, func(i, j int) bool {
    return orders[i].PriceKey < orders[j].PriceKey
})

// Use *big.Int for exact PnL calculations
pnl := new(big.Int).Mul(order.PriceBig, quantityBig)
```

## Examples

See `scaled_converter_examples.go` for comprehensive examples including:
- Orderbook price level keys
- User input conversion
- Fast price comparison and sorting
- Risk engine value aggregation
- Margin ratio calculation
- Price aggregation
- Position value calculation
- Batch orderbook processing
- EIP-712 payload processing
- ADL ranking
- Order cancellation
- Bounds validation

## Thread Safety

All functions in `scaled_converter.go` are **thread-safe** and can be called concurrently. The `sync.Pool` usage is thread-safe and designed for concurrent access.

## Memory Management

### Sync.Pool Usage

The module uses `sync.Pool` to reduce allocations:

```go
// big.Int pool (reused across calls)
scaledBigIntPool.Get()    // Get from pool
defer scaledBigIntPool.Put(bigInt) // Return to pool
```

**Benefits**:
- Reduced GC pressure
- Better performance in hot paths
- Lower memory usage

## See Also

- `wei.go` - Wei (1e18) conversions for Ethereum
- `satoshi.go` - Satoshi (1e8) conversions for Morphcore
- `safemath.go` - General-purpose multi-base conversions
- `scaled_converter_examples.go` - Comprehensive usage examples

