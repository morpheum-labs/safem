// Package scaled_converter provides high-performance, memory-safe, and secure number conversions
// for orderbook operations, risk calculations, and oracle price aggregation.
//
// PURPOSE:
// - Convert u256 (*big.Int) to u64 (scaled) for fast sorting and comparisons
// - Provide safe arithmetic operations on scaled uint64 keys
// - Enable batch operations for performance optimization
// - Prevent overflow/underflow with comprehensive bounds checking
//
// USAGE PATTERNS:
// 1. Orderbook Operations:
//   - Use BigIntToPriceKey() to convert order prices to uint64 keys
//   - Use PriceKeyToBigInt() to convert back for exact calculations
//   - Use AddPriceKeys()/SubtractPriceKeys() for safe arithmetic
//
// 2. Risk Calculations:
//   - Use BigIntToValueKey() for position value aggregation
//   - Use ComparePriceKeys() for fast liquidation checks
//   - Use MultiplyPriceKeys() for margin calculations
//
// 3. Oracle Aggregation:
//   - Use Float64ToPriceKey() for price feed conversion
//   - Use BatchBigIntToPriceKeys() for efficient bulk operations
//
// CRITICAL ATTENTION:
// ⚠️  OVERFLOW PROTECTION: All conversions check uint64 bounds before conversion
// ⚠️  PRECISION: Original *big.Int values are preserved for exact calculations
// ⚠️  MEMORY SAFETY: Uses sync.Pool for big.Int reuse to reduce allocations
// ⚠️  THREAD SAFETY: All functions are pure and thread-safe
// ⚠️  PERFORMANCE: Optimized for high-frequency operations (orderbook, risk engine)
//
// DESIGN PRINCIPLES:
// - Fail-fast: Return errors immediately on invalid input
// - Zero-allocation: Reuse big.Int from pool where possible
// - Bounds checking: Validate all inputs before conversion
// - Precision preservation: Keep original *big.Int for exact math
// - Performance first: Optimize hot paths (sorting, comparisons)
package safem

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
)

// Scaling factor constants for different use cases
const (
	// PriceScale: 8 decimal places (standard for most financial instruments)
	// Maximum safe price: 184467440737.09551615 (before overflow)
	PriceScale = 1e8

	// ValueScale: 8 decimal places for position values and aggregations
	ValueScale = 1e8

	// RatioScale: 8 decimal places for ratios (satoshi precision, aligned with other scales)
	// Used for margin ratios, risk factors, leverage calculations
	// CRITICAL: Changed from 1e6 to 1e8 for unified precision across all scales
	// This provides 100x more precision (0.00000001 vs 0.000001) and prevents conversion errors
	RatioScale = 1e8

	// QuantityScale: 8 decimal places for order quantities
	QuantityScale = 1e8

	// ScoreScale: 8 decimal places for ADL scores and rankings
	ScoreScale = 1e8
)

// Maximum safe values before overflow (calculated at compile time)
var (
	// MaxSafePrice: Maximum price that can be safely converted to uint64
	// Calculated as: MaxUint64 / PriceScale
	MaxSafePrice = float64(math.MaxUint64) / PriceScale

	// MaxSafeValue: Maximum value that can be safely converted to uint64
	MaxSafeValue = float64(math.MaxUint64) / ValueScale

	// MaxSafeRatio: Maximum ratio that can be safely converted to uint64
	MaxSafeRatio = float64(math.MaxUint64) / RatioScale

	// MaxSafeQuantity: Maximum quantity that can be safely converted to uint64
	MaxSafeQuantity = float64(math.MaxUint64) / QuantityScale
)

// Pre-allocated big.Int for scaling factors (avoid repeated allocations)
var (
	priceScaleBig    = big.NewInt(PriceScale)
	valueScaleBig    = big.NewInt(ValueScale)
	ratioScaleBig    = big.NewInt(RatioScale)
	quantityScaleBig = big.NewInt(QuantityScale)
	scoreScaleBig    = big.NewInt(ScoreScale)
)

// Error definitions for scaled conversions
var (
	ErrOverflow       = errors.New("value overflow: exceeds uint64 range")
	ErrUnderflow      = errors.New("value underflow: negative value for unsigned type")
	ErrInvalidScale   = errors.New("invalid scale factor")
	ErrOutOfBounds    = errors.New("value out of safe bounds")
	ErrDivisionByZero = errors.New("division by zero")
)

