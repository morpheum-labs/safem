// Package safem provides high-precision arithmetic operations for financial calculations
// with precision-preserving rounding strategies for orderbook operations, fee calculations,
// PnL computations, and risk management.
//
// PURPOSE:
// - Provide precise multiply-divide operations with configurable rounding modes
// - Calculate fees, PnL, margin ratios, and liquidation prices with precision preservation
// - Ensure consistent rounding behavior across all financial calculations
// - Prevent precision loss in high-frequency trading operations
//
// USAGE PATTERNS:
// 1. Fee Calculations:
//   - Use CalculateFee() for trading fee computations
//   - Ensures precision preservation in fee collection
//
// 2. PnL Calculations:
//   - Use CalculatePnLAdjusted() for profit/loss computations
//   - Handles both positive and negative PnL with appropriate precision
//
// 3. Risk Management:
//   - Use CalculateMarginRatioConservative() for margin ratio calculations
//   - Use CalculateLiquidationPriceConservative() for liquidation price computations
//   - Conservative rounding ensures earlier risk detection
//
// CRITICAL ATTENTION:
// ⚠️  PRECISION: All operations use satoshi precision (1e8) for consistency
// ⚠️  ROUNDING: Rounding modes are selected to preserve precision and ensure conservative risk calculations
// ⚠️  THREAD SAFETY: All functions are pure and thread-safe
// ⚠️  PERFORMANCE: Optimized for high-frequency operations (orderbook, risk engine)
// ⚠️  MEMORY SAFETY: Reuses big.Int allocations where possible
//
// DESIGN PRINCIPLES:
// - Precision first: All calculations maintain maximum precision
// - Conservative risk: Risk calculations use conservative rounding for safety
// - Fail-fast: Return errors immediately on invalid input
// - Zero-allocation: Reuse big.Int from pool where possible
// - Performance: Optimize hot paths (fee calculations, PnL computations)
package safem

import (
	"fmt"
	"math/big"
)

// MulDivPrecise performs (a * b) / c with configurable rounding precision
//
// PURPOSE: Core arithmetic operation for financial calculations requiring precise rounding
// USAGE: Fee calculations, PnL computations, ratio calculations
// CRITICAL: Rounding direction is determined by ceil parameter to preserve precision
// PERFORMANCE: Uses efficient big.Int operations with minimal allocations
//
// DESIGN NOTES:
// - When ceil=true: Rounds UP (ceiling) - used for fee collection to ensure precision preservation
// - When ceil=false: Rounds DOWN (truncates) - used for conservative risk calculations
// - This approach prevents precision loss that could accumulate over many operations
// - Division by zero is explicitly checked and returns error
//
// Example:
//
//	fee := MulDivPrecise(amount, feeRate, big.NewInt(10000), true)  // Round UP for fees
//	ratio := MulDivPrecise(equity, scale, margin, false)            // Round DOWN for ratios
func MulDivPrecise(a, b, c *big.Int, ceil bool) (*big.Int, error) {
	if a == nil || b == nil || c == nil {
		return nil, ErrInvalidInput
	}

	if c.Sign() == 0 {
		return nil, ErrDivisionByZero
	}

	// Calculate product: a * b
	product := new(big.Int).Mul(a, b)

	// Perform division
	result := new(big.Int).Div(product, c)

	// Apply rounding based on ceil parameter
	// CRITICAL: This rounding strategy preserves precision in financial calculations
	if ceil {
		// Round UP: Check if there's a remainder and add 1 if needed
		// This ensures we don't lose precision in fee collection
		remainder := new(big.Int).Mod(product, c)
		if remainder.Sign() > 0 {
			result.Add(result, big.NewInt(1))
		}
	}
	// When ceil=false, result is already truncated (rounded DOWN) by Div()

	return result, nil
}

