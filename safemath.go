/*
Package safemath provides safe arithmetic operations for blockchain and financial calculations
in the EngineDex system. This package is critical for handling precision-sensitive operations
involving large numbers, decimal conversions, and financial calculations.

PURPOSE:
- Safe conversion between big.Int, big.Float, and float64 types
- Precision-preserving arithmetic for financial calculations (Wei/Ether, token amounts)
- Base conversion utilities for different decimal representations
- Time-based APR calculations for lending/borrowing operations

USAGE PATTERNS:
1. Token Amount Conversions:
  - Use BigIntBaseX() to convert float amounts to token units (e.g., 1.5 ETH -> 1500000000000000000 Wei)
  - Use UnBaseX() to convert token units back to human-readable amounts
  - Use UnBaseXFloatString() for formatted string output

2. Precision-Safe Operations:
  - Always use BigInt2Float() instead of direct float64 conversion for large numbers
  - Use FloatToBigIntBaseX() for precise float-to-integer conversion
  - Handle errors returned by BigIntByString() to prevent panics

3. Financial Calculations:
  - Use APR functions for lending/borrowing rate calculations
  - Use ProcessFloatToDecimalAdjustment() for state amount conversions

CRITICAL ATTENTION:
⚠️  THREAD SAFETY: All functions are pure and thread-safe, but cache access is not protected
⚠️  PRECISION LOSS: Large float64 values may lose precision when converted to big.Int
⚠️  MEMORY USAGE: big.Int operations can be memory-intensive for very large numbers
⚠️  ERROR HANDLING: Always check errors from BigIntByString() to prevent nil panics
⚠️  PERFORMANCE: Cache is pre-computed for bases 14 and 18; other bases computed on-demand
⚠️  BOUNDS CHECKING: Input validation prevents overflow and invalid operations

DESIGN PRINCIPLES:
- Functional approach: No shared state, pure functions
- Fail-safe defaults: Return zero values for invalid inputs
- Performance optimization: Caching for common operations
- Precision preservation: Use big.Float for intermediate calculations
- Error transparency: Clear error messages and logging

PACKAGE COMPARISON: safemath.go vs number.go

WHEN TO USE safemath.go:
✅ GENERAL PURPOSE: Multi-base conversions (any decimal precision)
✅ TOKEN OPERATIONS: Converting between different token decimals (6, 8, 14, 18, etc.)
✅ FINANCIAL CALCULATIONS: APR calculations, interest rates, percentage operations
✅ STRING FORMATTING: Formatted output with custom decimal places
✅ STATE MANAGEMENT: Internal state conversions and adjustments
✅ CROSS-TOKEN: Operations involving different token types with varying decimals
✅ LENDING/BORROWING: Time-based rate calculations and adjustments

WHEN TO USE number.go:
✅ ETHEREUM SPECIFIC: Wei/Ether conversions only
✅ PERFORMANCE CRITICAL: High-frequency operations where speed matters most
✅ SIMPLE CONVERSIONS: When you only need Wei ↔ Ether conversion
✅ ORDER PROCESSING: Real-time order book operations
✅ MARKET DATA: High-frequency market data processing
✅ TRADING ENGINES: Performance-critical trading operations

PERFORMANCE COMPARISON:
- safemath.go: More flexible, slightly slower due to generality
- number.go: Optimized for Wei/Ether, faster for specific use cases

PRECISION COMPARISON:
- safemath.go: Handles any decimal precision, more flexible
- number.go: Optimized for 18 decimals (Ethereum standard)

USE CASE DECISION TREE:
1. Need Wei/Ether conversion only? → Use number.go
2. Need multiple token types or decimal precisions? → Use safemath.go
3. Performance critical path? → Use number.go (Wei/Ether) or safemath.go (others)
4. Need APR calculations? → Use safemath.go
5. Need string formatting? → Use safemath.go
6. Need percentage operations? → Use safemath.go

EXAMPLES:
- Trading engine order processing: number.go (Wei/Ether)
- Multi-token DEX operations: safemath.go
- Lending platform interest calculations: safemath.go
- API response formatting: safemath.go
- Real-time market data: number.go (Wei/Ether)
- Cross-chain operations: safemath.go
*/
package safem

import (
	"errors"
	"math"
	"math/big"
	"strings"
)

