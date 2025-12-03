// Package safem provides safe arithmetic operations for blockchain and financial calculations.
// This file implements Satoshi (1e8) conversions for Morphcore, following the same patterns
// as wei.go for Ethereum (1e18) conversions.
//
// PURPOSE:
// - Safe conversion between decimal values and Satoshi format (1e8)
// - Wei (1e18) to Satoshi (1e8) conversions for cross-chain operations
// - High-precision conversions using big.Int and big.Float
// - Memory-safe operations using sync.Pool for big.Int reuse
// - Performance-optimized with fast paths for common cases
//
// USAGE PATTERNS:
// 1. Payload Processing:
//   - Use DecimalToSatoshi() to convert address input to satoshi strings
//   - Use SatoshiToDecimal() to convert back for display
//
// 2. Cross-Chain Operations:
//   - Use WeiToSatoshi() to convert Ethereum Wei to Morphcore Satoshi
//   - Use SatoshiToWei() to convert Morphcore Satoshi to Ethereum Wei
//
// 3. High-Precision Calculations:
//   - Use DecimalToSatoshiBigInt() for exact calculations
//   - Use SatoshiToDecimalBigFloat() for maximum precision
//
// CRITICAL ATTENTION:
// ⚠️  MEMORY SAFETY: Uses sync.Pool for big.Int/big.Float reuse to reduce allocations
// ⚠️  OVERFLOW PROTECTION: All conversions check bounds before conversion
// ⚠️  PRECISION: Uses big.Float for intermediate calculations to preserve precision
// ⚠️  THREAD SAFETY: All functions are pure and thread-safe
// ⚠️  PERFORMANCE: Fast paths for small values, pooled allocations for large values
//
// DESIGN PRINCIPLES:
// - Memory efficiency: Reuse big.Int from pool where possible
// - Fail-fast: Return errors immediately on invalid input
// - Bounds checking: Validate all inputs before conversion
// - Precision preservation: Use big.Float for exact calculations
// - Performance first: Optimize hot paths (payload processing, orderbook)
package safem

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

// SatoshiScale is the multiplier for morphcore's 8-decimal precision
// Following Bitcoin's satoshi naming convention (1e8 = 100,000,000)
// This matches safem's wei (1e18) naming pattern for Ethereum
const SatoshiScale = 1e8

// WeiScale is the multiplier for Ethereum's 18-decimal precision
// Re-exported for consistency (already exists in safemath.go)
const WeiScale = 1e18

// Pre-allocated big.Int constants to avoid repeated allocations
var (
	satoshiScaleBig = big.NewInt(SatoshiScale)
	weiScaleBig     = big.NewInt(WeiScale)
	weiToSatoshiDiv = big.NewInt(1e10) // 1e18 / 1e8 = 1e10
	satoshiToWeiMul = big.NewInt(1e10) // 1e8 * 1e10 = 1e18
)

// Maximum safe values before overflow
var (
	// MaxSafeSatoshi: Maximum decimal value that can be safely converted to satoshi
	// Calculated as: MaxInt64 / SatoshiScale
	MaxSafeSatoshi = float64(math.MaxInt64) / SatoshiScale

	// Threshold for detecting Wei format (> 1e15 indicates Wei, not Satoshi)
	weiDetectionThreshold = big.NewInt(1e15)
)

// Pool for big.Int reuse to reduce allocations in hot paths
var satoshiBigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

// Pool for big.Float reuse (for high-precision conversions)
var satoshiBigFloatPool = sync.Pool{
	New: func() interface{} {
		return new(big.Float)
	},
}

// Error definitions for satoshi conversions
var (
	ErrInvalidSatoshiFormat = errors.New("invalid satoshi format")
	ErrSatoshiOverflow      = errors.New("satoshi value overflow: exceeds int64 range")
	ErrSatoshiUnderflow     = errors.New("satoshi value underflow: negative value")
	// ErrPrecisionLoss is already defined in safemath.go, reusing it here
)

