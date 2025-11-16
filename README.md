# SafeMath - Safe Arithmetic Operations for Blockchain and Financial Calculations

A high-performance Go library providing safe arithmetic operations for blockchain and financial calculations in the EngineDex system. This package is critical for handling precision-sensitive operations involving large numbers, decimal conversions, and financial calculations.

## Features

- **Safe Conversion**: Between `big.Int`, `big.Float`, and `float64` types
- **Precision-Preserving**: Arithmetic for financial calculations (Wei/Ether, token amounts)
- **Base Conversion**: Utilities for different decimal representations
- **APR Calculations**: Time-based calculations for lending/borrowing operations
- **Thread-Safe**: All functions are pure and thread-safe
- **Performance Optimized**: Caching for common operations and optimized paths

## Installation

```bash
go get github.com/morpheum-labs/safem
```

## Quick Start

### Basic Wei/Ether Conversion

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/morpheum-labs/safem"
)

func main() {
    // Convert 1.5 ETH to Wei
    wei, err := safem.EtherToWei(1.5)
    if err != nil {
        panic(err)
    }
    fmt.Printf("1.5 ETH = %s Wei\n", wei.String())

    // Convert Wei back to Ether
    eth, err := safem.WeiToEther(wei)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%s Wei = %.18f ETH\n", wei.String(), eth)
}
```

### Multi-Token Operations

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/morpheum-labs/safem"
)

func main() {
    // USDC has 6 decimals
    usdcAmount := big.NewInt(1000000) // 1 USDC
    usdcFloat, err := safem.BigInt2Float(usdcAmount, 6)
    if err != nil {
        panic(err)
    }
    fmt.Printf("USDC: %s units = %.6f USDC\n", usdcAmount.String(), usdcFloat)

    // ETH has 18 decimals
    ethAmount := big.NewInt(1500000000000000000) // 1.5 ETH
    ethFloat, err := safem.BigInt2Float(ethAmount, 18)
    if err != nil {
        panic(err)
    }
    fmt.Printf("ETH: %s units = %.18f ETH\n", ethAmount.String(), ethFloat)
}
```

## API Reference

### Core Functions

#### Wei/Ether Conversions (number.go)

**When to use**: Ethereum-specific operations, performance-critical paths

```go
// General purpose Wei to Ether conversion
func WeiToEther(wei *big.Int) (float64, error)

// High-performance Wei to Ether conversion (optimized for small values)
func WeiToEtherOptimized(wei *big.Int) (float64, error)

// Maximum precision Wei to Ether conversion (for critical calculations)
func WeiToEtherSafe(wei *big.Int) (float64, error)

// Ether to Wei conversion
func EtherToWei(ether float64) (*big.Int, error)
```

#### Multi-Base Conversions (safemath.go)

**When to use**: Multi-token operations, different decimal precisions

```go
// Convert big.Int to float64 with specified decimal places
func BigInt2Float(value *big.Int, decimal uint8) (float64, error)

// Convert float64 to big.Int with specified decimal places
func FloatToBigIntBaseX(value float64, decimal uint8) (*big.Int, error)

// Convert big.Int to string with specified decimal places
func UnBaseXFloatString(value *big.Int, decimal uint8) (string, error)

// Convert string to big.Int with specified decimal places
func BigIntByString(value string, decimal uint8) (*big.Int, error)
```

#### Scaled Uint64 Conversions (scaled_converter.go)

**When to use**: High-performance orderbook operations, risk calculations, oracle price aggregation

**Purpose**: Convert u256 (*big.Int) to u64 (scaled) for fast sorting, comparisons, and map operations while preserving precision for exact calculations.

