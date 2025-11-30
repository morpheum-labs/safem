# Wei Conversion API Documentation

## Overview

The `wei.go` file provides Ethereum-specific Wei (1e18) to Ether conversion functions. Wei is the smallest unit of Ether, similar to how satoshi is the smallest unit of Bitcoin. This module is optimized for Ethereum blockchain operations and high-performance trading scenarios.

## Purpose

- **Ethereum Operations**: Convert between Wei and Ether for blockchain transactions
- **Performance Optimization**: Fast-path conversions for high-frequency operations
- **Precision Safety**: Multiple conversion strategies for different precision requirements
- **Address Interface**: Convert display values to/from blockchain format

## Key Concepts

### Wei Scale
- **Wei**: 1e18 (18 decimal places) - Ethereum's smallest unit
- **Ether**: 1.0 = 1,000,000,000,000,000,000 Wei
- **Example**: 1.5 ETH = 1,500,000,000,000,000,000 Wei

### Conversion Strategies

The module provides three conversion strategies optimized for different use cases:

1. **WeiToEther**: Balanced performance and precision (general purpose)
2. **WeiToEtherOptimized**: Maximum performance for small values (high-frequency)
3. **WeiToEtherSafe**: Maximum precision with strict limits (critical calculations)

## API Reference

### Constants

```go
const WeiScale = 1e18  // Ethereum's 18-decimal precision
```

### Functions

#### WeiToEther

```go
func WeiToEther(wei *big.Int) (float64, error)
```

**Purpose**: General-purpose Wei to Ether conversion

**Use When**:
- API responses and logging
- Address interface displays
- General calculations
- Balanced performance and precision needed

**Performance**: ~670 ns/op, 168 B/op

**Example**:
```go
balanceWei := big.NewInt(1500000000000000000) // 1.5 ETH
balanceEth, err := WeiToEther(balanceWei)
if err != nil {
    // Handle error
}
fmt.Printf("Balance: %.18f ETH\n", balanceEth)
// Output: Balance: 1.500000000000000000 ETH
```

#### WeiToEtherOptimized

```go
func WeiToEtherOptimized(wei *big.Int) (float64, error)
```

**Purpose**: High-performance conversion optimized for small values

**Use When**:
- Order book processing
- Real-time trading operations
- High-frequency market data processing
- Performance-critical paths
- Most Wei values are < 9.22e18

**Performance**: 
- Fast path: 6.5 ns/op, 0 B/op (for small values)
- Slow path: 362 ns/op, 160 B/op (for large values)

**Threshold**: 9.223372036854775807e18 Wei for fast path

**Example**:
```go
// Processing many small Wei values in orderbook
orderAmounts := []*big.Int{
    big.NewInt(100000000000000000),  // 0.1 ETH
    big.NewInt(50000000000000000),   // 0.05 ETH
}

for _, amountWei := range orderAmounts {
    amountEth, err := WeiToEtherOptimized(amountWei)
    // Fast path used for small values
}
```

#### WeiToEtherSafe

```go
func WeiToEtherSafe(wei *big.Int) (float64, error)
```

**Purpose**: Maximum precision conversion for critical financial calculations

**Use When**:
- Settlement systems
- Accounting and audit trails
- Regulatory reporting
- Compliance requirements
- Final balance calculations
- Risk management

**Performance**: ~347 ns/op (consistent regardless of input size)

**Safety Features**:
- Strict upper limit: 2^53 * 10^18 Wei (~9.007e27 Wei)
- Always uses big.Float for maximum precision
- Comprehensive error checking

**Example**:
```go
// Critical financial calculation
settlementAmountWei, _ := new(big.Int).SetString("1000000000000000000000", 10) // 1000 ETH
settlementAmountEth, err := WeiToEtherSafe(settlementAmountWei)
if err != nil {
    // Handle precision error
}
```

#### EtherToWei

```go
func EtherToWei(ether float64) (*big.Int, error)
```

**Purpose**: Convert human-readable Ether values to Wei

**Use When**:
- Address input processing
- Configuration parsing
- API request processing
- Order placement
- Deposit calculations

**Performance**: 248 B/op

**Precision Handling**:
- Tolerates small remainders (< 0.5 Wei)
- Validates NaN, Inf, negative values
- Returns error for significant precision loss

**Example**:
```go
// Address enters amount in Ether
userInputEth := 2.5
amountWei, err := EtherToWei(userInputEth)
if err != nil {
    // Handle error
}
// amountWei = 2500000000000000000 Wei
```

## Usage Patterns

### Pattern 1: Display Address Balance

```go
// Get balance from blockchain (in Wei)
balanceWei := getUserBalanceFromChain()

// Convert to Ether for display
balanceEth, err := WeiToEther(balanceWei)
if err != nil {
    log.Printf("Conversion error: %v", err)
    return
}

fmt.Printf("Your balance: %.18f ETH\n", balanceEth)
```

### Pattern 2: High-Frequency Order Processing

```go
// Processing many orders in orderbook
for _, order := range orders {
    // Use optimized version for performance
    quantityEth, err := WeiToEtherOptimized(order.QuantityWei)
    if err != nil {
        continue // Skip invalid orders
    }
    // Process order...
}
```

### Pattern 3: Critical Settlement

```go
// Final settlement calculation
totalWei, _ := new(big.Int).SetString("1000000000000000000000", 10) // 1000 ETH

// Use safe version for maximum precision
totalEth, err := WeiToEtherSafe(totalWei)
if err != nil {
    return fmt.Errorf("settlement precision error: %w", err)
}

// Use for regulatory reporting
reportSettlement(totalEth)
```