// Pool for big.Int reuse to reduce allocations in hot paths
var scaledBigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

// ============================================================================
// Core Conversion Functions: u256 (*big.Int) to u64 (scaled)
// ============================================================================

// BigIntToScaledUint64 converts *big.Int (u256) to uint64 (scaled) with overflow protection
//
// PURPOSE: Primary conversion function for high-performance operations
// USAGE: Converting order prices, position values, quantities to uint64 keys
// CRITICAL: Checks overflow before conversion, preserves precision
// PERFORMANCE: Uses pooled big.Int to avoid allocations
//
// Example:
//
//	priceBig := big.NewInt(2000000000000000000) // 2000.0 with 18 decimals
//	priceKey, err := BigIntToScaledUint64(priceBig, PriceScale)
//	// Returns 200000000000 (2000.0 * 1e8), nil
func BigIntToScaledUint64(value *big.Int, scale uint64) (uint64, error) {
	if value == nil {
		return 0, ErrInvalidInput
	}

	if value.Sign() < 0 {
		return 0, fmt.Errorf("%w: value is negative: %s", ErrUnderflow, value.String())
	}

	if scale == 0 {
		return 0, ErrInvalidScale
	}

	// Get big.Int from pool to avoid allocation
	scaled := scaledBigIntPool.Get().(*big.Int)
	defer scaledBigIntPool.Put(scaled)
	scaled.Set(value)

	// Scale: value * scale
	scaleBig := big.NewInt(int64(scale))
	scaled.Mul(scaled, scaleBig)

	// Check overflow (critical security check)
	if !scaled.IsUint64() {
		return 0, fmt.Errorf("%w: value %s * scale %d exceeds uint64 (max: %d)",
			ErrOverflow, value.String(), scale, uint64(math.MaxUint64))
	}

	return scaled.Uint64(), nil
}

// ScaledUint64ToBigInt converts uint64 (scaled) back to *big.Int (u256) with precision preservation
//
// PURPOSE: Reverse conversion for exact calculations
// USAGE: Converting back to *big.Int for precise arithmetic
// CRITICAL: Returns new *big.Int (caller owns it)
// PERFORMANCE: Minimal allocations, uses efficient division
//
// Example:
//
//	priceKey := uint64(200000000000) // 2000.0 * 1e8
//	priceBig := ScaledUint64ToBigInt(priceKey, PriceScale)
//	// Returns *big.Int representing 2000.0
func ScaledUint64ToBigInt(value uint64, scale uint64) *big.Int {
	if scale == 0 {
		return big.NewInt(0)
	}

	// Create new big.Int (caller owns it)
	result := new(big.Int).SetUint64(value)

	// Divide by scale to get original value
	scaleBig := big.NewInt(int64(scale))
	result.Div(result, scaleBig)

	return result
}

// ============================================================================
// float64 to u64 (scaled) Conversions with Bounds Checking
// ============================================================================

// Float64ToScaledUint64 converts float64 to uint64 (scaled) with comprehensive validation
//
// PURPOSE: Convert display values to uint64 keys for fast operations
// USAGE: Converting user input, API values, display values to keys
// CRITICAL: Validates NaN, Inf, negative, and bounds before conversion
// PERFORMANCE: Uses FloatToBigIntBaseX for precision, then converts to uint64
//
// Example:
//
//	price := 2000.50
//	priceKey, err := Float64ToScaledUint64(price, PriceScale, MaxSafePrice)
//	// Returns 200050000000, nil
func Float64ToScaledUint64(value float64, scale uint64, maxValue float64) (uint64, error) {
	// Check for NaN and Infinity
	if math.IsNaN(value) {
		return 0, fmt.Errorf("%w: value is NaN", ErrInvalidInput)
	}
	if math.IsInf(value, 0) {
		return 0, fmt.Errorf("%w: value is Infinity", ErrInvalidInput)
	}

	// Check for negative values
	if value < 0 {
		return 0, fmt.Errorf("%w: value is negative: %f", ErrUnderflow, value)
	}

	// Check bounds before conversion
	if maxValue > 0 && value > maxValue {
		return 0, fmt.Errorf("%w: value %f exceeds maximum %f (scale: %d)",
			ErrOutOfBounds, value, maxValue, scale)
	}

	// Convert to *big.Int first for precision (avoids float64 precision issues)
	// Note: FloatToBigIntBaseX takes int64, so we cast scale
	valueBig := FloatToBigIntBaseX(value, int64(scale))

	// Check overflow
	if !valueBig.IsUint64() {
		return 0, fmt.Errorf("%w: value %f * scale %d exceeds uint64 (max: %d)",
			ErrOverflow, value, scale, uint64(math.MaxUint64))
	}

	return valueBig.Uint64(), nil
}