```go
// Core conversion: u256 to u64 (scaled)
func BigIntToScaledUint64(value *big.Int, scale uint64) (uint64, error)
func ScaledUint64ToBigInt(value uint64, scale uint64) *big.Int

// Price key conversions (most common use case)
func BigIntToPriceKey(priceBig *big.Int) (uint64, error)
func PriceKeyToBigInt(priceKey uint64) *big.Int
func Float64ToPriceKey(price float64) (uint64, error)
func PriceKeyToFloat64(priceKey uint64) float64
func StringToPriceKey(priceStr string) (uint64, error)

// Value key conversions (for aggregations)
func BigIntToValueKey(valueBig *big.Int) (uint64, error)
func ValueKeyToBigInt(valueKey uint64) *big.Int
func Float64ToValueKey(value float64) (uint64, error)

// Safe arithmetic operations on keys
func AddPriceKeys(a, b uint64) (uint64, error)
func SubtractPriceKeys(a, b uint64) (uint64, error)
func MultiplyPriceKeys(a, b uint64) (uint64, error)
func ComparePriceKeys(a, b uint64) int

// Batch operations for performance
func BatchBigIntToPriceKeys(prices []*big.Int) ([]uint64, error)
func BatchPriceKeysToBigInt(keys []uint64) []*big.Int

// Advanced operations for risk calculations
func CalculateMarginRatioKey(equityKey, marginRequirementKey uint64) (uint64, error)
func IsLiquidatableKey(marginRatioKey, liquidationThresholdKey uint64) bool
```

**Scaling Constants**:
- `PriceScale = 1e8` (8 decimal places for prices)
- `ValueScale = 1e8` (8 decimal places for position values)
- `RatioScale = 1e6` (6 decimal places for ratios)
- `QuantityScale = 1e8` (8 decimal places for quantities)
- `ScoreScale = 1e8` (8 decimal places for ADL scores)

#### APR Calculations

```go
// Calculate APR for lending/borrowing operations
func CalculateAPR(principal, interest *big.Int, timeInSeconds int64) (float64, error)

// Calculate compound interest
func CalculateCompoundInterest(principal *big.Int, rate float64, timeInSeconds int64) (*big.Int, error)
```

#### BigInt Wrapper

```go
// Safe BigInt with JSON marshaling/unmarshaling
type BigInt struct {
    *big.Int
}

// Create new BigInt instances
func NewBigInt(x *big.Int) *BigInt
func NewBigIntFromString(s string, base int) (*BigInt, bool)
func NewBigIntFromInt64(x int64) *BigInt
func NewBigIntFromUint64(x uint64) *BigInt
```

## Usage Examples

### 1. Trading Engine Operations

```go
// High-frequency order processing (use number.go for performance)
func processOrder(weiAmount *big.Int) {
    // Use optimized conversion for performance
    eth, err := safem.WeiToEtherOptimized(weiAmount)
    if err != nil {
        log.Printf("Conversion error: %v", err)
        return
    }
    
    // Process order with eth value
    fmt.Printf("Processing order: %.18f ETH\n", eth)
}
```

### 2. Multi-Token DEX Operations

```go
// Handle different token types with varying decimals
func calculateSwapRate(tokenA, tokenB *big.Int, decimalsA, decimalsB uint8) (float64, error) {
    // Convert both tokens to their human-readable amounts
    amountA, err := safem.BigInt2Float(tokenA, decimalsA)
    if err != nil {
        return 0, err
    }
    
    amountB, err := safem.BigInt2Float(tokenB, decimalsB)
    if err != nil {
        return 0, err
    }
    
    // Calculate swap rate
    rate := amountB / amountA
    return rate, nil
}
```

### 3. Lending Platform Calculations

```go
// Calculate interest for lending operations
func calculateInterest(principal *big.Int, apr float64, timeInSeconds int64) (*big.Int, error) {
    // Use APR calculation function
    interest, err := safem.CalculateCompoundInterest(principal, apr, timeInSeconds)
    if err != nil {
        return nil, err
    }
    
    return interest, nil
}
```

### 4. API Response Formatting

```go
// Format token amounts for API responses
func formatTokenAmount(amount *big.Int, decimals uint8) string {
    formatted, err := safem.UnBaseXFloatString(amount, decimals)
    if err != nil {
        return "0"
    }
    return formatted
}
```

### 5. High-Performance Orderbook Operations

```go
// Convert order prices to uint64 keys for fast sorting and map lookups
func processOrderbookOrder(priceStr string) {
    // Convert u256 string (from EIP-712) to uint64 key
    priceKey, err := safem.StringToPriceKey(priceStr)
    if err != nil {
        log.Printf("Invalid price: %v", err)
        return
    }
    
    // Use priceKey for fast map lookups and sorting
    orderbookMap[priceKey] = order
    
    // For exact calculations, convert back to *big.Int
    exactPrice := safem.PriceKeyToBigInt(priceKey)
    // Perform precise arithmetic with exactPrice
}

// Batch conversion for orderbook snapshots
func getOrderbookDepth(prices []*big.Int) ([]uint64, error) {
    return safem.BatchBigIntToPriceKeys(prices)
}
```