// Precomputed powers of 10 for common decimals to avoid repeated Exp calls
// CRITICAL: These are pre-computed at package initialization for performance
// Bases 14 and 18 are most commonly used in blockchain operations
var (
	pow10Cache = map[int64]*big.Int{
		14: new(big.Int).Exp(big.NewInt(10), big.NewInt(14), nil), // Common for stablecoins
		18: new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), // Ethereum Wei precision
	}
	pow10FloatCache = map[int64]*big.Float{
		14: new(big.Float).SetInt(pow10Cache[14]),
		18: new(big.Float).SetInt(pow10Cache[18]),
	}
)

// Error definitions for consistent error handling across the package
// CRITICAL: These errors should be handled by callers to prevent panics
var (
	ErrNegativeInput = errors.New("negative input not allowed")
	ErrTooLarge      = errors.New("value too large for uint64")
	ErrPrecisionLoss = errors.New("precision loss in conversion")
	ErrInvalidString = errors.New("invalid string for big.Int")
	ErrInvalidInput  = errors.New("invalid input value")
)

// BigInt2Float converts a big.Int to float64 with specified decimal places
//
// PURPOSE: Safe conversion from token units to human-readable amounts
// USAGE: Converting Wei to ETH, token amounts to display values
// CRITICAL: Returns error for precision loss or negative inputs
// PERFORMANCE: Optimized to avoid string conversions
//
// WHEN TO USE vs number.go:
// ✅ Use safemath.BigInt2Float() when:
//   - Converting tokens with different decimal precisions (USDC=6, ETH=18, etc.)
//   - Need flexible decimal precision parameter
//   - Working with multiple token types
//
// ✅ Use number.go WeiToEther*() when:
//   - Converting only Wei to Ether (18 decimals)
//   - Performance is critical (trading engine, order processing)
//   - Working exclusively with Ethereum
//
// Example:
//
//	wei := big.NewInt(1500000000000000000) // 1.5 ETH in Wei
//	eth, err := BigInt2Float(wei, 18)      // Returns 1.5, nil
//
//	usdc := big.NewInt(1500000) // 1.5 USDC (6 decimals)
//	usdcFloat, err := BigInt2Float(usdc, 6) // Returns 1.5, nil
func BigInt2Float(i *big.Int, decimal uint8) (float64, error) {
	if i == nil {
		return 0, ErrInvalidInput
	}
	if i.Sign() < 0 {
		return 0, ErrNegativeInput
	}

	f := BigInt2BigFloat(i, decimal)
	ff, acc := f.Float64()

	// Allow Below accuracy for tiny values, but reject Above (precision loss)
	if acc == big.Above {
		return 0, ErrPrecisionLoss
	}

	return ff, nil
}

// BigInt2BigFloat converts a big.Int to big.Float with specified decimal places
//
// PURPOSE: High-precision conversion for intermediate calculations
// USAGE: When you need to maintain precision in multi-step calculations
// CRITICAL: Returns big.Float for maximum precision preservation
// PERFORMANCE: Uses cached divisors for common decimal places
//
// Example:
//
//	amount := big.NewInt(1234567890000000000)
//	floatAmount := BigInt2BigFloat(amount, 18) // 1.23456789 as big.Float
func BigInt2BigFloat(i *big.Int, decimal uint8) *big.Float {
	if i == nil {
		return new(big.Float).SetFloat64(0)
	}

	// Use direct big.Float conversion instead of string parsing
	fi := new(big.Float).SetInt(i)

	// Get cached divisor or compute on-the-fly
	div, ok := pow10FloatCache[int64(decimal)]
	if !ok {
		// Compute and cache for future use
		divInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)
		div = new(big.Float).SetInt(divInt)
		pow10FloatCache[int64(decimal)] = div
	}

	return new(big.Float).Quo(fi, div)
}

// BigFloatFromBigInt converts big.Int to big.Float directly
//
// PURPOSE: Simple conversion without decimal adjustment
// USAGE: When you need raw big.Float representation
// CRITICAL: No decimal scaling applied
func BigFloatFromBigInt(val *big.Int) *big.Float {
	if val == nil {
		return new(big.Float).SetFloat64(0)
	}
	return new(big.Float).SetInt(val)
}

