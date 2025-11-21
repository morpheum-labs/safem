# SafeMath API Documentation

## Overview

The `safemath.go` file provides general-purpose safe arithmetic operations for blockchain and financial calculations. It supports multiple decimal precisions (6, 8, 14, 18, etc.) and is designed for multi-token operations, cross-chain conversions, and financial calculations including APR (Annual Percentage Rate) computations.

## Purpose

- **Multi-Token Operations**: Convert between different token decimals (USDC=6, ETH=18, etc.)
- **Flexible Precision**: Support any decimal precision (not just 18)
- **Financial Calculations**: APR calculations, interest rates, percentage operations
- **String Formatting**: Formatted output with custom decimal places
- **State Management**: Internal state conversions and adjustments

## Key Concepts

### Decimal Bases

Different tokens use different decimal precisions:

- **USDC/USDT**: 6 decimals (1 USDC = 1,000,000 units)
- **Morphcore Satoshi**: 8 decimals (1 unit = 100,000,000 satoshi)
- **Stablecoins**: 14 decimals (common for some stablecoins)
- **Ethereum/ETH**: 18 decimals (1 ETH = 1,000,000,000,000,000,000 wei)

### Base Conversion

The module uses "base" terminology for decimal precision:

```go
BigIntBaseX(amount, 18)  // ETH (18 decimals)
BigIntBaseX(amount, 6)   // USDC (6 decimals)
BigIntBaseX(amount, 8)   // Satoshi (8 decimals)
BigIntBaseX(amount, 14)  // Stablecoin (14 decimals)
```

## API Reference

### Core Conversion Functions

#### BigInt2Float

```go
func BigInt2Float(i *big.Int, decimal uint8) (float64, error)
```

**Purpose**: Convert token units to human-readable amounts

**Use When**:
- Converting tokens with different decimal precisions
- Need flexible decimal precision parameter
- Working with multiple token types

**Example**:
```go
// ETH (18 decimals)
wei := big.NewInt(1500000000000000000) // 1.5 ETH
eth, err := BigInt2Float(wei, 18)
// Returns: 1.5, nil

// USDC (6 decimals)
usdc := big.NewInt(1500000) // 1.5 USDC
usdcFloat, err := BigInt2Float(usdc, 6)
// Returns: 1.5, nil
```

#### BigInt2BigFloat

```go
func BigInt2BigFloat(i *big.Int, decimal uint8) *big.Float
```

**Purpose**: High-precision conversion for intermediate calculations

**Use When**:
- Need to maintain precision in multi-step calculations
- Intermediate calculations requiring exact precision
- Complex financial calculations

**Example**:
```go
amount := big.NewInt(1234567890000000000)
floatAmount := BigInt2BigFloat(amount, 18)
// Returns: *big.Float with maximum precision
```

#### BigIntByString

```go
func BigIntByString(val string) (*big.Int, error)
```

**Purpose**: Safe string-to-big.Int conversion

**Use When**:
- Parsing user-provided amounts
- API inputs
- Configuration values

**CRITICAL**: Always check returned error to prevent nil panics

**Example**:
```go
amount, err := BigIntByString("123456789")
if err != nil {
    return fmt.Errorf("invalid amount: %w", err)
}
```

#### FloatToBigIntBaseX

```go
func FloatToBigIntBaseX(val float64, y int64) *big.Int
```

**Purpose**: High-precision float-to-integer conversion

**Use When**:
- Converting user input amounts to token units
- Need specified decimal precision
- Multi-token operations

**Example**:
```go
ethAmount := 1.5
weiAmount := FloatToBigIntBaseX(ethAmount, 18)
// Returns: *big.Int("1500000000000000000")

usdcAmount := 1.5
usdcUnits := FloatToBigIntBaseX(usdcAmount, 6)
// Returns: *big.Int("1500000")
```

#### BigIntBaseX

```go
func BigIntBaseX(f float64, y int64) *big.Int
```

**Purpose**: Optimized conversion with fast path for common cases

**Use When**:
- Converting to tokens with different decimal precisions
- Need flexible base parameter
- Working with multiple token types

**Performance**: Fast path for small values (y <= 14), falls back to FloatToBigIntBaseX for precision

