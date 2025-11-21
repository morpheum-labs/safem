# Satoshi Conversion API Documentation

## Overview

The `satoshi.go` file provides Morphcore-specific Satoshi (1e8) conversion functions for financial exchange operations. Satoshi follows Bitcoin's naming convention (8 decimal places) and is used throughout Morphcore for prices, quantities, and amounts. This module also provides cross-chain conversion between Ethereum Wei (1e18) and Morphcore Satoshi (1e8).

## Purpose

- **Morphcore Operations**: Convert between decimal values and Satoshi format (1e8)
- **Cross-Chain Operations**: Convert between Ethereum Wei and Morphcore Satoshi
- **Payload Processing**: Format values for EIP-712 signature payloads
- **High-Performance**: Memory-safe operations using sync.Pool for allocation reuse

## Key Concepts

### Satoshi Scale
- **Satoshi**: 1e8 (8 decimal places) - Morphcore's precision standard
- **Decimal**: 1.0 = 100,000,000 Satoshi
- **Example**: 50,000.0 USD = 5,000,000,000,000 Satoshi

### Wei vs Satoshi
- **Wei (Ethereum)**: 1e18 (18 decimals) - Used for blockchain transactions
- **Satoshi (Morphcore)**: 1e8 (8 decimals) - Used for exchange operations
- **Conversion**: 1 Wei = 1e10 Satoshi (divide Wei by 1e10 to get Satoshi)

### Why 1e8 Instead of 1e18?

Morphcore uses 1e8 (satoshi) instead of 1e18 (wei) for several reasons:

1. **Int64 Overflow Prevention**: 
   - 50,000.00 * 1e18 = 48.1 quintillion (overflows int64)
   - 50,000.00 * 1e8 = 4.8 trillion (fits in int64)

2. **Sufficient Precision**: 
   - 8 decimals = $0.00000001 precision
   - More than adequate for any financial instrument

3. **Performance**: 
   - Faster operations with smaller numbers
   - Better cache locality
   - Reduced memory usage

## API Reference

### Constants

```go
const SatoshiScale = 1e8   // Morphcore's 8-decimal precision
const WeiScale = 1e18      // Ethereum's 18-decimal precision
```

### Core Conversion Functions

#### DecimalToSatoshi

```go
func DecimalToSatoshi(decimal float64) string
```

**Purpose**: Convert decimal value to satoshi format string

**Use When**:
- Order submission payloads
- EIP-712 signature construction
- User input processing
- Price/quantity formatting

**Performance**: Fast path for small values, uses big.Float for precision

**Example**:
```go
price := 50000.0
priceSatoshi := DecimalToSatoshi(price)
// Returns: "5000000000000"
```

#### SatoshiToDecimal

```go
func SatoshiToDecimal(satoshiStr string) (float64, error)
```

**Purpose**: Convert satoshi string back to decimal for display

**Use When**:
- API responses
- UI display
- Logging
- User-facing values

**Example**:
```go
satoshiStr := "5000000000000"
price, err := SatoshiToDecimal(satoshiStr)
if err != nil {
    // Handle error
}
// Returns: 50000.0, nil
```

#### DecimalToSatoshiBigInt

```go
func DecimalToSatoshiBigInt(decimal float64) (*big.Int, error)
```

**Purpose**: High-precision conversion returning *big.Int

**Use When**:
- Critical financial calculations
- PnL calculations
- Margin requirements
- Exact arithmetic needed

**Example**:
```go
amount := 12345.67890123
amountSatoshi, err := DecimalToSatoshiBigInt(amount)
if err != nil {
    // Handle error
}
// Returns: *big.Int("1234567890123"), nil
```

#### SatoshiToDecimalBigFloat

```go
func SatoshiToDecimalBigFloat(satoshiStr string) (*big.Float, error)
```

**Purpose**: Maximum precision conversion for audit trails

**Use When**:
- Exact financial calculations
- Audit trails
- Regulatory reporting
- Maximum precision required