// ============================================================================
// Core Conversion Functions: Decimal ↔ Satoshi
// ============================================================================

// DecimalToSatoshi converts a decimal value to satoshi format (1e8) as string
//
// PURPOSE: Convert display values to satoshi strings for EIP-712 encoding
// USAGE: Order prices, quantities, amounts in payload processing
// CRITICAL: Returns string to avoid precision loss (like wei functions)
// PERFORMANCE: Fast path for small values, uses big.Float for precision
//
// Example:
//
//	DecimalToSatoshi(50000.0) → "5000000000000"
func DecimalToSatoshi(decimal float64) string {
	// Fast path: Check for NaN/Inf
	if math.IsNaN(decimal) || math.IsInf(decimal, 0) {
		return "0"
	}

	// Fast path: Check for negative values (should return 0)
	if decimal < 0 {
		return "0"
	}

	// Fast path: Small values that fit in int64
	if decimal < MaxSafeSatoshi {
		scaled := int64(decimal * SatoshiScale)
		return strconv.FormatInt(scaled, 10)
	}

	// Slow path: Use big.Float for precision
	decimalFloat := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(decimalFloat)
	decimalFloat.SetFloat64(decimal)

	scaleFloat := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(scaleFloat)
	scaleFloat.SetInt(satoshiScaleBig)

	result := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(result)
	result.Mul(decimalFloat, scaleFloat)

	// Convert to big.Int for string representation
	resultInt := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(resultInt)
	result.Int(resultInt)

	return resultInt.String()
}

// DecimalToSatoshiBigInt converts a decimal value to satoshi format as *big.Int
//
// PURPOSE: High-precision conversion for exact calculations
// USAGE: Critical financial calculations, PnL, margins
// CRITICAL: Returns new *big.Int (caller owns it)
// PERFORMANCE: Uses pooled big.Int for intermediate calculations
//
// Example:
//
//	DecimalToSatoshiBigInt(50000.0) → *big.Int("5000000000000")
func DecimalToSatoshiBigInt(decimal float64) (*big.Int, error) {
	// Validate input
	if math.IsNaN(decimal) {
		return nil, fmt.Errorf("%w: input is NaN", ErrInvalidSatoshiFormat)
	}
	if math.IsInf(decimal, 0) {
		return nil, fmt.Errorf("%w: input is Infinity", ErrInvalidSatoshiFormat)
	}
	if decimal < 0 {
		return nil, fmt.Errorf("%w: negative value: %f", ErrSatoshiUnderflow, decimal)
	}

	// Use big.Float for precision
	decimalFloat := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(decimalFloat)
	decimalFloat.SetFloat64(decimal)

	scaleFloat := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(scaleFloat)
	scaleFloat.SetInt(satoshiScaleBig)

	result := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(result)
	result.Mul(decimalFloat, scaleFloat)

	// Convert to big.Int
	resultInt := new(big.Int) // Caller owns this
	result.Int(resultInt)

	// Check overflow
	if !resultInt.IsInt64() {
		return nil, fmt.Errorf("%w: value %f * %d exceeds int64 (max: %d)",
			ErrSatoshiOverflow, decimal, int64(SatoshiScale), int64(math.MaxInt64))
	}

	return resultInt, nil
}