**Example**:
```go
amount := 123.456
ethAmount := BigIntBaseX(amount, 18)  // ETH conversion
usdcAmount := BigIntBaseX(amount, 6)  // USDC conversion
usdtAmount := BigIntBaseX(amount, 14) // USDT conversion
```

### Token-Specific Functions

#### BigIntByFloatBase14

```go
func BigIntByFloatBase14(f float64) *big.Int
```

**Purpose**: Quick conversion for stablecoin operations (14 decimals)

**Use When**:
- Working with USDC, USDT, and other stablecoins
- Need 14 decimal precision specifically

**Example**:
```go
usdcAmount := 123.45
usdcWei := BigIntByFloatBase14(usdcAmount)
```

#### BigIntBaseFloatBase18

```go
func BigIntBaseFloatBase18(f float64) *big.Int
```

**Purpose**: Quick conversion for Ethereum operations (18 decimals)

**Use When**:
- Part of multi-token operations
- Need consistent API with other token conversions

**Note**: For pure Ethereum operations, consider using `wei.go` for better performance

**Example**:
```go
ethAmount := 1.5
ethWei := BigIntBaseFloatBase18(ethAmount)
```

### Reverse Conversion Functions

#### UnBaseX

```go
func UnBaseX(f *big.Int, y int64) *big.Int
```

**Purpose**: Reverse conversion from token units to base units

**Use When**:
- Converting token amounts back to human-readable values
- Need big.Int result for further calculations
- Multi-token operations

**Example**:
```go
weiAmount := big.NewInt(1500000000000000000)
ethAmount := UnBaseX(weiAmount, 18) // Returns *big.Int representing 1.5

usdcAmount := big.NewInt(1500000)
usdcFloat := UnBaseX(usdcAmount, 6) // Returns *big.Int representing 1.5
```

#### UnBaseXFloatString

```go
func UnBaseXFloatString(f *big.Int, y int64, show_dec int) string
```

**Purpose**: Formatted string output for display

**Use When**:
- UI display
- Logging
- API responses
- Need custom decimal place formatting

**Example**:
```go
ethAmount := big.NewInt(1234567890000000000)
ethDisplay := UnBaseXFloatString(ethAmount, 18, 8)
// Returns: "1.23456789"

usdcAmount := big.NewInt(1234567)
usdcDisplay := UnBaseXFloatString(usdcAmount, 6, 2)
// Returns: "1.23"
```

### APR Calculation Functions

#### BigIntDailyAPR

```go
func BigIntDailyAPR(f *big.Int) *big.Int
```

**Purpose**: Convert annual percentage rates to daily rates

**Use When**:
- Lending/borrowing interest calculations
- Daily rate calculations
- Financial calculations requiring daily breakdowns

**Assumes**: 360-day year (financial standard)

**Example**:
```go
annualRate := big.NewInt(36000) // 100% APR
dailyRate := BigIntDailyAPR(annualRate)
// Returns: *big.Int("100") (0.277...% daily)
```

#### BigInt4HrAPR

```go
func BigInt4HrAPR(f *big.Int) *big.Int
```

**Purpose**: Calculate 4-hour APR from annual rate

**Use When**:
- Short-term interest rate calculations
- Intraday lending
- Flash loan calculations

**Example**:
```go
annualRate := big.NewInt(36000) // 100% APR
fourHourRate := BigInt4HrAPR(annualRate)
```

#### BigIntHrAPR

```go
func BigIntHrAPR(f *big.Int, hour *big.Int) *big.Int
```

**Purpose**: Flexible time-based APR calculations

**Use When**:
- Custom time period interest rates
- Flexible hourly rate calculations
- Lending platform interest calculations

**Example**:
```go
annualRate := big.NewInt(36000) // 100% APR
hourlyRate := BigIntHrAPR(annualRate, big.NewInt(1)) // 1-hour rate
```

### State Management Functions

#### ProcessFloatToDecimalAdjustment

```go
func ProcessFloatToDecimalAdjustment(decimal64b int, state_amt float64) *big.Int
```

**Purpose**: State amount conversions for system operations

**Use When**:
- Internal state management
- Balance calculations
- System-level amount conversions