**Example**:
```go
satoshiStr := "12345678901234567890"
decimal, err := SatoshiToDecimalBigFloat(satoshiStr)
if err != nil {
    // Handle error
}
// Returns: *big.Float with maximum precision
```

### Cross-Chain Conversion Functions

#### WeiToSatoshi

```go
func WeiToSatoshi(weiStr string) (string, error)
```

**Purpose**: Convert Ethereum Wei to Morphcore Satoshi

**Use When**:
- Cross-chain token transfers
- Bridge operations
- Converting Ethereum amounts to Morphcore format

**Conversion**: Divides by 1e10 (1e18 / 1e8 = 1e10)

**Example**:
```go
ethereumAmountWei := "50000000000000000000000" // 50,000 tokens
morphcoreAmountSatoshi, err := WeiToSatoshi(ethereumAmountWei)
if err != nil {
    // Handle error
}
// Returns: "5000000000000", nil
```

#### SatoshiToWei

```go
func SatoshiToWei(satoshiStr string) (string, error)
```

**Purpose**: Convert Morphcore Satoshi to Ethereum Wei

**Use When**:
- Withdrawing from Morphcore to Ethereum
- Bridge operations
- Converting Morphcore amounts to Ethereum format

**Conversion**: Multiplies by 1e10 (1e8 * 1e10 = 1e18)

**Example**:
```go
morphcoreAmountSatoshi := "5000000000000" // 50,000 tokens
ethereumAmountWei, err := SatoshiToWei(morphcoreAmountSatoshi)
if err != nil {
    // Handle error
}
// Returns: "50000000000000000000000", nil
```

### Smart Normalization Function

#### NormalizeToSatoshi

```go
func NormalizeToSatoshi(value interface{}) (string, error)
```

**Purpose**: Universal converter that accepts multiple input formats

**Use When**:
- Payload processing with unknown input format
- API parsing
- User input handling
- Auto-detecting Wei vs Satoshi format

**Accepts**:
- Decimal strings: `"50000.0"` → `"5000000000000"`
- Float64: `50000.0` → `"5000000000000"`
- Wei strings (1e18): `"50000000000000000000000"` → `"5000000000000"` (auto-detected)
- Already satoshi (1e8): `"5000000000000"` → `"5000000000000"` (no change)
- *big.Int: Automatically detects format

**Example**:
```go
// Various input formats
inputs := []interface{}{
    "50000.0",                        // Decimal string
    50000.0,                          // Float64
    "5000000000000",                  // Already satoshi
    "50000000000000000000000",        // Wei format (auto-detected)
}

for _, input := range inputs {
    satoshi, err := NormalizeToSatoshi(input)
    // All normalize to: "5000000000000"
}
```

### Batch Operations

#### BatchDecimalToSatoshi

```go
func BatchDecimalToSatoshi(decimals []float64) []string
```

**Purpose**: Efficient bulk conversion for orderbook operations

**Use When**:
- Orderbook snapshots
- Batch processing
- Market data aggregation

**Example**:
```go
prices := []float64{50000.0, 50001.0, 50002.0}
pricesSatoshi := BatchDecimalToSatoshi(prices)
// Returns: ["5000000000000", "5000100000000", "5000200000000"]
```

#### BatchSatoshiToDecimal

```go
func BatchSatoshiToDecimal(satoshiStrs []string) ([]float64, error)
```

**Purpose**: Bulk conversion for API responses

**Use When**:
- Orderbook depth queries
- Batch display formatting
- API response preparation

**Example**:
```go
satoshiValues := []string{"5000000000000", "10000000", "250000000000"}
decimalValues, err := BatchSatoshiToDecimal(satoshiValues)
if err != nil {
    // Handle error
}
// Returns: [50000.0, 0.1, 2500.0], nil
```

## Usage Patterns

### Pattern 1: Order Submission

```go
// User input from UI
orderPrice := 50000.0   // $50,000
orderQuantity := 0.1    // 0.1 BTC

// Convert to satoshi for EIP-712 payload
priceSatoshi := DecimalToSatoshi(orderPrice)
quantitySatoshi := DecimalToSatoshi(orderQuantity)

// Build payload
payload := map[string]interface{}{
    "price":    priceSatoshi,    // "5000000000000"
    "quantity": quantitySatoshi,  // "10000000"
}
```