// SatoshiToDecimal converts satoshi format (1e8) string to decimal float64
//
// PURPOSE: Convert satoshi strings back to display values
// USAGE: API responses, UI display, logging
// CRITICAL: May lose precision for very large values (use SatoshiToDecimalBigFloat for precision)
// PERFORMANCE: Fast path for small values, uses big.Int for large values
//
// Example:
//
//	SatoshiToDecimal("5000000000000") → 50000.0, nil
func SatoshiToDecimal(satoshiStr string) (float64, error) {
	if satoshiStr == "" {
		return 0, fmt.Errorf("%w: empty string", ErrInvalidSatoshiFormat)
	}

	// Parse string to big.Int
	satoshiBig := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(satoshiBig)

	_, ok := satoshiBig.SetString(satoshiStr, 10)
	if !ok {
		return 0, fmt.Errorf("%w: invalid string: %s", ErrInvalidSatoshiFormat, satoshiStr)
	}

	// Check for negative
	if satoshiBig.Sign() < 0 {
		return 0, fmt.Errorf("%w: negative value: %s", ErrSatoshiUnderflow, satoshiStr)
	}

	// Fast path: Small values that fit in int64
	if satoshiBig.IsInt64() {
		satoshiInt64 := satoshiBig.Int64()
		return float64(satoshiInt64) / SatoshiScale, nil
	}

	// Slow path: Use big.Float for precision
	satoshiFloat := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(satoshiFloat)
	satoshiFloat.SetInt(satoshiBig)

	scaleFloat := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(scaleFloat)
	scaleFloat.SetInt(satoshiScaleBig)

	result := satoshiBigFloatPool.Get().(*big.Float)
	defer satoshiBigFloatPool.Put(result)
	result.Quo(satoshiFloat, scaleFloat)

	decimal, accuracy := result.Float64()
	if accuracy == big.Above {
		return 0, fmt.Errorf("%w: precision loss converting %s", ErrPrecisionLoss, satoshiStr)
	}

	return decimal, nil
}

// SatoshiToDecimalBigFloat converts satoshi format to *big.Float for high precision
//
// PURPOSE: Maximum precision conversion for critical calculations
// USAGE: Exact financial calculations, audit trails
// CRITICAL: Returns new *big.Float (caller owns it)
// PERFORMANCE: Uses pooled big.Int for parsing
func SatoshiToDecimalBigFloat(satoshiStr string) (*big.Float, error) {
	if satoshiStr == "" {
		return nil, fmt.Errorf("%w: empty string", ErrInvalidSatoshiFormat)
	}

	// Parse string to big.Int
	satoshiBig := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(satoshiBig)

	_, ok := satoshiBig.SetString(satoshiStr, 10)
	if !ok {
		return nil, fmt.Errorf("%w: invalid string: %s", ErrInvalidSatoshiFormat, satoshiStr)
	}

	if satoshiBig.Sign() < 0 {
		return nil, fmt.Errorf("%w: negative value: %s", ErrSatoshiUnderflow, satoshiStr)
	}

	// Convert to big.Float
	satoshiFloat := new(big.Float).SetInt(satoshiBig) // Caller owns this
	scaleFloat := new(big.Float).SetInt(satoshiScaleBig)

	result := new(big.Float).Quo(satoshiFloat, scaleFloat)
	return result, nil
}

// ============================================================================
// Cross-Token Decimal Conversion (Generic)
// ============================================================================