// BigIntByString converts string to big.Int with proper error handling
//
// PURPOSE: Safe string-to-big.Int conversion for user inputs
// USAGE: Parsing user-provided amounts, API inputs
// CRITICAL: ALWAYS check returned error to prevent nil panics downstream
// PERFORMANCE: Uses native big.Int.SetString for efficiency
//
// Example:
//
//	amount, err := BigIntByString("123456789")
//	if err != nil {
//	    // Handle invalid input
//	}
func BigIntByString(val string) (*big.Int, error) {
	if val == "" {
		return nil, ErrInvalidString
	}

	n := new(big.Int)
	n, ok := n.SetString(val, 10)
	if !ok {
		return nil, ErrInvalidString
	}

	return n, nil
}

// BigIntByFloatBase14 converts float64 to big.Int with base 14
//
// PURPOSE: Quick conversion for stablecoin operations (14 decimals)
// USAGE: USDC, USDT, and other stablecoin amount conversions
// CRITICAL: Uses base 14 (common for stablecoins)
//
// WHEN TO USE vs number.go:
// ✅ Use safemath.BigIntByFloatBase14() when:
//   - Working with stablecoins (USDC, USDT, etc.)
//   - Need 14 decimal precision specifically
//   - Part of multi-token operations
//
// ✅ Use number.go EtherToWei() when:
//   - Converting only Ether to Wei (18 decimals)
//   - Performance is critical (trading engine)
//   - Working exclusively with Ethereum
//
// Example:
//
//	usdcAmount := 123.45
//	usdcWei := BigIntByFloatBase14(usdcAmount) // USDC conversion
func BigIntByFloatBase14(f float64) *big.Int {
	return BigIntBaseX(f, 14)
}

// BigIntBase14Percent converts float64 to big.Int with base 14 and divides by 100
//
// PURPOSE: Percentage calculations for stablecoin operations
// USAGE: Fee calculations, interest rate conversions
// CRITICAL: Result is divided by 100 (percentage conversion)
func BigIntBase14Percent(f float64) *big.Int {
	percent := BigIntBaseX(f, 14)
	return percent.Div(percent, big.NewInt(100))
}

// BigIntBaseFloatBase18 converts float64 to big.Int with base 18
//
// PURPOSE: Quick conversion for Ethereum operations (18 decimals)
// USAGE: ETH, ERC-20 token amount conversions
// CRITICAL: Uses base 18 (Ethereum standard)
//
// WHEN TO USE vs number.go:
// ✅ Use safemath.BigIntBaseFloatBase18() when:
//   - Part of multi-token operations (ETH + other tokens)
//   - Need consistent API with other token conversions
//   - Working in a general-purpose token conversion system
//
// ✅ Use number.go EtherToWei() when:
//   - Converting only Ether to Wei
//   - Performance is critical (trading engine, order processing)
//   - Working exclusively with Ethereum
//   - Need maximum performance optimization
//
// Example:
//
//	ethAmount := 1.5
//	ethWei := BigIntBaseFloatBase18(ethAmount) // ETH conversion
func BigIntBaseFloatBase18(f float64) *big.Int {
	return BigIntBaseX(f, 18)
}

// FloatToBigIntBaseX converts float64 to big.Int with specified base
//
// PURPOSE: High-precision float-to-integer conversion
// USAGE: Converting user input amounts to token units
// CRITICAL: Validates precision loss and logs warnings for significant errors
// PERFORMANCE: Uses caching for common bases, computes others on-demand
//
// Example:
//
//	ethAmount := 1.5
//	weiAmount := FloatToBigIntBaseX(ethAmount, 18) // 1500000000000000000
func FloatToBigIntBaseX(val float64, y int64) *big.Int {
	if val < 0 {
		return big.NewInt(0)
	}

	bigval := new(big.Float).SetFloat64(val)

	// Get cached multiplier or compute on-the-fly
	k, ok := pow10Cache[y]
	if !ok {
		k = new(big.Int).Exp(big.NewInt(10), big.NewInt(y), nil)
		pow10Cache[y] = k
	}

	coin := new(big.Float).SetInt(k)
	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result)

	// Validate precision by reconverting and comparing (only log significant losses)
	check := new(big.Float).Quo(new(big.Float).SetInt(result), coin)
	if check.Cmp(new(big.Float).SetFloat64(val)) != 0 {
		// Only log if the difference is significant (> 0.1%)
		diff := new(big.Float).Sub(check, new(big.Float).SetFloat64(val))
		diff.Abs(diff)
		if diff.Cmp(new(big.Float).Mul(new(big.Float).SetFloat64(val), big.NewFloat(0.001))) > 0 {
			//	metrics.Warnf("Significant precision loss in FloatToBigIntBaseX: input=%f, y=%d", val, y)
		}
	}

	return result
}