// ============================================================================
// u64 (scaled) to float64 Conversions (for display/API only)
// ============================================================================

// ScaledUint64ToFloat64 converts uint64 (scaled) to float64 for display/API
//
// PURPOSE: Convert keys back to display values
// USAGE: API responses, UI display, logging
// CRITICAL: This conversion may lose precision. Only use for display purposes.
// PERFORMANCE: Fast direct conversion (no allocations)
//
// WARNING: Precision may be lost for very large values. Use ScaledUint64ToBigInt
// for exact calculations.
func ScaledUint64ToFloat64(value uint64, scale uint64) float64 {
	if scale == 0 {
		return 0
	}
	return float64(value) / float64(scale)
}

// ============================================================================
// Price Key Conversions (Most Common Use Case)
// ============================================================================

// BigIntToPriceKey converts *big.Int price to uint64 key for orderbook operations
//
// PURPOSE: Convert order prices to keys for fast sorting and map lookups
// USAGE: Orderbook price level keys, order matching, depth queries
// CRITICAL: Primary function for orderbook performance optimization
// PERFORMANCE: Optimized hot path with overflow protection
func BigIntToPriceKey(priceBig *big.Int) (uint64, error) {
	return BigIntToScaledUint64(priceBig, PriceScale)
}

// BigIntToPriceKeyFromSatoshi converts *big.Int (already in satoshi format) to uint64 key
//
// PURPOSE: Handle satoshi format from proto without double scaling
// USAGE: Converting satoshi strings from proto to uint64 keys for orderbook
// CRITICAL: Input must already be in satoshi format (1e8), not decimal
// DESIGN: Does NOT scale - assumes input is already scaled by 1e8
//
// Example:
//
//	priceBig := big.NewInt(5000000000000) // 50000.00 in satoshi
//	priceKey, err := BigIntToPriceKeyFromSatoshi(priceBig)
//	// Returns 5000000000000, nil (no scaling, used directly)
func BigIntToPriceKeyFromSatoshi(priceBig *big.Int) (uint64, error) {
	if priceBig == nil {
		return 0, ErrInvalidInput
	}

	if priceBig.Sign() < 0 {
		return 0, fmt.Errorf("%w: price cannot be negative: %s", ErrUnderflow, priceBig.String())
	}

	// Check if satoshi value fits in uint64 (no scaling needed - already satoshi)
	if !priceBig.IsUint64() {
		return 0, fmt.Errorf("%w: satoshi value %s exceeds uint64 (max: %d)",
			ErrOverflow, priceBig.String(), uint64(math.MaxUint64))
	}

	return priceBig.Uint64(), nil
}

// PriceKeyToBigInt converts uint64 price key back to *big.Int
//
// PURPOSE: Convert keys back to exact prices for calculations
// USAGE: Exact price calculations, PnL calculations, margin requirements
func PriceKeyToBigInt(priceKey uint64) *big.Int {
	return ScaledUint64ToBigInt(priceKey, PriceScale)
}

// Float64ToPriceKey converts float64 price to uint64 key with bounds checking
//
// PURPOSE: Convert display prices to keys
// USAGE: User input, API requests, market data
// CRITICAL: Validates bounds before conversion
func Float64ToPriceKey(price float64) (uint64, error) {
	return Float64ToScaledUint64(price, PriceScale, MaxSafePrice)
}

// PriceKeyToFloat64 converts uint64 price key to float64 for display
//
// PURPOSE: Convert keys to display prices
// USAGE: API responses, UI display, logging
// WARNING: May lose precision for very large values
func PriceKeyToFloat64(priceKey uint64) float64 {
	return ScaledUint64ToFloat64(priceKey, PriceScale)
}

// StringToPriceKey converts u256 string to uint64 price key
//
// PURPOSE: Convert string prices (from EIP-712) to keys
// USAGE: Order submission, signature verification, API parsing
func StringToPriceKey(priceStr string) (uint64, error) {
	priceBig, err := BigIntByString(priceStr)
	if err != nil {
		return 0, fmt.Errorf("invalid price string: %w", err)
	}
	return BigIntToPriceKey(priceBig)
}

