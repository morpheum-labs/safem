package safem

import (
	"math/big"
	"testing"
)

// ============================================================================
// Benchmarks for MulDivPrecise
// ============================================================================

func BenchmarkMulDivPrecise_RoundDown(b *testing.B) {
	a := big.NewInt(100000000000)    // 1000.0 scaled by 1e8
	bVal := big.NewInt(200000000000) // 2000.0 scaled by 1e8
	c := big.NewInt(100000000)       // 1e8 scale

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MulDivPrecise(a, bVal, c, false)
	}
}

func BenchmarkMulDivPrecise_RoundUp(b *testing.B) {
	a := big.NewInt(100000000000)
	bVal := big.NewInt(200000000000)
	c := big.NewInt(100000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MulDivPrecise(a, bVal, c, true)
	}
}

func BenchmarkBigIntDiv_Standard(b *testing.B) {
	a := big.NewInt(100000000000)
	bVal := big.NewInt(200000000000)
	c := big.NewInt(100000000)

	product := new(big.Int).Mul(a, bVal)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = new(big.Int).Div(product, c)
	}
}

// ============================================================================
// Benchmarks for CalculateFee
// ============================================================================

func BenchmarkCalculateFee(b *testing.B) {
	amountIn := big.NewInt(100000000000) // 1000.0 scaled by 1e8
	feePips := big.NewInt(10)            // 0.1% = 10 pips

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateFee(amountIn, feePips)
	}
}

// ============================================================================
// Benchmarks for CalculatePnLAdjusted
// ============================================================================

func BenchmarkCalculatePnLAdjusted_Profit(b *testing.B) {
	priceDiff := big.NewInt(100000000) // +1.0 price diff
	size := big.NewInt(10000000000)    // 100.0 size
	scale := big.NewInt(100000000)     // 1e8 scale

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculatePnLAdjusted(priceDiff, size, scale)
	}
}

func BenchmarkCalculatePnLAdjusted_Loss(b *testing.B) {
	priceDiff := big.NewInt(-100000000) // -1.0 price diff
	size := big.NewInt(10000000000)     // 100.0 size
	scale := big.NewInt(100000000)      // 1e8 scale

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculatePnLAdjusted(priceDiff, size, scale)
	}
}

// ============================================================================
// Benchmarks for CalculateMarginRatioConservative
// ============================================================================

func BenchmarkCalculateMarginRatioConservative(b *testing.B) {
	equity := big.NewInt(100000000000)           // 1000.0 equity
	marginRequirement := big.NewInt(50000000000) // 500.0 margin

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateMarginRatioConservative(equity, marginRequirement)
	}
}

// ============================================================================
// Benchmarks for CalculateLiquidationPriceConservative
// ============================================================================

func BenchmarkCalculateLiquidationPriceConservative(b *testing.B) {
	entryPrice := big.NewInt(200000000000)       // 2000.0 entry price
	pnlAtLiquidation := big.NewInt(-50000000000) // -500.0 PnL
	size := big.NewInt(10000000000)              // 100.0 size

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateLiquidationPriceConservative(entryPrice, pnlAtLiquidation, size)
	}
}

// ============================================================================
// Comparison Benchmarks: New vs Old Approach
// ============================================================================

func BenchmarkFeeCalculation_NewApproach(b *testing.B) {
	amountIn := big.NewInt(100000000000)
	feePips := big.NewInt(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculateFee(amountIn, feePips)
	}
}

func BenchmarkFeeCalculation_OldApproach(b *testing.B) {
	amountIn := big.NewInt(100000000000)
	feePips := big.NewInt(10)
	feeDenominator := big.NewInt(10000)

	product := new(big.Int).Mul(amountIn, feePips)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = new(big.Int).Div(product, feeDenominator)
	}
}

func BenchmarkPnLCalculation_NewApproach(b *testing.B) {
	priceDiff := big.NewInt(100000000)
	size := big.NewInt(10000000000)
	scale := big.NewInt(100000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculatePnLAdjusted(priceDiff, size, scale)
	}
}

func BenchmarkPnLCalculation_OldApproach(b *testing.B) {
	priceDiff := big.NewInt(100000000)
	size := big.NewInt(10000000000)
	scale := big.NewInt(100000000)

	product := new(big.Int).Mul(priceDiff, size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = new(big.Int).Div(product, scale)
	}
}