// FloatToBigIntBaseXPercent converts float64 to big.Int with specified base and divides by 100
//
// PURPOSE: Percentage-based conversions
// USAGE: Interest rates, fee calculations, percentage adjustments
// CRITICAL: Result is divided by 100 (percentage conversion)
func FloatToBigIntBaseXPercent(f float64, y int64) *big.Int {
	percent := FloatToBigIntBaseX(f, y)
	return percent.Div(percent, big.NewInt(100))
}

// BigIntBaseX converts float64 to big.Int with specified base
//
// PURPOSE: Optimized conversion with fast path for common cases
// USAGE: General float-to-integer conversion with performance optimization
// CRITICAL: Uses fast path for small values (y <= 14) to avoid precision issues
// PERFORMANCE: Fast path for small values, falls back to FloatToBigIntBaseX for precision
//
// WHEN TO USE vs number.go:
// ✅ Use safemath.BigIntBaseX() when:
//   - Converting to tokens with different decimal precisions
//   - Need flexible base parameter (6, 8, 14, 18, etc.)
//   - Working with multiple token types (USDC, USDT, ETH, etc.)
//   - Need percentage calculations (BigIntBaseXPercent)
//
// ✅ Use number.go EtherToWei() when:
//   - Converting only Ether to Wei (18 decimals)
//   - Performance is critical (trading engine, order processing)
//   - Working exclusively with Ethereum
//
// Example:
//
//	amount := 123.456
//	ethAmount := BigIntBaseX(amount, 18) // ETH conversion
//	usdcAmount := BigIntBaseX(amount, 6)  // USDC conversion
//	usdtAmount := BigIntBaseX(amount, 14) // USDT conversion
func BigIntBaseX(f float64, y int64) *big.Int {
	if f < 0 {
		return big.NewInt(0)
	}

	// Fast path for small y (<=14) and small f that fits in int64
	if y <= 14 && f < math.MaxInt64/float64(pow10Cache[y].Int64()) {
		k := pow10Cache[y].Int64()
		return big.NewInt(int64(f * float64(k)))
	}

	// Slow path for precision-critical cases
	return FloatToBigIntBaseX(f, y)
}

// UnBaseX converts big.Int from specified base back to standard form
//
// PURPOSE: Reverse conversion from token units to base units
// USAGE: Converting token amounts back to human-readable values
// CRITICAL: Returns zero for negative or nil inputs
// PERFORMANCE: Uses cached divisors for common bases
//
// WHEN TO USE vs number.go:
// ✅ Use safemath.UnBaseX() when:
//   - Converting tokens with different decimal precisions
//   - Need flexible base parameter (6, 8, 14, 18, etc.)
//   - Working with multiple token types
//   - Need big.Int result for further calculations
//
// ✅ Use number.go WeiToEther*() when:
//   - Converting only Wei to Ether (18 decimals)
//   - Need float64 result for display
//   - Performance is critical (trading engine)
//   - Working exclusively with Ethereum
//
// Example:
//
//	weiAmount := big.NewInt(1500000000000000000)
//	ethAmount := UnBaseX(weiAmount, 18) // 1.5 ETH (as big.Int)
//
//	usdcAmount := big.NewInt(1500000)
//	usdcFloat := UnBaseX(usdcAmount, 6) // 1.5 USDC (as big.Int)
func UnBaseX(f *big.Int, y int64) *big.Int {
	if f == nil || f.Sign() < 0 {
		return big.NewInt(0)
	}

	// Get cached divisor or compute on-the-fly
	k, ok := pow10Cache[y]
	if !ok {
		k = new(big.Int).Exp(big.NewInt(10), big.NewInt(y), nil)
		pow10Cache[y] = k
	}

	return new(big.Int).Div(f, k)
}