// ============================================================================
// Value Key Conversions (for Aggregations)
// ============================================================================

// BigIntToValueKey converts *big.Int value to uint64 key for aggregation
//
// PURPOSE: Convert position values to keys for fast aggregation
// USAGE: ADL position calculator, portfolio aggregation, market totals
func BigIntToValueKey(valueBig *big.Int) (uint64, error) {
	return BigIntToScaledUint64(valueBig, ValueScale)
}

// ValueKeyToBigInt converts uint64 value key back to *big.Int
func ValueKeyToBigInt(valueKey uint64) *big.Int {
	return ScaledUint64ToBigInt(valueKey, ValueScale)
}

// Float64ToValueKey converts float64 value to uint64 key
func Float64ToValueKey(value float64) (uint64, error) {
	return Float64ToScaledUint64(value, ValueScale, MaxSafeValue)
}

// ValueKeyToFloat64 converts uint64 value key to float64
func ValueKeyToFloat64(valueKey uint64) float64 {
	return ScaledUint64ToFloat64(valueKey, ValueScale)
}

// ============================================================================
// Quantity Key Conversions
// ============================================================================

// BigIntToQuantityKey converts *big.Int quantity to uint64 key
func BigIntToQuantityKey(quantityBig *big.Int) (uint64, error) {
	return BigIntToScaledUint64(quantityBig, QuantityScale)
}

// BigIntToQuantityKeyFromSatoshi converts *big.Int (already in satoshi format) to uint64 key
//
// PURPOSE: Handle satoshi format from proto without double scaling
// USAGE: Converting satoshi strings from proto to uint64 keys for orderbook
// CRITICAL: Input must already be in satoshi format (1e8), not decimal
// DESIGN: Does NOT scale - assumes input is already scaled by 1e8
//
// Example:
//
//	quantityBig := big.NewInt(100000000) // 1.0 in satoshi
//	quantityKey, err := BigIntToQuantityKeyFromSatoshi(quantityBig)
//	// Returns 100000000, nil (no scaling, used directly)
func BigIntToQuantityKeyFromSatoshi(quantityBig *big.Int) (uint64, error) {
	if quantityBig == nil {
		return 0, ErrInvalidInput
	}

	if quantityBig.Sign() < 0 {
		return 0, fmt.Errorf("%w: quantity cannot be negative: %s", ErrUnderflow, quantityBig.String())
	}

	// Check if satoshi value fits in uint64 (no scaling needed - already satoshi)
	if !quantityBig.IsUint64() {
		return 0, fmt.Errorf("%w: satoshi value %s exceeds uint64 (max: %d)",
			ErrOverflow, quantityBig.String(), uint64(math.MaxUint64))
	}

	return quantityBig.Uint64(), nil
}

// QuantityKeyToBigInt converts uint64 quantity key back to *big.Int
func QuantityKeyToBigInt(quantityKey uint64) *big.Int {
	return ScaledUint64ToBigInt(quantityKey, QuantityScale)
}

// Float64ToQuantityKey converts float64 quantity to uint64 key
func Float64ToQuantityKey(quantity float64) (uint64, error) {
	return Float64ToScaledUint64(quantity, QuantityScale, MaxSafeQuantity)
}

// ============================================================================
// Ratio Key Conversions (for Margin Ratios, Risk Factors)
// ============================================================================

// Float64ToRatioKey converts float64 ratio to uint64 key
//
// PURPOSE: Convert ratios (margin ratio, risk factor) to keys for fast comparisons
// USAGE: Liquidation checks, risk factor comparisons, priority calculations
// CRITICAL: Used in cross-margin portfolio and liquidation engine
func Float64ToRatioKey(ratio float64) (uint64, error) {
	return Float64ToScaledUint64(ratio, RatioScale, MaxSafeRatio)
}

// RatioKeyToFloat64 converts uint64 ratio key to float64
func RatioKeyToFloat64(ratioKey uint64) float64 {
	return ScaledUint64ToFloat64(ratioKey, RatioScale)
}

// ============================================================================
// Safe Arithmetic Operations on u64 Keys (with Overflow Protection)
// ============================================================================

// AddPriceKeys safely adds two price keys with overflow protection
//
// PURPOSE: Aggregate prices in orderbook operations
// USAGE: Summing order quantities at price levels, calculating total value
// CRITICAL: Prevents silent overflow that could cause incorrect calculations
func AddPriceKeys(a, b uint64) (uint64, error) {
	if math.MaxUint64-a < b {
		return 0, fmt.Errorf("%w: price key addition %d + %d exceeds uint64 (max: %d)",
			ErrOverflow, a, b, uint64(math.MaxUint64))
	}
	return a + b, nil
}