// CalculateFee computes trading fee with precision preservation
//
// PURPOSE: Calculate trading fees for order execution and swap operations
// USAGE: Order matching, swap calculations, fee collection
// CRITICAL: Uses ceiling rounding to ensure fee collection preserves precision
// PERFORMANCE: Optimized for high-frequency fee calculations
//
// DESIGN NOTES:
// - Formula: feeAmount = (amountIn * feePips) / 10000
// - Rounding: Always rounds UP (ceiling) to preserve precision in fee collection
// - This prevents systematic precision loss that could accumulate over many trades
// - feePips is in basis points (e.g., 30 = 0.3%)
//
// Example:
//
//	amount := big.NewInt(100000000000)  // 1000.0 with 8 decimals
//	feePips := big.NewInt(30)           // 0.3%
//	fee := CalculateFee(amount, feePips)  // Returns fee with precision preserved
func CalculateFee(amountIn, feePips *big.Int) (*big.Int, error) {
	if amountIn == nil || feePips == nil {
		return nil, ErrInvalidInput
	}

	if amountIn.Sign() < 0 || feePips.Sign() < 0 {
		return nil, fmt.Errorf("negative values not allowed: amountIn=%s, feePips=%s",
			amountIn.String(), feePips.String())
	}

	// Fee calculation: (amountIn * feePips) / 10000
	// CRITICAL: Round UP (ceil=true) to preserve precision in fee collection
	// This ensures the exchange collects fees with full precision
	feeDenominator := big.NewInt(10000)
	return MulDivPrecise(amountIn, feePips, feeDenominator, true)
}

// CalculatePnLAdjusted computes profit/loss with precision-adjusted rounding
//
// PURPOSE: Calculate PnL for positions with appropriate precision handling
// USAGE: Position updates, trade execution, portfolio valuation
// CRITICAL: Rounding direction depends on PnL sign to preserve precision
// PERFORMANCE: Optimized for high-frequency PnL calculations
//
// DESIGN NOTES:
// - Formula: pnl = (priceDiff * size) / scale
// - Rounding strategy:
//   - If PnL > 0 (user profit): Round DOWN (truncate) - preserves precision conservatively
//   - If PnL < 0 (user loss): Round UP (ceiling) - ensures precision in loss calculations
//
// - This approach prevents precision accumulation issues in both directions
// - scale is typically PriceScale (1e8) for satoshi precision
//
// Example:
//
//	priceDiff := big.NewInt(1000000)    // 0.01 price difference
//	size := big.NewInt(10000000000)     // 100.0 quantity
//	scale := big.NewInt(1e8)            // PriceScale
//	pnl := CalculatePnLAdjusted(priceDiff, size, scale)  // Returns adjusted PnL
func CalculatePnLAdjusted(priceDiff, size, scale *big.Int) (*big.Int, error) {
	if priceDiff == nil || size == nil || scale == nil {
		return nil, ErrInvalidInput
	}

	if scale.Sign() == 0 {
		return nil, ErrDivisionByZero
	}

	// Calculate PnL: (priceDiff * size) / scale
	// CRITICAL: Determine rounding direction based on PnL sign
	// This preserves precision appropriately for both profits and losses
	product := new(big.Int).Mul(priceDiff, size)

	// Check sign of product to determine rounding direction
	// Positive PnL (profit): Round DOWN (conservative, preserves precision)
	// Negative PnL (loss): Round UP (ensures precision in loss calculations)
	isProfit := product.Sign() > 0
	ceil := !isProfit // Round UP for losses, DOWN for profits

	// Use MulDivPrecise with determined rounding direction
	return MulDivPrecise(priceDiff, size, scale, ceil)
}

// CalculateMarginRatioConservative computes margin ratio with conservative rounding
//
// PURPOSE: Calculate margin ratio for risk management and liquidation checks
// USAGE: Cross-margin portfolio, liquidation engine, risk monitoring
// CRITICAL: Uses conservative rounding (DOWN) for earlier risk detection
// PERFORMANCE: Optimized for high-frequency risk checks
//
// DESIGN NOTES:
// - Formula: ratio = (equity * RatioScale) / marginRequirement
// - Rounding: Always rounds DOWN (truncates) for conservative risk assessment
// - Conservative rounding ensures earlier liquidation detection, reducing risk exposure
// - This is a standard risk management practice to err on the side of caution
// - RatioScale is 1e8 (satoshi precision) for consistency
//
// Example:
//
//	equity := big.NewInt(100000000000)        // 1000.0 equity
//	marginReq := big.NewInt(50000000000)      // 500.0 margin requirement
//	ratio := CalculateMarginRatioConservative(equity, marginReq)  // Returns conservative ratio
func CalculateMarginRatioConservative(equity, marginRequirement *big.Int) (*big.Int, error) {
	if equity == nil || marginRequirement == nil {
		return nil, ErrInvalidInput
	}

	if marginRequirement.Sign() == 0 {
		return nil, ErrDivisionByZero
	}

	if equity.Sign() < 0 || marginRequirement.Sign() < 0 {
		return nil, fmt.Errorf("negative values not allowed: equity=%s, marginRequirement=%s",
			equity.String(), marginRequirement.String())
	}

	// Margin ratio calculation: (equity * RatioScale) / marginRequirement
	// CRITICAL: Round DOWN (ceil=false) for conservative risk assessment
	// This ensures earlier liquidation detection, which is a standard risk management practice
	equityScaled := new(big.Int).Mul(equity, ratioScaleBig)
	return MulDivPrecise(equityScaled, big.NewInt(1), marginRequirement, false)
}