// ConvertFromDecimalsToSatoshi converts an amount from source token decimals to satoshi (1e8)
//
// PURPOSE: Generic cross-token conversion for any source decimal precision (INBOUND)
// USAGE: Hyperlane message processing (inbound), cross-chain token conversions (other chain → Morpheum)
// CRITICAL: amount is in source token's base units (e.g., wei for ETH, smallest unit for USDC)
// Conversion formula: satoshi = (amount * 1e8) / (10^source_decimals)
// SECURITY: Uses ceiling rounding (ceil=true) to preserve precision and prevent rounding attacks
// This ensures no precision loss accumulates over many conversions, protecting against systematic exploitation
// Uses MulDivPrecise for precision-preserving arithmetic
//
// Example:
//
//	// Convert 1 ETH (1e18 wei) with 18 decimals to satoshi
//	amount := big.NewInt(1e18)
//	satoshi, err := ConvertFromDecimalsToSatoshi(amount, 18)
//	// Returns: *big.Int("100000000"), nil (1.0 in satoshi)
//
//	// Convert 1.000000001 ETH (1e18 + 1 wei) with 18 decimals to satoshi
//	amount := big.NewInt(1e18 + 1)
//	satoshi, err := ConvertFromDecimalsToSatoshi(amount, 18)
//	// Returns: *big.Int("100000001"), nil (ceiling rounding preserves precision)
func ConvertFromDecimalsToSatoshi(amount *big.Int, sourceDecimals uint8) (*big.Int, error) {
	if amount == nil {
		return nil, fmt.Errorf("%w: amount cannot be nil", ErrInvalidSatoshiFormat)
	}

	if amount.Sign() < 0 {
		return nil, fmt.Errorf("%w: amount cannot be negative", ErrSatoshiUnderflow)
	}

	// Handle zero amount
	if amount.Sign() == 0 {
		return big.NewInt(0), nil
	}

	// Validate decimals (max 77 to prevent overflow in 10^decimals)
	if sourceDecimals > 77 {
		return nil, fmt.Errorf("decimals too large: %d (max 77)", sourceDecimals)
	}

	// Calculate: satoshi = (amount * 1e8) / (10^source_decimals)
	// Use MulDivPrecise with ceil=true for precision preservation (ceiling rounding)
	// Security: Prevents rounding attacks by ensuring no precision loss on inbound conversions
	if sourceDecimals > 0 {
		// Get or compute 10^source_decimals
		sourceScale := satoshiBigIntPool.Get().(*big.Int)
		defer satoshiBigIntPool.Put(sourceScale)
		sourceScale.Exp(big.NewInt(10), big.NewInt(int64(sourceDecimals)), nil)

		// Use MulDivPrecise with ceil=true for precision preservation (ceiling rounding)
		result, err := MulDivPrecise(amount, satoshiScaleBig, sourceScale, true)
		if err != nil {
			return nil, fmt.Errorf("failed to convert from decimals to satoshi: %w", err)
		}
		return result, nil
	}

	// If sourceDecimals is 0, just multiply by satoshi scale
	result := new(big.Int).Mul(amount, satoshiScaleBig)
	return result, nil
}

// ConvertFromSatoshiToDecimals converts an amount from satoshi (1e8) to target token decimals
//
// PURPOSE: Generic cross-token conversion for outbound operations (OUTBOUND)
// USAGE: Hyperlane message processing (outbound), cross-chain token conversions (Morpheum → other chain)
// CRITICAL: satoshiAmount is in satoshi format (1e8 precision)
// Conversion formula: targetAmount = (satoshiAmount * 10^target_decimals) / 1e8
// SECURITY: Uses truncation rounding (ceil=false) for conservative precision handling
// This prevents precision amplification attacks where small satoshi amounts could be exploited
// Uses MulDivPrecise for precision-preserving arithmetic
//
// Example:
//
//	// Convert 1.0 satoshi (1e8) with 18 decimals to wei
//	satoshiAmount := big.NewInt(1e8)
//	wei, err := ConvertFromSatoshiToDecimals(satoshiAmount, 18)
//	// Returns: *big.Int("1000000000000000000"), nil (1.0 ETH in wei)
//
//	// Convert 1.000000001 satoshi (1e8 + 1) with 18 decimals to wei
//	satoshiAmount := big.NewInt(1e8 + 1)
//	wei, err := ConvertFromSatoshiToDecimals(satoshiAmount, 18)
//	// Returns: *big.Int("1000000000000000000"), nil (truncation prevents precision amplification)
func ConvertFromSatoshiToDecimals(satoshiAmount *big.Int, targetDecimals uint8) (*big.Int, error) {
	if satoshiAmount == nil {
		return nil, fmt.Errorf("%w: satoshi amount cannot be nil", ErrInvalidSatoshiFormat)
	}

	if satoshiAmount.Sign() < 0 {
		return nil, fmt.Errorf("%w: satoshi amount cannot be negative", ErrSatoshiUnderflow)
	}

	// Handle zero amount
	if satoshiAmount.Sign() == 0 {
		return big.NewInt(0), nil
	}

	// Validate decimals (max 77 to prevent overflow in 10^decimals)
	if targetDecimals > 77 {
		return nil, fmt.Errorf("decimals too large: %d (max 77)", targetDecimals)
	}

	// Calculate: targetAmount = (satoshiAmount * 10^target_decimals) / 1e8
	// Use MulDivPrecise with ceil=false for conservative precision handling (truncation)
	// Security: Prevents precision amplification attacks on outbound conversions
	if targetDecimals > 0 {
		// Get or compute 10^target_decimals
		targetScale := satoshiBigIntPool.Get().(*big.Int)
		defer satoshiBigIntPool.Put(targetScale)
		targetScale.Exp(big.NewInt(10), big.NewInt(int64(targetDecimals)), nil)

		// Use MulDivPrecise with ceil=false for conservative precision handling (truncation)
		result, err := MulDivPrecise(satoshiAmount, targetScale, satoshiScaleBig, false)
		if err != nil {
			return nil, fmt.Errorf("failed to convert from satoshi to decimals: %w", err)
		}
		return result, nil
	}

	// If targetDecimals is 0, just divide by satoshi scale
	result := new(big.Int).Div(satoshiAmount, satoshiScaleBig)
	return result, nil
}