### Pattern 4: Address Input Processing

```go
// Address enters amount in UI
userInput := 1.5 // Ether

// Convert to Wei for transaction
amountWei, err := EtherToWei(userInput)
if err != nil {
    return fmt.Errorf("invalid amount: %w", err)
}

// Submit transaction
submitTransaction(amountWei)
```

## Performance Comparison

| Function | Small Values | Large Values | Memory | Use Case |
|----------|-------------|--------------|--------|----------|
| `WeiToEther` | ~670 ns/op | ~670 ns/op | 168 B/op | General purpose |
| `WeiToEtherOptimized` | 6.5 ns/op | 362 ns/op | 0-160 B/op | High-frequency |
| `WeiToEtherSafe` | ~347 ns/op | ~347 ns/op | Variable | Critical calculations |

## Error Handling

### Common Errors

1. **Nil Input**: `"wei value is nil"`
   - **Cause**: Passed nil *big.Int
   - **Solution**: Check for nil before calling

2. **Negative Value**: `"negative wei value is invalid"`
   - **Cause**: Wei values cannot be negative
   - **Solution**: Validate input before conversion

3. **Precision Loss**: `"precision loss during conversion to float64"`
   - **Cause**: Value too large or too small for float64 precision
   - **Solution**: Use `WeiToEtherSafe` or handle with big.Float

4. **Value Too Large**: `"wei value too large for precise float64 conversion"`
   - **Cause**: Exceeds 2^53 * 10^18 Wei (WeiToEtherSafe only)
   - **Solution**: Use big.Float for calculations instead

### Error Handling Best Practices

```go
// Always check errors
eth, err := WeiToEther(wei)
if err != nil {
    // Log error with context
    log.Printf("Failed to convert Wei %s: %v", wei.String(), err)
    // Return zero or handle gracefully
    return 0, err
}
```

## When to Use vs Other Packages

### vs safemath.go

**Use `wei.go` when**:
- ✅ Working exclusively with Ethereum (Wei/Ether)
- ✅ Performance is critical (trading engine, order processing)
- ✅ Need optimized fast paths for small values
- ✅ Simple Wei ↔ Ether conversions

**Use `safemath.go` when**:
- ✅ Working with multiple token types (different decimals)
- ✅ Need flexible decimal precision (6, 8, 14, 18, etc.)
- ✅ Cross-token operations
- ✅ Need APR calculations or string formatting

### vs satoshi.go

**Use `wei.go` when**:
- ✅ Ethereum blockchain operations
- ✅ Working with 18-decimal precision
- ✅ Cross-chain operations from Ethereum

**Use `satoshi.go` when**:
- ✅ Morphcore financial exchange operations
- ✅ Working with 8-decimal precision (satoshi)
- ✅ Converting from Wei to Satoshi for cross-chain

## Best Practices

### 1. Choose the Right Function

```go
// ❌ WRONG: Using general function for high-frequency operations
for _, order := range orders {
    eth, _ := WeiToEther(order.Wei) // Slow for many iterations
}

// ✅ CORRECT: Use optimized version
for _, order := range orders {
    eth, _ := WeiToEtherOptimized(order.Wei) // Fast path
}
```

### 2. Handle Errors Properly

```go
// ❌ WRONG: Ignoring errors
eth, _ := WeiToEther(wei)

// ✅ CORRECT: Check and handle errors
eth, err := WeiToEther(wei)
if err != nil {
    return fmt.Errorf("conversion failed: %w", err)
}
```

### 3. Use Appropriate Precision

```go
// ❌ WRONG: Using general function for critical calculations
settlementEth, _ := WeiToEther(settlementWei) // May lose precision

// ✅ CORRECT: Use safe version for critical operations
settlementEth, err := WeiToEtherSafe(settlementWei)
if err != nil {
    return fmt.Errorf("precision error: %w", err)
}
```

### 4. Batch Processing

```go
// Process multiple balances efficiently
balancesWei := []*big.Int{...}
for i, balanceWei := range balancesWei {
    // Use optimized version for performance
    balanceEth, err := WeiToEtherOptimized(balanceWei)
    if err != nil {
        log.Printf("Balance %d conversion error: %v", i, err)
        continue
    }
    // Process balance...
}
```

## Examples

See `wei_examples.go` for comprehensive examples including:
- Basic conversions
- High-frequency operations
- Critical financial calculations
- Address input processing
- Order submission
- Balance checking
- Gas fee calculations
- Batch processing
- Error handling

## Migration Guide

### From safemath.go

If you're currently using `safemath.BigInt2Float(wei, 18)`:

```go
// Old (safemath.go)
eth, err := safemath.BigInt2Float(wei, 18)

// New (wei.go) - Better performance for Ethereum
eth, err := WeiToEther(wei)
// Or for high-frequency:
eth, err := WeiToEtherOptimized(wei)
```

### From manual conversion

If you're manually converting:

```go
// Old (manual, error-prone)
eth := float64(wei.Int64()) / 1e18

// New (safe, validated)
eth, err := WeiToEther(wei)
if err != nil {
    // Handle error
}
```

## Thread Safety

All functions in `wei.go` are **thread-safe** and can be called concurrently from multiple goroutines. They are pure functions with no shared mutable state.

## See Also

- `satoshi.go` - Satoshi (1e8) conversions for Morphcore
- `safemath.go` - General-purpose multi-base conversions
- `scaled_converter.go` - High-performance uint64 key conversions
- `wei_examples.go` - Comprehensive usage examples