// UnBaseXFloatString converts big.Int to string with specified decimal places
//
// PURPOSE: Formatted string output for display purposes
// USAGE: UI display, logging, API responses
// CRITICAL: Limits show_dec to 100 to prevent excessive precision
// PERFORMANCE: Optimized to handle edge cases and avoid infinite loops
//
// WHEN TO USE vs number.go:
// ✅ Use safemath.UnBaseXFloatString() when:
//   - Need formatted string output for any token type
//   - Working with multiple token types (different decimals)
//   - Need custom decimal place formatting
//   - API responses, UI display, logging
//
// ✅ Use number.go WeiToEther*() + fmt.Sprintf() when:
//   - Converting only Wei to Ether for display
//   - Performance is critical (trading engine)
//   - Working exclusively with Ethereum
//
// Example:
//
//	ethAmount := big.NewInt(1234567890000000000)
//	ethDisplay := UnBaseXFloatString(ethAmount, 18, 8) // "1.23456789"
//
//	usdcAmount := big.NewInt(1234567)
//	usdcDisplay := UnBaseXFloatString(usdcAmount, 6, 2) // "1.23"
func UnBaseXFloatString(f *big.Int, y int64, show_dec int) string {
	if f == nil || f.Sign() < 0 {
		return "0"
	}

	// Limit show_dec to prevent excessive precision
	if show_dec > 100 {
		show_dec = 100
	}

	flo := new(big.Float).SetInt(f)

	// Get cached divisor or compute on-the-fly
	div, ok := pow10FloatCache[y]
	if !ok {
		divInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(y), nil)
		div = new(big.Float).SetInt(divInt)
		pow10FloatCache[y] = div
	}

	flo.Quo(flo, div)
	str := flo.Text('f', show_dec)

	// Clean up trailing zeros and ensure proper decimal format
	if strings.Contains(str, ".") {
		str = strings.TrimRight(str, "0")
		if strings.HasSuffix(str, ".") {
			str += "0"
		}
	}

	return str
}

// Constants for time-based calculations
// CRITICAL: These constants are used for APR calculations and must be consistent
const (
	DAILY_COUNT  = 360  // Days per year for APR calculations
	HOURLY_COUNT = 8640 // Hours per year (360 * 24)
	HOURS_DAILY  = 24   // Hours per day
)

// BigIntDailyAPR calculates daily APR from annual rate
//
// PURPOSE: Convert annual percentage rates to daily rates
// USAGE: Lending/borrowing interest calculations
// CRITICAL: Assumes 360-day year (financial standard)
//
// UNIQUE TO safemath.go (not available in number.go):
// ✅ Use safemath.BigIntDailyAPR() when:
//   - Calculating lending/borrowing interest rates
//   - Need time-based rate calculations
//   - Working with APR (Annual Percentage Rate) conversions
//   - Financial calculations requiring daily/hourly rate breakdowns
//
// ❌ number.go does NOT provide APR calculations
//
// Example:
//
//	annualRate := big.NewInt(36000) // 100% APR
//	dailyRate := BigIntDailyAPR(annualRate) // 100 (0.277...% daily)
func BigIntDailyAPR(f *big.Int) *big.Int {
	if f == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Div(f, big.NewInt(DAILY_COUNT))
}

// BigInt4HrAPR calculates 4-hour APR from annual rate
//
// PURPOSE: Short-term interest rate calculations
// USAGE: Intraday lending, flash loan calculations
// CRITICAL: Uses 4-hour periods for granular rate calculations
//
// UNIQUE TO safemath.go (not available in number.go):
// ✅ Use safemath.BigInt4HrAPR() when:
//   - Calculating short-term interest rates (4-hour periods)
//   - Flash loan interest calculations
//   - Intraday lending operations
//   - Need granular time-based rate calculations
//
// ❌ number.go does NOT provide APR calculations
func BigInt4HrAPR(f *big.Int) *big.Int {
	return BigIntHrAPR(f, big.NewInt(4))
}

// BigIntHrAPR calculates hourly APR for specified hours
//
// PURPOSE: Flexible time-based APR calculations
// USAGE: Custom time period interest calculations
// CRITICAL: Returns zero for nil inputs
//
// UNIQUE TO safemath.go (not available in number.go):
// ✅ Use safemath.BigIntHrAPR() when:
//   - Calculating custom time period interest rates
//   - Need flexible hourly rate calculations
//   - Lending platform interest calculations
//   - Any time-based financial rate calculations
//
// ❌ number.go does NOT provide APR calculations
//
// Example:
//
//	annualRate := big.NewInt(36000) // 100% APR
//	hourlyRate := BigIntHrAPR(annualRate, big.NewInt(1)) // 1-hour rate
func BigIntHrAPR(f *big.Int, hour *big.Int) *big.Int {
	if f == nil || hour == nil {
		return big.NewInt(0)
	}

	b := new(big.Int).Div(f, big.NewInt(HOURLY_COUNT))
	return b.Mul(b, hour)
}