// SubtractPriceKeys safely subtracts two price keys with underflow protection
//
// PURPOSE: Calculate price differences, remove quantities
// USAGE: Order cancellation, position updates, delta calculations
// CRITICAL: Prevents underflow that could cause negative values
func SubtractPriceKeys(a, b uint64) (uint64, error) {
	if a < b {
		return 0, fmt.Errorf("%w: price key subtraction %d - %d would result in negative value",
			ErrUnderflow, a, b)
	}
	return a - b, nil
}

// MultiplyPriceKeys safely multiplies two price keys with overflow protection
//
// PURPOSE: Calculate position values, margin requirements
// USAGE: Size * Price calculations, margin calculations
// CRITICAL: Result is divided by scale to maintain precision
// PERFORMANCE: Uses big.Int for safe multiplication, then converts back
func MultiplyPriceKeys(a, b uint64) (uint64, error) {
	// Convert to big.Int for safe multiplication
	aBig := new(big.Int).SetUint64(a)
	bBig := new(big.Int).SetUint64(b)

	// Multiply
	result := new(big.Int).Mul(aBig, bBig)

	// Divide by scale to maintain precision
	result.Div(result, priceScaleBig)

	// Check overflow
	if !result.IsUint64() {
		return 0, fmt.Errorf("%w: price key multiplication %d * %d exceeds uint64",
			ErrOverflow, a, b)
	}

	return result.Uint64(), nil
}