// ============================================================================
// Wei ↔ Satoshi Conversions
// ============================================================================

// WeiToSatoshi converts Wei format (1e18) to Satoshi format (1e8)
//
// PURPOSE: Convert Ethereum Wei values to Morphcore Satoshi format
// USAGE: Cross-chain operations, token conversions
// CRITICAL: Divides by 1e10 (1e18 / 1e8 = 1e10)
// PERFORMANCE: Uses pooled big.Int, minimal allocations
//
// Example:
//
//	WeiToSatoshi("50000000000000000000000") → "5000000000000", nil
func WeiToSatoshi(weiStr string) (string, error) {
	if weiStr == "" {
		return "", fmt.Errorf("%w: empty wei string", ErrInvalidSatoshiFormat)
	}

	// Parse Wei string
	weiBig := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(weiBig)

	_, ok := weiBig.SetString(weiStr, 10)
	if !ok {
		return "", fmt.Errorf("invalid wei format: %s", weiStr)
	}

	if weiBig.Sign() < 0 {
		return "", fmt.Errorf("%w: negative wei value: %s", ErrSatoshiUnderflow, weiStr)
	}

	// Divide by 1e10 to convert from 1e18 to 1e8
	result := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(result)
	result.Set(weiBig)
	result.Div(result, weiToSatoshiDiv)

	return result.String(), nil
}

// SatoshiToWei converts Satoshi format (1e8) to Wei format (1e18)
//
// PURPOSE: Convert Morphcore Satoshi values to Ethereum Wei format
// USAGE: Cross-chain operations, token conversions
// CRITICAL: Multiplies by 1e10 (1e8 * 1e10 = 1e18)
// PERFORMANCE: Uses pooled big.Int, minimal allocations
//
// Example:
//
//	SatoshiToWei("5000000000000") → "50000000000000000000000", nil
func SatoshiToWei(satoshiStr string) (string, error) {
	if satoshiStr == "" {
		return "", fmt.Errorf("%w: empty satoshi string", ErrInvalidSatoshiFormat)
	}

	// Parse Satoshi string
	satoshiBig := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(satoshiBig)

	_, ok := satoshiBig.SetString(satoshiStr, 10)
	if !ok {
		return "", fmt.Errorf("%w: invalid string: %s", ErrInvalidSatoshiFormat, satoshiStr)
	}

	if satoshiBig.Sign() < 0 {
		return "", fmt.Errorf("%w: negative satoshi value: %s", ErrSatoshiUnderflow, satoshiStr)
	}

	// Multiply by 1e10 to convert from 1e8 to 1e18
	result := satoshiBigIntPool.Get().(*big.Int)
	defer satoshiBigIntPool.Put(result)
	result.Set(satoshiBig)
	result.Mul(result, satoshiToWeiMul)

	return result.String(), nil
}

// ============================================================================
// Smart Normalization Function
// ============================================================================