### Pattern 2: Cross-Chain Token Bridge

```go
// User wants to bridge tokens from Ethereum
ethereumBalanceWei := "1000000000000000000000" // 1000 tokens

// Convert to Morphcore format
morphcoreBalanceSatoshi, err := WeiToSatoshi(ethereumBalanceWei)
if err != nil {
    return fmt.Errorf("bridge conversion error: %w", err)
}

// Use in Morphcore operations
processMorphcoreTransaction(morphcoreBalanceSatoshi)
```

### Pattern 3: Universal Input Normalization

```go
// Payload data from different sources
payloadData := map[string]interface{}{
    "price":    "50000.0",              // User input (decimal string)
    "quantity": 0.1,                    // User input (float64)
    "amount":   "5000000000000",        // Already satoshi
    "fee":      "500000000000000000000", // Wei format (from Ethereum)
}

// Normalize all values to satoshi
normalizedPayload := make(map[string]string)
for key, value := range payloadData {
    satoshi, err := NormalizeToSatoshi(value)
    if err != nil {
        return fmt.Errorf("error normalizing %s: %w", key, err)
    }
    normalizedPayload[key] = satoshi
}
```

### Pattern 4: Display Formatting

```go
// Values from API in satoshi format
priceSatoshi := "5000000000000"    // 50,000 USD
quantitySatoshi := "10000000"      // 0.1 BTC

// Convert to decimal for display
priceDecimal, err := SatoshiToDecimal(priceSatoshi)
if err != nil {
    // Handle error
}

quantityDecimal, err := SatoshiToDecimal(quantitySatoshi)
if err != nil {
    // Handle error
}

// Display to user
fmt.Printf("Price: $%.2f\n", priceDecimal)        // $50000.00
fmt.Printf("Quantity: %.8f BTC\n", quantityDecimal) // 0.10000000 BTC
```

## Performance Characteristics

### Memory Safety

All functions use `sync.Pool` for big.Int and big.Float reuse, reducing allocations in hot paths:

```go
// Pooled allocations (reused across calls)
satoshiBigIntPool    // For big.Int operations
satoshiBigFloatPool  // For big.Float operations
```

### Fast Paths

- **DecimalToSatoshi**: Fast path for values < MaxSafeSatoshi (fits in int64)
- **SatoshiToDecimal**: Fast path for values that fit in int64
- **NormalizeToSatoshi**: Auto-detects format to avoid unnecessary conversions

## Error Handling

### Common Errors

1. **Invalid Format**: `"invalid satoshi format"`
   - **Cause**: Invalid string format or empty string
   - **Solution**: Validate input before conversion

2. **Overflow**: `"satoshi value overflow: exceeds int64 range"`
   - **Cause**: Value too large for int64 conversion
   - **Solution**: Use *big.Int versions for large values

3. **Underflow**: `"satoshi value underflow: negative value"`
   - **Cause**: Negative values not allowed
   - **Solution**: Validate input is non-negative

### Error Handling Best Practices

```go
// Always check errors
satoshi, err := NormalizeToSatoshi(input)
if err != nil {
    return fmt.Errorf("failed to normalize %v: %w", input, err)
}

// Use appropriate precision level
if isCriticalCalculation {
    amountBig, err := DecimalToSatoshiBigInt(amount)
    if err != nil {
        return fmt.Errorf("precision error: %w", err)
    }
    // Use *big.Int for exact calculations
}
```

## When to Use vs Other Packages

### vs wei.go

**Use `satoshi.go` when**:
- ✅ Morphcore financial exchange operations
- ✅ Working with 8-decimal precision (satoshi)
- ✅ Order submission and payload processing
- ✅ Cross-chain operations (Wei ↔ Satoshi)

**Use `wei.go` when**:
- ✅ Ethereum blockchain operations
- ✅ Working with 18-decimal precision (wei)
- ✅ Pure Ethereum operations (no Morphcore)