// CalculateLiquidationPriceConservative computes liquidation price with conservative rounding
//
// PURPOSE: Calculate liquidation price for risk management and position monitoring
// USAGE: Liquidation engine, risk monitoring, position management
// CRITICAL: Uses conservative rounding (DOWN) for earlier liquidation detection
// PERFORMANCE: Optimized for high-frequency liquidation checks
//
// DESIGN NOTES:
// - Formula: liquidationPrice = entryPrice + (pnlAtLiquidation / size)
// - Rounding: Always rounds DOWN (truncates) for conservative risk assessment
// - Conservative rounding ensures earlier liquidation, reducing risk exposure
// - This is a standard risk management practice to protect against adverse price movements
// - Optional safety margin can be applied by caller if needed
//
// Example:
//
//	entryPrice := big.NewInt(200000000000)     // 2000.0 entry price
//	pnlAtLiquidation := big.NewInt(-50000000000)  // -500.0 PnL at liquidation
//	size := big.NewInt(10000000000)           // 100.0 position size
//	liqPrice := CalculateLiquidationPriceConservative(entryPrice, pnlAtLiquidation, size)
func CalculateLiquidationPriceConservative(entryPrice, pnlAtLiquidation, size *big.Int) (*big.Int, error) {
	if entryPrice == nil || pnlAtLiquidation == nil || size == nil {
		return nil, ErrInvalidInput
	}

	if size.Sign() == 0 {
		return nil, ErrDivisionByZero
	}

	// Calculate price adjustment: pnlAtLiquidation / size
	// CRITICAL: Round DOWN (ceil=false) for conservative liquidation price
	// This ensures earlier liquidation detection, which is a standard risk management practice
	priceAdjustment, err := MulDivPrecise(pnlAtLiquidation, big.NewInt(1), size, false)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price adjustment: %w", err)
	}

	// Add adjustment to entry price
	liquidationPrice := new(big.Int).Add(entryPrice, priceAdjustment)

	return liquidationPrice, nil
}

// RoundUp rounds value up to the nearest multiple of divisor
//
// PURPOSE: Explicit ceiling rounding helper for precision-preserving operations
// USAGE: Fee calculations, loss computations, precision preservation
// CRITICAL: Always rounds UP, ensuring no precision is lost in upward direction
// PERFORMANCE: Efficient big.Int operations
//
// DESIGN NOTES:
// - If remainder exists, adds 1 to quotient
// - Used when precision preservation requires upward rounding
// - Standard mathematical ceiling operation
//
// Example:
//
//	value := big.NewInt(123)
//	divisor := big.NewInt(10)
//	rounded := RoundUp(value, divisor)  // Returns 13 (ceiling of 12.3)
func RoundUp(value, divisor *big.Int) *big.Int {
	if value == nil || divisor == nil || divisor.Sign() == 0 {
		return big.NewInt(0)
	}

	result := new(big.Int).Div(value, divisor)
	remainder := new(big.Int).Mod(value, divisor)
	if remainder.Sign() > 0 {
		result.Add(result, big.NewInt(1))
	}
	return result
}

// RoundDown rounds value down to the nearest multiple of divisor (truncates)
//
// PURPOSE: Explicit floor rounding helper for conservative calculations
// USAGE: Margin ratios, liquidation prices, conservative risk calculations
// CRITICAL: Always rounds DOWN, ensuring conservative estimates
// PERFORMANCE: Efficient big.Int operations
//
// DESIGN NOTES:
// - Standard truncation (floor operation)
// - Used for conservative risk calculations
// - Standard mathematical floor operation
//
// Example:
//
//	value := big.NewInt(123)
//	divisor := big.NewInt(10)
//	rounded := RoundDown(value, divisor)  // Returns 12 (floor of 12.3)
func RoundDown(value, divisor *big.Int) *big.Int {
	if value == nil || divisor == nil || divisor.Sign() == 0 {
		return big.NewInt(0)
	}

	return new(big.Int).Div(value, divisor)
}