**Performance**: Optimized to use native big.Float (faster than shopspring/decimal)

**Example**:
```go
stateAmount := 5.42323
adjustedAmount := ProcessFloatToDecimalAdjustment(18, stateAmount)
// Returns: *big.Int("5423230000000000000") (5.42323 * 10^18)
```

## Usage Patterns

### Pattern 1: Multi-Token Operations

```go
// Convert different token types
tokens := map[string]struct {
    amount float64
    decimals int64
}{
    "ETH":  {1.5, 18},
    "USDC": {100.0, 6},
    "USDT": {50.0, 14},
}

for token, data := range tokens {
    units := BigIntBaseX(data.amount, data.decimals)
    fmt.Printf("%s: %s units\n", token, units.String())
}
```

### Pattern 2: APR Calculations

```go
// Calculate daily interest from annual rate
annualAPR := big.NewInt(7200) // 20% APR
dailyAPR := BigIntDailyAPR(annualAPR)

// Calculate interest for a loan
loanAmount := big.NewInt(1000000000000000000) // 1 ETH
dailyInterest := new(big.Int).Mul(loanAmount, dailyAPR)
dailyInterest.Div(dailyInterest, big.NewInt(10000)) // Divide by 10000 for percentage
```

### Pattern 3: String Formatting

```go
// Format token amounts for display
tokenAmounts := map[string]*big.Int{
    "ETH":  big.NewInt(1234567890000000000),
    "USDC": big.NewInt(1234567),
}

for token, amount := range tokenAmounts {
    decimals := int64(18)
    if token == "USDC" {
        decimals = 6
    }
    display := UnBaseXFloatString(amount, decimals, 8)
    fmt.Printf("%s: %s\n", token, display)
}
```

### Pattern 4: Cross-Token Calculations

```go
// Convert between different token types
ethAmount := 1.5
usdcAmount := 2000.0

// Convert to common units
ethUnits := BigIntBaseX(ethAmount, 18)
usdcUnits := BigIntBaseX(usdcAmount, 6)

// Perform calculations
totalValue := new(big.Int).Add(ethUnits, usdcUnits)
```

## Performance Comparison

### vs wei.go

| Aspect | safemath.go | wei.go |
|--------|-------------|--------|
| Flexibility | ✅ High (any decimals) | ⚠️ Limited (18 only) |
| Performance | ⚠️ Good | ✅ Excellent |
| Use Case | Multi-token | Ethereum only |
| APR Calculations | ✅ Yes | ❌ No |

### vs satoshi.go

| Aspect | safemath.go | satoshi.go |
|--------|-------------|------------|
| Precision | Flexible (any) | Fixed (8) |
| Cross-Chain | ❌ No | ✅ Yes (Wei↔Satoshi) |
| Memory Safety | ⚠️ Standard | ✅ sync.Pool |
| Use Case | Multi-token | Morphcore only |

## When to Use vs Other Packages

### Use safemath.go when:

✅ **Multi-Token Operations**
- Working with different token types (ETH, USDC, USDT, etc.)
- Need flexible decimal precision
- Cross-token calculations

✅ **Financial Calculations**
- APR calculations
- Interest rate computations
- Percentage operations
- Time-based rate calculations

✅ **String Formatting**
- Formatted output for display
- Custom decimal place formatting
- API responses

✅ **General Purpose**
- When you need flexibility over performance
- Working with various decimal precisions
- State management conversions

### Use wei.go when:

✅ **Ethereum-Only Operations**
- Working exclusively with Wei/Ether
- Performance is critical
- High-frequency operations

### Use satoshi.go when:

✅ **Morphcore Operations**
- Working with 8-decimal precision
- Cross-chain Wei ↔ Satoshi conversion
- Payload processing

### Use scaled_converter.go when:

✅ **High-Performance Operations**
- Orderbook operations
- Need uint64 keys for fast sorting
- Risk engine calculations

## Best Practices

### 1. Choose Appropriate Base

```go
// ❌ WRONG: Using wrong decimal precision
usdcUnits := BigIntBaseX(amount, 18) // Wrong! USDC uses 6 decimals

// ✅ CORRECT: Use correct decimal precision
usdcUnits := BigIntBaseX(amount, 6) // Correct for USDC
```