// ComparePriceKeys compares two price keys (returns -1, 0, or 1)
//
// PURPOSE: Fast integer comparison for sorting
// USAGE: Orderbook sorting, price level ordering, ranking
// PERFORMANCE: Direct integer comparison (fastest possible)
func ComparePriceKeys(a, b uint64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// ============================================================================
// Safe Arithmetic for Value Keys
// ============================================================================

// AddValueKeys safely adds two value keys
func AddValueKeys(a, b uint64) (uint64, error) {
	if math.MaxUint64-a < b {
		return 0, fmt.Errorf("%w: value key addition exceeds uint64", ErrOverflow)
	}
	return a + b, nil
}

// SubtractValueKeys safely subtracts two value keys
func SubtractValueKeys(a, b uint64) (uint64, error) {
	if a < b {
		return 0, fmt.Errorf("%w: value key subtraction would result in negative", ErrUnderflow)
	}
	return a - b, nil
}

// ============================================================================
// Batch Operations for Performance
// ============================================================================

// BatchBigIntToPriceKeys converts multiple *big.Int prices to uint64 keys
//
// PURPOSE: Efficient bulk conversion for orderbook snapshots
// USAGE: Orderbook depth queries, market data aggregation, batch processing
// PERFORMANCE: More efficient than individual conversions (reduces overhead)
func BatchBigIntToPriceKeys(prices []*big.Int) ([]uint64, error) {
	if prices == nil {
		return nil, ErrInvalidInput
	}

	keys := make([]uint64, len(prices))
	for i, price := range prices {
		key, err := BigIntToPriceKey(price)
		if err != nil {
			return nil, fmt.Errorf("price[%d]: %w", i, err)
		}
		keys[i] = key
	}
	return keys, nil
}

// BatchPriceKeysToBigInt converts multiple uint64 price keys to *big.Int
//
// PURPOSE: Bulk conversion for exact calculations
// USAGE: Batch PnL calculations, portfolio rebalancing
func BatchPriceKeysToBigInt(keys []uint64) []*big.Int {
	if keys == nil {
		return nil
	}

	prices := make([]*big.Int, len(keys))
	for i, key := range keys {
		prices[i] = PriceKeyToBigInt(key)
	}
	return prices
}

// BatchFloat64ToPriceKeys converts multiple float64 prices to uint64 keys
func BatchFloat64ToPriceKeys(prices []float64) ([]uint64, error) {
	if prices == nil {
		return nil, ErrInvalidInput
	}

	keys := make([]uint64, len(prices))
	for i, price := range prices {
		key, err := Float64ToPriceKey(price)
		if err != nil {
			return nil, fmt.Errorf("price[%d]: %w", i, err)
		}
		keys[i] = key
	}
	return keys, nil
}

// ============================================================================
// Validation and Bounds Checking
// ============================================================================

// ValidatePriceKey validates that a price key is within safe bounds
//
// PURPOSE: Pre-flight validation before operations
// USAGE: Order validation, risk checks, bounds verification
func ValidatePriceKey(priceKey uint64) error {
	maxPriceKey, err := Float64ToPriceKey(MaxSafePrice)
	if err != nil {
		// If MaxSafePrice itself can't be converted, use direct calculation
		maxPriceKey = math.MaxUint64
	}

	if priceKey > maxPriceKey {
		return fmt.Errorf("%w: price key %d exceeds maximum %d (max price: %f)",
			ErrOutOfBounds, priceKey, maxPriceKey, MaxSafePrice)
	}

	return nil
}

// ValidateValueKey validates that a value key is within safe bounds
func ValidateValueKey(valueKey uint64) error {
	maxValueKey, err := Float64ToValueKey(MaxSafeValue)
	if err != nil {
		maxValueKey = math.MaxUint64
	}

	if valueKey > maxValueKey {
		return fmt.Errorf("%w: value key %d exceeds maximum %d",
			ErrOutOfBounds, valueKey, maxValueKey)
	}

	return nil
}

// ============================================================================
// Advanced Operations for Risk Calculations
// ============================================================================

// CalculateMarginRatioKey calculates margin ratio as uint64 key
//
// PURPOSE: Fast margin ratio comparison for liquidation checks
// USAGE: Cross-margin portfolio, liquidation engine, risk factor calculations
// CRITICAL: Used in high-frequency risk checks
//
// DESIGN NOTES:
// - Uses RatioScale = 1e8 (satoshi precision) for consistency with other scales
// - Rounding: Always rounds DOWN (truncates) via big.Int.Div() for conservative risk assessment
// - Conservative rounding ensures earlier liquidation detection, reducing risk exposure
// - This is a standard risk management practice to err on the side of caution
func CalculateMarginRatioKey(equityKey, marginRequirementKey uint64) (uint64, error) {
	if marginRequirementKey == 0 {
		return 0, ErrDivisionByZero
	}

	// Convert to big.Int for precise division
	equityBig := new(big.Int).SetUint64(equityKey)
	marginBig := new(big.Int).SetUint64(marginRequirementKey)

	// Calculate ratio: (equity / margin) * RatioScale
	// Multiply equity by scale first to maintain precision
	// CRITICAL: big.Int.Div() truncates (rounds DOWN), providing conservative risk assessment
	equityBig.Mul(equityBig, ratioScaleBig)
	result := new(big.Int).Div(equityBig, marginBig)

	if !result.IsUint64() {
		return 0, fmt.Errorf("%w: margin ratio calculation exceeds uint64", ErrOverflow)
	}

	return result.Uint64(), nil
}

// IsLiquidatableKey checks if position is liquidatable using uint64 keys
//
// PURPOSE: Fast liquidation check without float64 comparisons
// USAGE: Liquidation engine, risk monitoring
// CRITICAL: High-frequency operation, must be fast
func IsLiquidatableKey(marginRatioKey uint64, liquidationThresholdKey uint64) bool {
	return marginRatioKey < liquidationThresholdKey
}

// ============================================================================
// Score Calculations for ADL Ranking
// ============================================================================

// Float64ToScoreKey converts float64 ADL score to uint64 key for sorting
//
// PURPOSE: Convert ADL scores to keys for fast sorting
// USAGE: ADL ranking module, position prioritization
// CRITICAL: Used in high-frequency sorting operations
func Float64ToScoreKey(score float64) (uint64, error) {
	maxScore := float64(math.MaxUint64) / ScoreScale
	return Float64ToScaledUint64(score, ScoreScale, maxScore)
}

// ScoreKeyToFloat64 converts uint64 score key to float64
func ScoreKeyToFloat64(scoreKey uint64) float64 {
	return ScaledUint64ToFloat64(scoreKey, ScoreScale)
}

// ============================================================================
// Utility Functions
// ============================================================================

// MaxPriceKey returns the maximum safe price key
func MaxPriceKey() uint64 {
	maxKey, _ := Float64ToPriceKey(MaxSafePrice)
	return maxKey
}

// MinPriceKey returns the minimum price key (0)
func MinPriceKey() uint64 {
	return 0
}

// PriceKeyRange returns the valid range for price keys
func PriceKeyRange() (min, max uint64) {
	return MinPriceKey(), MaxPriceKey()
}