### vs safemath.go

**Use `satoshi.go` when**:
- ✅ Morphcore-specific operations (8 decimals)
- ✅ Need Wei ↔ Satoshi cross-chain conversion
- ✅ Payload processing for EIP-712
- ✅ Memory-efficient operations (sync.Pool)

**Use `safemath.go` when**:
- ✅ Multi-token operations (different decimals)
- ✅ Need flexible decimal precision
- ✅ APR calculations or string formatting

### vs scaled_converter.go

**Use `satoshi.go` when**:
- ✅ String-based conversions (EIP-712 payloads)
- ✅ Cross-chain conversions
- ✅ User input/output formatting

**Use `scaled_converter.go` when**:
- ✅ High-performance orderbook operations
- ✅ Need uint64 keys for fast sorting
- ✅ Risk engine calculations
- ✅ Performance-critical paths

## Best Practices

### 1. Use NormalizeToSatoshi for Unknown Inputs

```go
// ❌ WRONG: Assuming input format
priceSatoshi := DecimalToSatoshi(input) // Fails if input is already satoshi

// ✅ CORRECT: Normalize unknown inputs
priceSatoshi, err := NormalizeToSatoshi(input) // Handles all formats
```

### 2. Choose Appropriate Precision Level

```go
// ❌ WRONG: Using string for critical calculations
amountStr := DecimalToSatoshi(amount)
// Loses precision in string operations

// ✅ CORRECT: Use *big.Int for exact calculations
amountBig, err := DecimalToSatoshiBigInt(amount)
// Preserves exact precision
```

### 3. Batch Operations for Performance

```go
// ❌ WRONG: Individual conversions
for _, price := range prices {
    satoshi := DecimalToSatoshi(price) // Multiple allocations
}

// ✅ CORRECT: Batch conversion
satoshiValues := BatchDecimalToSatoshi(prices) // Single pass
```

### 4. Handle Cross-Chain Conversions Properly

```go
// ❌ WRONG: Manual conversion (error-prone)
satoshi := wei / 1e10 // May lose precision

// ✅ CORRECT: Use dedicated function
satoshi, err := WeiToSatoshi(weiStr) // Safe, validated conversion
```

## Examples

See `satoshi_examples.go` for comprehensive examples including:
- Order submission with satoshi format
- Display formatting
- Cross-chain Wei ↔ Satoshi conversions
- Universal input normalization
- High-precision conversions
- Batch processing
- Payload processing
- Token bridge operations
- Error handling

## Memory Management

### Sync.Pool Usage

The module uses `sync.Pool` to reduce allocations:

```go
// big.Int pool (reused across calls)
satoshiBigIntPool.Get()    // Get from pool
defer satoshiBigIntPool.Put(bigInt) // Return to pool

// big.Float pool (reused across calls)
satoshiBigFloatPool.Get()    // Get from pool
defer satoshiBigFloatPool.Put(bigFloat) // Return to pool
```

**Benefits**:
- Reduced GC pressure
- Better performance in hot paths
- Lower memory usage

## Thread Safety

All functions in `satoshi.go` are **thread-safe** and can be called concurrently. The `sync.Pool` usage is thread-safe and designed for concurrent access.

## Migration Guide

### From Manual Conversion

If you're manually converting:

```go
// Old (manual, error-prone)
satoshi := int64(price * 1e8)

// New (safe, validated)
satoshi := DecimalToSatoshi(price)
```

### From Wei Format

If you have Wei values and need Satoshi:

```go
// Old (manual conversion, may lose precision)
satoshi := wei / 1e10

// New (safe, validated)
satoshi, err := WeiToSatoshi(weiStr)
if err != nil {
    // Handle error
}
```

## See Also

- `wei.go` - Wei (1e18) conversions for Ethereum
- `safemath.go` - General-purpose multi-base conversions
- `scaled_converter.go` - High-performance uint64 key conversions
- `satoshi_examples.go` - Comprehensive usage examples