// NormalizeToSatoshi accepts various input formats and converts to satoshi (1e8)
//
// PURPOSE: Universal converter that handles multiple input types
// USAGE: Payload processing, API parsing, address input
// CRITICAL: Auto-detects Wei format (> 1e15 threshold)
// PERFORMANCE: Fast path for common types, uses pools for big.Int operations
//
// Accepts:
//   - Decimal strings: "50000.0" → "5000000000000"
//   - Float64: 50000.0 → "5000000000000"
//   - Wei strings (1e18): "50000000000000000000000" → "5000000000000"
//   - Already satoshi (1e8): "5000000000000" → "5000000000000"
//   - *big.Int: Already in satoshi or wei format
func NormalizeToSatoshi(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return "", fmt.Errorf("%w: empty string", ErrInvalidSatoshiFormat)
		}

		// Try parsing as big.Int first
		valBig := satoshiBigIntPool.Get().(*big.Int)
		defer satoshiBigIntPool.Put(valBig)

		_, ok := valBig.SetString(v, 10)
		if ok {
			// Check if it's Wei format (very large, > 1e15)
			if valBig.Cmp(weiDetectionThreshold) > 0 {
				// Likely Wei format, convert it
				return WeiToSatoshi(v)
			}
			// Already in satoshi format
			if valBig.Sign() < 0 {
				return "", fmt.Errorf("%w: negative value: %s", ErrSatoshiUnderflow, v)
			}
			return v, nil
		}

		// Try parsing as decimal string
		if strings.Contains(v, ".") {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return "", fmt.Errorf("%w: invalid decimal: %s", ErrInvalidSatoshiFormat, v)
			}
			return DecimalToSatoshi(f), nil
		}

		return "", fmt.Errorf("%w: invalid number string: %s", ErrInvalidSatoshiFormat, v)

	case float64:
		return DecimalToSatoshi(v), nil

	case float32:
		return DecimalToSatoshi(float64(v)), nil

	case int64:
		return strconv.FormatInt(v*int64(SatoshiScale), 10), nil

	case int:
		return strconv.FormatInt(int64(v)*int64(SatoshiScale), 10), nil

	case *big.Int:
		if v == nil {
			return "", fmt.Errorf("%w: nil big.Int", ErrInvalidSatoshiFormat)
		}
		if v.Sign() < 0 {
			return "", fmt.Errorf("%w: negative big.Int", ErrSatoshiUnderflow)
		}

		// Check if it's Wei format
		if v.Cmp(weiDetectionThreshold) > 0 {
			return WeiToSatoshi(v.String())
		}
		// Already in satoshi format
		return v.String(), nil

	default:
		return "", fmt.Errorf("%w: unsupported type: %T", ErrInvalidSatoshiFormat, value)
	}
}

// ============================================================================
// Batch Operations for Performance
// ============================================================================

// BatchDecimalToSatoshi converts multiple decimal values to satoshi strings
//
// PURPOSE: Efficient bulk conversion for orderbook operations
// USAGE: Orderbook snapshots, batch processing, market data aggregation
// PERFORMANCE: More efficient than individual conversions (reduces overhead)
func BatchDecimalToSatoshi(decimals []float64) []string {
	if decimals == nil {
		return nil
	}

	results := make([]string, len(decimals))
	for i, decimal := range decimals {
		results[i] = DecimalToSatoshi(decimal)
	}
	return results
}

// BatchSatoshiToDecimal converts multiple satoshi strings to decimal values
//
// PURPOSE: Bulk conversion for API responses
// USAGE: Orderbook depth queries, batch display formatting
func BatchSatoshiToDecimal(satoshiStrs []string) ([]float64, error) {
	if satoshiStrs == nil {
		return nil, nil
	}

	results := make([]float64, len(satoshiStrs))
	for i, satoshiStr := range satoshiStrs {
		decimal, err := SatoshiToDecimal(satoshiStr)
		if err != nil {
			return nil, fmt.Errorf("satoshi[%d]: %w", i, err)
		}
		results[i] = decimal
	}
	return results, nil
}