### 2. Always Check Errors

```go
// ❌ WRONG: Ignoring errors
amount, _ := BigIntByString(input)

// ✅ CORRECT: Check and handle errors
amount, err := BigIntByString(input)
if err != nil {
    return fmt.Errorf("invalid amount: %w", err)
}
```

### 3. Use Appropriate Precision

```go
// ❌ WRONG: Using float64 for exact calculations
total := ethAmount + usdcAmount // May lose precision

// ✅ CORRECT: Use big.Int for exact calculations
ethBig := BigIntBaseX(ethAmount, 18)
usdcBig := BigIntBaseX(usdcAmount, 6)
total := new(big.Int).Add(ethBig, usdcBig)
```

### 4. Cache Common Operations

```go
// Pre-computed powers of 10 are cached for bases 14 and 18
// Other bases are computed on-demand and cached
// No need to manually cache - handled internally
```

## Error Handling

### Common Errors

1. **Invalid Input**: `"invalid input value"`
   - **Cause**: Nil *big.Int or invalid value
   - **Solution**: Validate input before calling

2. **Negative Input**: `"negative input not allowed"`
   - **Cause**: Negative values not allowed
   - **Solution**: Validate input is non-negative

3. **Precision Loss**: `"precision loss in conversion"`
   - **Cause**: Value too large for float64 precision
   - **Solution**: Use big.Float versions for large values

4. **Invalid String**: `"invalid string for big.Int"`
   - **Cause**: Cannot parse string as number
   - **Solution**: Validate string format before parsing

## Examples

### Multi-Token DEX Operations

```go
// Handle multiple token types in DEX
func processSwap(fromToken string, toToken string, amount float64) {
    // Get decimal precision for each token
    fromDecimals := getTokenDecimals(fromToken)
    toDecimals := getTokenDecimals(toToken)
    
    // Convert to units
    fromUnits := BigIntBaseX(amount, fromDecimals)
    toUnits := calculateSwapAmount(fromUnits, fromDecimals, toDecimals)
    
    // Convert back for display
    toAmount, _ := BigInt2Float(toUnits, uint8(toDecimals))
    fmt.Printf("Swap: %.8f %s → %.8f %s\n", amount, fromToken, toAmount, toToken)
}
```

### Lending Platform Interest

```go
// Calculate daily interest
func calculateDailyInterest(loanAmount *big.Int, annualAPR *big.Int) *big.Int {
    dailyAPR := BigIntDailyAPR(annualAPR)
    // Interest = (loanAmount * dailyAPR) / 10000
    interest := new(big.Int).Mul(loanAmount, dailyAPR)
    interest.Div(interest, big.NewInt(10000))
    return interest
}
```

## Thread Safety

All functions in `safemath.go` are **thread-safe** and can be called concurrently. The internal caches (pow10Cache, pow10FloatCache) are safe for concurrent read access.

## Performance Notes

### Caching

The module pre-computes and caches powers of 10 for bases 14 and 18 (most common). Other bases are computed on-demand and cached for future use.

### Fast Paths

- `BigIntBaseX`: Fast path for small values (y <= 14) that fit in int64
- Falls back to `FloatToBigIntBaseX` for precision-critical cases

## Migration Guide

### From Manual Conversion

If you're manually converting:

```go
// Old (manual, error-prone)
units := int64(amount * math.Pow(10, decimals))

// New (safe, validated)
units := BigIntBaseX(amount, decimals)
```

### From wei.go (Ethereum-Only)

If you're using wei.go but need multi-token support:

```go
// Old (wei.go - Ethereum only)
ethWei, _ := EtherToWei(amount)

// New (safemath.go - Multi-token)
ethWei := BigIntBaseX(amount, 18)  // ETH
usdcUnits := BigIntBaseX(amount, 6) // USDC
```

## See Also

- `wei.go` - Optimized Wei/Ether conversions (Ethereum only)
- `satoshi.go` - Satoshi conversions (Morphcore, 8 decimals)
- `scaled_converter.go` - High-performance uint64 key conversions
- `safemath.go` examples in code comments