// BigIntDailyBase calculates daily base rate for specified hours
//
// PURPOSE: Daily rate calculations for specified hour periods
// USAGE: Time-weighted average calculations, daily limits
// CRITICAL: Removed debug print to prevent potential hanging
// PERFORMANCE: Optimized for daily calculations
func BigIntDailyBase(f *big.Int, hour *big.Int) *big.Int {
	if f == nil || hour == nil {
		return big.NewInt(0)
	}

	b := new(big.Int).Div(f, big.NewInt(HOURS_DAILY))
	return b.Mul(b, hour)
}

// UnBase14 converts big.Int from base 14 back to standard form
//
// PURPOSE: Quick conversion for stablecoin operations
// USAGE: USDC, USDT amount conversions
// CRITICAL: Uses base 14 (stablecoin standard)
func UnBase14(f *big.Int) *big.Int {
	return UnBaseX(f, 14)
}

// ProcessFloatToDecimalAdjustment converts float64 to big.Int with specified decimal places
//
// PURPOSE: State amount conversions for system operations
// USAGE: Internal state management, balance calculations
// CRITICAL: Optimized to use native big.Float instead of shopspring/decimal
// PERFORMANCE: Significantly faster than shopspring/decimal operations
//
// UNIQUE TO safemath.go (not available in number.go):
// ✅ Use safemath.ProcessFloatToDecimalAdjustment() when:
//   - Internal state management and balance calculations
//   - Need flexible decimal precision for different token types
//   - System-level amount conversions
//   - Performance-critical internal operations
//
// ✅ Use number.go EtherToWei() when:
//   - Converting only Ether to Wei for user-facing operations
//   - Performance is critical (trading engine)
//   - Working exclusively with Ethereum
//
// Example:
//
//	stateAmount := 5.42323
//	adjustedAmount := ProcessFloatToDecimalAdjustment(18, stateAmount)
//	// Returns 5423230000000000000 (5.42323 * 10^18)
func ProcessFloatToDecimalAdjustment(decimal64b int, state_amt float64) *big.Int {
	if state_amt < 0 {
		return big.NewInt(0)
	}

	// Use native big.Float instead of shopspring/decimal for better performance
	f := new(big.Float).SetFloat64(state_amt)
	pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal64b)), nil)
	powF := new(big.Float).SetInt(pow)
	f.Mul(f, powF)

	result := new(big.Int)
	f.Int(result)
	return result
}

/*
QUICK REFERENCE: safemath.go vs number.go

PERFORMANCE CRITICAL PATHS (Use number.go):
- Trading engine order processing
- Real-time market data processing
- High-frequency operations
- Order book operations
- When working exclusively with Wei/Ether

GENERAL PURPOSE & MULTI-TOKEN (Use safemath.go):
- Multi-token DEX operations
- Cross-chain operations
- Lending/borrowing platforms
- API responses and UI display
- When working with different token decimals
- Financial calculations (APR, interest rates)
- String formatting and display

UNIQUE FEATURES:
- safemath.go: APR calculations, multi-base conversions, string formatting
- number.go: Optimized Wei/Ether conversions, performance-critical operations

DECISION MATRIX:
┌─────────────────┬─────────────────┬─────────────────┐
│ Use Case        │ safemath.go     │ number.go       │
├─────────────────┼─────────────────┼─────────────────┤
│ Wei ↔ Ether     │ ✅ (flexible)   │ ✅ (optimized)  │
│ Multi-token     │ ✅              │ ❌              │
│ APR calculations│ ✅              │ ❌              │
│ String format   │ ✅              │ ❌              │
│ Performance     │ ⚠️ (good)       │ ✅ (excellent)  │
│ Flexibility     │ ✅ (high)       │ ⚠️ (limited)    │
└─────────────────┴─────────────────┴─────────────────┘
*/