### 6. Risk Engine Calculations

```go
// Fast liquidation checks using uint64 keys
func checkLiquidation(equity, marginRequirement *big.Int) bool {
    equityKey, _ := safem.BigIntToValueKey(equity)
    marginKey, _ := safem.BigIntToValueKey(marginRequirement)
    
    // Calculate margin ratio as uint64 key
    ratioKey, err := safem.CalculateMarginRatioKey(equityKey, marginKey)
    if err != nil {
        return false
    }
    
    // Fast comparison (no float64 precision issues)
    liquidationThreshold := uint64(1050000) // 1.05 * 1e6
    return safem.IsLiquidatableKey(ratioKey, liquidationThreshold)
}
```

### 7. Oracle Price Aggregation

```go
// Convert price feeds to keys for fast sorting and aggregation
func aggregateOraclePrices(prices []float64) (uint64, error) {
    // Convert all prices to keys
    keys, err := safem.BatchFloat64ToPriceKeys(prices)
    if err != nil {
        return 0, err
    }
    
    // Fast integer sorting (much faster than float64)
    sort.Slice(keys, func(i, j int) bool {
        return safem.ComparePriceKeys(keys[i], keys[j]) < 0
    })
    
    // Calculate median using integer arithmetic
    medianKey := keys[len(keys)/2]
    return medianKey, nil
}
```

### 8. JSON Handling

```go
// Safe JSON marshaling/unmarshaling of large numbers
type TokenBalance struct {
    Address string `json:"address"`
    Balance safem.BigInt `json:"balance"`
}

func handleTokenBalance() {
    balance := safem.BigInt{}
    balance.SetString("1000000000000000000", 10) // 1 ETH in Wei
    
    // Marshal to JSON (outputs as string)
    jsonData, _ := json.Marshal(TokenBalance{
        Address: "0x123...",
        Balance: balance,
    })
    
    // Unmarshal from JSON (handles both string and number inputs)
    var tokenBalance TokenBalance
    json.Unmarshal(jsonData, &tokenBalance)
}
```

## Performance Comparison

### number.go vs safemath.go

| Use Case | number.go | safemath.go | Recommendation |
|----------|-----------|-------------|----------------|
| Wei ↔ Ether | ✅ Optimized | ✅ Flexible | Use number.go for performance |
| Multi-token | ❌ | ✅ | Use safemath.go |
| APR calculations | ❌ | ✅ | Use safemath.go |
| String formatting | ❌ | ✅ | Use safemath.go |
| Performance | ✅ Excellent | ⚠️ Good | Use number.go for critical paths |
| Flexibility | ⚠️ Limited | ✅ High | Use safemath.go for general purpose |

### Performance Characteristics

- **WeiToEtherOptimized**: 6.5 ns/op (fast path), 362 ns/op (large values)
- **WeiToEther**: ~670 ns/op (balanced performance)
- **WeiToEtherSafe**: ~347 ns/op (maximum precision)
- **BigInt2Float**: Optimized with caching for bases 14 and 18

## Decision Matrix

### When to Use number.go
- ✅ Ethereum-specific operations (Wei/Ether only)
- ✅ Performance-critical paths (trading engines, order processing)
- ✅ High-frequency operations
- ✅ Real-time market data processing

### When to Use safemath.go
- ✅ Multi-token operations with different decimals
- ✅ Cross-chain operations
- ✅ Lending/borrowing platforms
- ✅ API responses and UI display
- ✅ Financial calculations (APR, interest rates)
- ✅ String formatting and display

## Error Handling

All functions return errors for invalid inputs. Always check errors:

```go
result, err := safem.WeiToEther(weiAmount)
if err != nil {
    // Handle error appropriately
    log.Printf("Conversion failed: %v", err)
    return
}
// Use result safely
```

## Thread Safety

- All functions are pure and thread-safe
- Cache access is not protected (but caches are read-only after initialization)
- BigInt wrapper uses sync.Pool for memory efficiency

## Precision Considerations

- **float64 precision**: Limited to ~15-17 significant digits
- **big.Int precision**: Unlimited precision
- **big.Float precision**: High precision for intermediate calculations
- **Precision loss**: Functions return errors when precision loss is detected

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support, please open an issue on GitHub or contact the development team.
